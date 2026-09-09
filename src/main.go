package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// config 读取系统配置文件 config.ini（位于系统根目录 runDir，缺省用内置默认值）
func config() int {
	f, err := os.Open(filepath.Join(runDir, "config.ini"))
	if err != nil {
		// 无 config.ini 时直接使用内置默认值（已与模板 config.ini 对齐），不视为错误
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		para := scanner.Text()
		if para == "end" || para == "END" {
			break
		}

		// 解析 "key=value" 格式，无空格
		idx := strings.Index(para, "=")
		if idx < 0 {
			continue
		}
		key := para[:idx]
		val := para[idx+1:]
		r := atoiSafe(val)

		switch key {
		case "delayMode":
			delayMode = r
		case "displayMode":
			displayMode = r
		case "reportLevel":
			reportLevel = r
		case "codeSize":
			if r > 0 {
				codeSize = r
			}
		case "dataSize":
			if r > 0 {
				dataSize = r
			}
		case "stackSize":
			if r > 0 {
				stackSize = r
			}
		case "sysfunCodeSize":
			if r > 0 {
				sysfunCodeSize = r
			}
		case "sysfunDataSize":
			if r > 0 {
				sysfunDataSize = r
			}
		case "sysfunStackSize":
			if r > 0 {
				sysfunStackSize = r
			}
		case "cycleTimes":
			if r > 0 {
				cycleTimes = r
			}
		case "pcbNum":
			if r > 0 {
				pcbNum = r
			}
		case "bitMode":
			if r == 32 || r == 64 {
				wordBits = r
			}
		case "diskRoot":
			// 字符串值需剥离行内注释与两端空白
			v := strings.SplitN(val, ";", 2)[0]
			v = strings.TrimSpace(v)
			if v != "" {
				diskRoot = v
			}
		case "sockTimeout":
			if r > 0 {
				sockTimeout = r
			}
		case "dump":
			dumpEnabled = r
		}
	}
	return 0
}

func atoiSafe(s string) int {
	n := 0
	negative := false
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	if negative {
		n = -n
	}
	return n
}

// osOpen 直接按完整路径打开文件（用于命令行指定的任意路径程序）
func osOpen(path string) (*os.File, int) {
	f, err := os.Open(path)
	if err != nil {
		return nil, -1
	}
	return f, 0
}

// removeCo 删除汇编生成的用户程序中间 .co 文件（生成在 cwd，运行后清理）
func removeCo(name string) {
	coName := name
	if d := strings.LastIndex(coName, "."); d >= 0 {
		coName = coName[:d]
	}
	os.Remove(coName + ".co")
}

// rebuildSysfun 重新汇编所有 .scp 系统函数，生成对应的 .co 文件；返回成功条数
func rebuildSysfun(e *Editor, a *Assembler, sys *Storage) int {
	cnt := 0
	for i := 0; i < sysFunNumber; i++ {
		if sysFunTable[i] == "" {
			continue
		}
		fp, ret := sys.getFile(sysFunTable[i], 0)
		if ret == 0 {
			if e.ASM(fp, a, sysFunTable[i], sysPath[0]) != -1 { // 写到 sysfun/ 目录
				cnt++
			}
			sys.releaseFile(fp)
		}
	}
	return cnt
}

func main() {
	// 系统/数据根目录：默认 ~/.xmips（可被环境变量 XMIPS_HOME 覆盖）。
	// xmips 安装到 /usr/local/bin 后，config.ini / userfile / sysfun / run.list / disk
	// 均定位到该目录，程序可从任意目录直接用 `xmips XXX.cupa` 运行
	runDir = os.Getenv("XMIPS_HOME")
	if runDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			runDir = filepath.Join(h, ".xmips")
		}
	}
	if runDir == "" {
		runDir, _ = os.Getwd()
	}

	// 系统目录基于 runDir（原相对 "./sysfun/"、"./userfile/" 随 cwd 变化，改为固定到系统根目录）
	sysPath[0] = filepath.Join(runDir, "sysfun") + string(filepath.Separator)
	sysPath[1] = filepath.Join(runDir, "userfile") + string(filepath.Separator)

	config() // 读取 config.ini 覆盖默认值；无该文件则沿用内置默认
	initDump() // 开启 dump 日志（dump=1 时创建 dump.log）

	// 独立命令：./xmips update → 重建系统函数库（重新汇编 sysfun/*.scp 生成 .co）
	if len(os.Args) >= 2 && os.Args[1] == "update" {
		ue := newEditor(103, sysfunCodeSize, sysfunDataSize)
		ua := newAssembler(102)
		usys := newStorage(106, sysPath[0])
		cnt := rebuildSysfun(ue, ua, usys)
		fmt.Printf("system functions rebuilt: %d\n", cnt)
		closeDump()
		return
	}

	// 构建硬件/系统对象
	im := newInterpreter(101, cycleTimes, 10, 20)
	e := newEditor(103, codeSize, dataSize)
	a := newAssembler(102)
	os_ := newDispatcher(105, 10, 30, pcbNum)
	os_.FS = newFsDev() // 注入文件系统/套接字子系统
	sys := newStorage(106, sysPath[0])
	disk := newStorage(107, sysPath[1])

	//************************************************************************************************************
	// 初始化系统函数
	{
		pptr := make([]*Process, sysFunNumber)
		// 进程模板：代码段大、数据段小、高优先级
		tp := newProcess(0, sysfunCodeSize, sysfunDataSize, sysfunStackSize, 0)

		for i := 0; i < sysFunNumber; i++ {
			if sysFunTable[i] == "" {
				continue
			}
			pptr[i] = newProcess(0, sysfunCodeSize, sysfunDataSize, sysfunStackSize, 0)
			pptr[i].copy(tp)
			pptr[i].ID = i

			if updateSysfun == 1 { // 重新汇编系统函数
					fp, ret := sys.getFile(sysFunTable[i], 0)
					if ret == 0 {
						if e.ASM(fp, a, sysFunTable[i], sysPath[0]) != -1 {
							Load(pptr[i], sysFunTable[i], sysPath[0])
							os_.SysCall.put(i, pptr[i])
						}
						sys.releaseFile(fp)
					}
				} else {
					Load(pptr[i], sysFunTable[i], sysPath[0]) // 从 sysfun/ 加载系统函数 .co（加速）
					os_.SysCall.put(i, pptr[i])
				}
		}
	}
	//************************************************************************************************************

	//************************************************************************************************************
	// 加载用户程序
	// 支持两种方式：
	//   1) ./xmips            -> 从 run.list 读取要运行的程序
	//   2) ./xmips prog.cupa  -> 直接用命令行参数指定程序，忽略 run.list
	{
		// 从 userfile 目录加载用户程序（run.list 场景）
		loadFromDisk := func(name string, id int) int {
			pfr, ret2 := disk.getFile(name, 0)
			if ret2 != 0 {
				fmt.Printf("cannot open %s\n", name)
				return ret2
			}
			defer disk.releaseFile(pfr)
			pptr := newProcess(0, codeSize, dataSize, stackSize, 1)
			pptr.ID = id
			if displayMode != 2 {
				fmt.Printf("process ID:%d\n", pptr.ID)
			}
			if e.ASM(pfr, a, name, "") != -1 { // 用户程序 .co 生成在 cwd
				Load(pptr, name, "")
				removeCo(name) // 运行前即删中间 .co
				os_.loader(pptr)
			}
			return 0
		}

		if len(os.Args) >= 2 { // 命令行指定程序（可直接传 userfile 目录内文件名或任意路径）
			singleProc = true // 单进程模式：网络 read 用阻塞式
			prog := os.Args[1]
			// 若参数带路径分隔符或文件不存在于 userfile 目录，则当作直接路径打开
			var pfr *os.File
			var ret int
			if strings.ContainsAny(prog, `/\`) {
				// 带路径分隔符 → 直接按该路径打开（相对当前目录或绝对路径）
				pfr, ret = osOpen(prog)
			} else {
				// 仅文件名 → 优先当前目录，其次系统根目录（~/.xmips）的 userfile/ 目录
				if f, err := os.Open(prog); err == nil {
					pfr, ret = f, 0
				} else {
					pfr, ret = disk.getFile(prog, 0)
				}
			}
			if ret != 0 {
				fmt.Printf("cannot open %s\n", prog)
				return
			}
			pptr := newProcess(0, codeSize, dataSize, stackSize, 1)
			pptr.ID = 50
			if displayMode != 2 {
				fmt.Printf("process ID:%d\n", pptr.ID)
			}
			if e.ASM(pfr, a, prog, "") != -1 { // 用户程序 .co 生成在 cwd
				Load(pptr, prog, "")
				removeCo(prog) // 运行前即删中间 .co
				os_.loader(pptr)
			}
			pfr.Close()
		} else { // 从 run.list 读取（run.list 位于系统根目录 ~/.xmips/，而非 userfile/）
			run, err := os.Open(filepath.Join(runDir, runList))
			if err != nil {
				fmt.Println("cannot open run.list")
				return
			}

			scanner := bufio.NewScanner(run)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
			scanner.Split(bufio.ScanWords)

			// 读取第一个程序名
			if !scanner.Scan() {
				run.Close()
				return
			}
			name := scanner.Text()

			id := 50 // 用户进程起始 ID
			loaded := 0
			for name != "end" && name != "END" {
				loadFromDisk(name, id)
				id++
				loaded++
				if !scanner.Scan() {
					break
				}
				name = scanner.Text()
			}
			singleProc = loaded <= 1 // 仅 1 个文件视为单进程模式
			run.Close()
		}
	}
	//*************************************************************************************************************

	// 调度执行
	os_.swap2(im)

	//*************************************************************************************************************
	// 显示已结束进程的数据段
	{
		if displayMode == 2 {
			// displayMode=2：不显示任何额外内容，直接结束
			return
		}
		var pptr *Process
		r := os_.Finished.deQueue(&pptr)

		fmt.Print("\ndisplay data?(y/n): ")
		reader := bufio.NewReader(os.Stdin)
		c, _ := reader.ReadByte()
		fmt.Println()

		for r != -1 {
			fmt.Printf("process ID:%d\n", pptr.ID)
			if c == 'y' || c == 'Y' {
				fmt.Printf("memory ID:%d\n", pptr.MData.ID)
				for i := 0; i < pptr.MData.mSize; i++ {
					fmt.Printf("%-4d%d\n", i, pptr.MData.read(i))
				}
			}
			fmt.Printf("execute times:%d\n", pptr.exetime)
			r = os_.Finished.deQueue(&pptr)
		}
	}

	fmt.Println("\nPress any key to exit.")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
}
