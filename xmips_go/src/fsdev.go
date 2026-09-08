package main

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 文件/套接字类型
const (
	fdFile = iota // 文件
	fdSock       // TCP 客户端
)

// fdNoData socket 非阻塞读无可读数据（仅多进程模式产生），
// dispatcher 据此识别当前进程应让出时间片（放就绪队尾）而非立即返回 0 自旋。
const fdNoData = -6

// fdEntry fd 表条目
type fdEntry struct {
	kind   int       // fdFile 或 fdSock
	path   string    // 路径 或 sock:host:port
	host   *os.File  // 文件
	offset int64     // 文件读写偏移
	conn   net.Conn  // socket 连接
	used   bool      // 槽位是否占用
}

// FsDev 文件系统/套接字子系统（模拟器单线程，无需加锁）
type FsDev struct {
	table []fdEntry // fd 表
}

// newFsDev 创建 fsdev，fd 从 3 起（0/1/2 预留 stdio 语义）
func newFsDev() *FsDev {
	return &FsDev{table: make([]fdEntry, 128)}
}

// errCheck 统一按 RES 输出（仅在有 Error 级状态时展示）
func errCheck(res int) {
	if systemChecker.showLevel(res, sysLog[:]) {
		systemChecker.check(res, sysLog[:])
	}
}

// cleanPath 将相对文件路径清洗并限定在 diskRoot 内；越界/非法返回空串
func cleanPath(p string) string {
	clean := filepath.Clean(strings.ReplaceAll(p, "\\", "/"))
	if filepath.IsAbs(clean) {
		return ""
	}
	// 目录穿越到 diskRoot 之外（含根目录）视为非法
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	return clean
}

// do syscall 调用：kind 0=open,1=read,2=write,3=close。
// 参数均从调用者寄存器/数据段读取，结果返回给调用者（#13 槽写回由 dispatcher 完成）
func (f *FsDev) fsOpen(pptr *Process, pathPtr, mode int) int {
	// 从数据段读取 0 结尾路径串
	var sb strings.Builder
	i := 0
	for {
		c := pptr.MData.read(pathPtr + i)
		if c == 0 || c < 0 {
			break
		}
		sb.WriteByte(byte(c & 0xff))
		i++
		if i > 1024 {
			break
		}
	}
	path := sb.String()

	// 寻找空闲 fd 槽
	fd := f.allocFd()
	if fd < 0 {
		return -2
	}

	if strings.HasPrefix(path, "sock:") || strings.HasPrefix(path, "tlssock:") {
		// TCP 客户端连接（tlssock: 强制 TLS；sock: 若端口 443 自动启用 TLS）
		isTLS := strings.HasPrefix(path, "tlssock:")
		addr := path
		addr = strings.TrimPrefix(addr, "tlssock:")
		addr = strings.TrimPrefix(addr, "sock:")
		addr = strings.TrimPrefix(addr, "//")
		if !isTLS {
			if _, port, err := net.SplitHostPort(addr); err == nil && port == "443" {
				isTLS = true
			}
		}
		timeout := time.Duration(sockTimeout) * time.Millisecond
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			errCheck(204)
			f.table[fd].used = false
			return -3
		}
		if isTLS {
			host, _, _ := net.SplitHostPort(addr)
			host = strings.Trim(host, "[]") // 去掉 IPv6 括号
			tconn := tls.Client(conn, &tls.Config{ServerName: host})
			// 握手同样受 sockTimeout 约束，避免卡死调度
			_ = tconn.SetDeadline(time.Now().Add(timeout))
			if err := tconn.Handshake(); err != nil {
				_ = conn.Close()
				f.table[fd].used = false
				return -3
			}
			_ = tconn.SetDeadline(time.Time{})
			conn = tconn
		}
		f.table[fd].kind = fdSock
		f.table[fd].path = path
		f.table[fd].conn = conn
		return fd
	}

	// 文件：限定在 diskRoot 内
	p := cleanPath(path)
	if p == "" {
		errCheck(200)
		f.table[fd].used = false
		return -1
	}
	// 确保 diskRoot 目录存在
	if err := os.MkdirAll(diskRoot, 0o755); err != nil {
		errCheck(200)
		f.table[fd].used = false
		return -1
	}
	full := filepath.Join(diskRoot, p)

	var flag int
	switch mode {
	case 0: // 只读
		flag = os.O_RDONLY
	case 1: // 写（截断）
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	case 2: // 读写追加
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	default:
		flag = os.O_RDWR
	}
	host, err := os.OpenFile(full, flag, 0o644)
	if err != nil {
		errCheck(200)
		f.table[fd].used = false
		return -1
	}
	f.table[fd].kind = fdFile
	f.table[fd].path = path
	f.table[fd].host = host
	// 每次 open 重置偏移，避免复用旧 fd 槽残留的 offset
	f.table[fd].offset = 0
	if mode == 2 { // 追加：偏移定位到文件尾
		if off, err := host.Seek(0, io.SeekEnd); err == nil {
			f.table[fd].offset = off
		}
	}
	return fd
}

// fsRead 从 fd 读 len 字节到数据段 buf 处，返回实际读到的字节数
func (f *FsDev) fsRead(pptr *Process, buf, fd, length int) int {
	if fd < 0 || fd >= len(f.table) || !f.table[fd].used {
		errCheck(201)
		return -2
	}
	e := &f.table[fd]

	if e.kind == fdFile {
		data := make([]byte, length)
		n, err := e.host.ReadAt(data, e.offset)
		e.offset += int64(n)
		if err != nil && err != io.EOF {
			return -5
		}
		for i := 0; i < n; i++ {
			pptr.MData.write(buf+i, int(data[i]))
		}
		return n
	}

	// socket
	if e.conn == nil {
		return -5
	}
	if singleProc {
		// 单进程阻塞：设 sockTimeout 超时，防止服务器 keep-alive 不关连接导致永久阻塞
		_ = e.conn.SetReadDeadline(time.Now().Add(time.Duration(sockTimeout) * time.Millisecond))
	} else {
		// 多进程非阻塞：deadline=now，有数据立即返回，无数据立即 fdNoData
		_ = e.conn.SetReadDeadline(time.Now())
	}
	data := make([]byte, length)
	n, err := e.conn.Read(data)
	_ = e.conn.SetReadDeadline(time.Time{}) // 复位
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		// 非阻塞下无数据可读：返回哨兵，令 dispatcher 让出时间片（多进程模式）
		return fdNoData
	}
	if err != nil {
		if err == io.EOF {
			return -5 // 连接关闭
		}
		return -5
	}
	for i := 0; i < n; i++ {
		pptr.MData.write(buf+i, int(data[i]))
	}
	return n
}

// fsWrite 从数据段 buf 处写 len 字节到 fd，返回实际写入字节数
func (f *FsDev) fsWrite(pptr *Process, buf, fd, length int) int {
	if fd < 0 || fd >= len(f.table) || !f.table[fd].used {
		errCheck(202)
		return -2
	}
	e := &f.table[fd]
	if length < 0 {
		return -4
	}
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(pptr.MData.read(buf+i) & 0xff)
	}

	if e.kind == fdFile {
		n, err := e.host.WriteAt(data, e.offset)
		if err != nil {
			return -5
		}
		e.offset += int64(n)
		return n
	}

	// socket
	if e.conn == nil {
		return -5
	}
	// 写也受 sockTimeout 约束（单/多进程均设，防止发送缓冲长期拥堵）
	_ = e.conn.SetWriteDeadline(time.Now().Add(time.Duration(sockTimeout) * time.Millisecond))
	n, err := e.conn.Write(data)
	_ = e.conn.SetWriteDeadline(time.Time{})
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return -4 // 发送缓冲满
	}
	if err != nil {
		return -5
	}
	return n
}

// fsClose 关闭 fd，返回 0 成功；<0 错误
func (f *FsDev) fsClose(fd int) int {
	if fd < 0 || fd >= len(f.table) || !f.table[fd].used {
		errCheck(203)
		return -2
	}
	e := &f.table[fd]
	if e.kind == fdFile && e.host != nil {
		_ = e.host.Close()
	}
	if e.kind == fdSock && e.conn != nil {
		_ = e.conn.Close()
	}
	e.host = nil
	e.conn = nil
	e.path = ""
	e.used = false
	return 0
}

// allocFd 分配一个空闲 fd
func (f *FsDev) allocFd() int {
	for i := 3; i < len(f.table); i++ {
		if !f.table[i].used {
			f.table[i].used = true
			return i
		}
	}
	return -1 // 表满
}