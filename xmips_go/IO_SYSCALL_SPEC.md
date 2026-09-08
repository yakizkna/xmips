# Xmips 文件与网络 Socket 功能规格

> 目标：为 xmips_go 增加**文件 I/O**（真实主机目录映射）与**真实网络客户端 Socket**（TCP 客户端，client-only，无 bind/listen）。
> 定稿方向：**四个系统调用全部用 INT（15~18），由 dispatcher 原生执行**（不写 `.scp`、无通道寄存器），**网络走真实 TCP，可选 TLS，阻塞/非阻塞由单进程模式决定，均有超时保护**。
> 状态：已实现并验证（HTML 抓取、文件读写、防死锁均通过）。

---

## 1. 背景与既定约束

### 1.1 现状
- 系统调用：用户 `MOV #10 调用号; INT`，dispatcher 按 `r-10` 处理。现有 `INT 13`(输出串)/`INT 14`(输入缓冲) 亦由 **dispatcher 原生执行**（原为 `.scp`，已改造，见 §2 注）。仅 `INT 10`（冒泡排序）仍为 `.scp` 系统函数。
- 原生 syscall 先例：`SYSR/SYSW/WAKE/SET` 由 dispatcher 在 `swap2` 的 `case 5/6/7` 直接执行（不建系统函数进程）。
- 关键约束：`INT` 把调用者 `#0~#14`、PC、flag 压栈，恢复时从栈弹出。**因此 dispatcher 原生 syscall 的返回值必须写进调用者栈的 `#13` 槽位**（复用 `case 3` 的 `up()` 写回：`S.write(sp-4, #13)`），写在 GM 寄存器会被覆盖。
- `.scp`（ABC 汇编）只能寻址 `#0~#14`，够不到 `#15~#19`。这是放弃 `.scp` 于本功能的主因。

### 1.2 定稿决策
| 决策 | 内容 |
|------|------|
| 接口形式 | 全统一为 INT（15=open,16=read,17=write,18=close），用户视角一致 |
| 实现落点 | 全部 dispatcher/fsdev 原生（SYSR 风格），无 `.scp`、无 `#17/#18` 通道寄存器 |
| 文件存储 | 真实主机目录 `disk/` 映射 |
| 网络 | 真实 TCP 客户端；**可选 TLS**（`tlssock:` 强制 / `sock:` 端口 443 自动）；**阻塞/非阻塞由单进程模式决定，均有 `sockTimeout` 超时保护** |
| 范围 | 只做读/写/开关；bind/listen/server 一律不做 |

---

## 2. 接口（ABI）

统一：`#10`=调用号，`#11/#12`=参数，`#9`=fd/返回值。（均在 `#0~#14`，跨 INT 可靠）

| 调用号 | 功能 | 传入 | 返回 `#9`=fd/结果 |
|--------|------|------|-----------|
| 15 | open | `#11`=路径串地址(数据段,0结尾)、`#12`=模式 | fd（≥0）；<0 错误码 |
| 16 | read | `#11`=缓冲地址、`#12`=最大长度、`#9`=fd | 实际读到的字节数（0=暂无数据/EOF） |
| 17 | write | `#11`=缓冲地址、`#12`=写入长度、`#9`=fd | 实际写入字节数 |
| 18 | close | `#9`=fd | 0 成功；<0 错误码 |

> **fd 寄存器为 `#9`**：不使用 `#13`，因为 `#13` 是输出寄存器，用户 `MOV #13 x` 会触发 stdout 副作用。`#9` 为通用寄存器，无副作用，且在 INT 压栈/恢复范围（`#0~#14`）内，跨系统调用可靠。

- **模式 `#12`**：0=只读，1=写（截断），2=读写追加（仅文件有效）。
- **路径前缀**：
  - `"sock:host:port"` → TCP 客户端；**端口 443 自动启用 TLS**（`ServerName`=host，握手受 `sockTimeout` 约束）。cupa 无需关心 TLS。
  - `"tlssock:host:port"` → TCP 客户端，**强制 TLS**。
  - 其它 → 文件（限定在虚拟磁盘内）。
- 错误码（负值）：-1 open 失败/越界，-2 非法 fd，-3 socket 连接失败，-4 write 缓冲满，-5 socket 已关闭/read 出错，其余见 §7。
- `read` 对 socket 的 0 语义：**0 = 当前无数据（非阻塞直接返回）**，区别于"连接关闭"的负错误码。文件则 0=EOF。

用户程序示例（文件写后读回，fd 存 #1，用前移入 #9）：

```
MOV #11 PATH        ~ 数据段路径串
MOV #12 1           ~ 模式1=写
MOV #10 15
INT                 ~ open → #9=fd
MOV #1 #9           ~ 暂存 fd
MOV #11 BUF
MOV #12 LEN
MOV #9 #1           ~ 恢复 fd（写 #9 无副作用）
MOV #10 17
INT                 ~ write → #9=已写数
MOV #9 #1
MOV #10 18
INT                 ~ close（#9 为 fd）
```

---

## 3. 实现机制（dispatcher 原生）

### 3.1 `swap2` 新增 `case 15/16/17/18`
在 dispatcher 主循环的 switch 中、`default`(r>=10 走 .scp) 之前显式拦截 15~18。执行流程（仿 `case 5` SYSR）：

```
case 15..18:
    resp := fsdev.do(r-10, /#11 #12 #9 参数, 调用者 MData)
    // 把结果写进调用者栈的 fdReg(#9) 槽（复用 case3 的栈槽写回机制）
    sp := d.runb.pptr.S.SP
    val := resp; if resp == fdNoData { val = 0 }   // socket 无数据 → 用户读到 0
    d.runb.pptr.S.write(sp-(17-fdReg), val)        // fdReg=9 → sp-8
    if resp == fdNoData && !singleProc {
        d.readyb.enQueue(d.runb)                   // 多进程：放就绪队尾，让出时间片
    } else {
        d.readyb.insertToHead(d.runb)              // 有数据/单进程：回到队首继续
    }
    RES 打印; r = 1
```

- **无数据让出时间片**：多进程下 socket read 无数据（`fdNoData`）时，dispatcher 把进程**放回就绪队尾**再调度，其他进程插队先跑，本进程稍后回来重试——避免忙自旋的同时不阻塞时间片轮转。单进程则走 `insertToHead`（无其他进程，让出无意义）。

- 不创建系统函数进程，不占 PCB、无调度往返、无 `.scp`。
- open/read/write/close 可由任意用户进程调用（与 `INT 13/14` 一致，**无需** SYSR 的优先级守卫）。

### 3.2 为什么返回值写栈槽
`INT` 已在调用者栈压入 `#0~#14`；恢复现场时 exer 会完整弹出覆盖 GM。故写 GM 无效。写栈对应槽位才是恢复后真正生效的结果：GM[n] 帧内槽 = `sp-(17-n)`，fdReg(9)=`sp-8`。此机制与现有 `case 3` INT 14 写回（#13=`sp-4`）完全同源。

### 3.3 备选（记录，不实现）
将 read/write 改为 `.scp` 逐字节（沿袭 13/14）以增强教学，但需在 `#0~#14` 腾出字节通道寄存器且 open 返回 fd 也需寄存器，增加复杂度和出错面。定稿放弃，保持统一原生。

---

## 4. fsdev 与 fd 表（新增 `fsdev.go`）

```
type fdEntry struct {
    kind   int          // FILE 或 SOCK(TCP)
    path   string       // 路径或 sock:host:port
    host   *os.File     // 文件
    mode   int
    offset int64        // 文件读写偏移（独立维护）
    conn   net.Conn     // socket 连接
}
```
- fd 从 3 起（0/1/2 预留 stdio 语义）。
- 文件字节 ↔ 数据段字节一一对应（一个 int 存一个字节低 8 位）。
- 并发安全：模拟器为单线程，无需加锁。

### fsdev 操作
- **open**：前缀 `sock:`/`tlssock:` → `net.DialTimeout("tcp", host:port, sockTimeout)`；`tlssock:` 或端口 443 时用 `crypto/tls` 包裹并握手（`ServerName`=host，握手受 `sockTimeout` 约束）；否则文件 → 「路径清洗 + `diskRoot` 限定」后 `os.OpenFile`。
- **read**：文件 → 按 `offset` 用 `ReadAt`；socket → 阻塞等待/非阻塞由单进程模式决定，超时返回 `fdNoData`（见 §5）。
- **write**：文件 → `WriteAt` 更新 `offset`；socket → 受 `sockTimeout` 写超时，返回实际字节。
- **close**：关闭 host/conn，释放 fd（标记槽可复用）。

---

## 5. 真实网络：阻塞/非阻塞语义（由单进程模式决定）

读是否阻塞取决于**单进程模式**：

单进程模式 = `./xmips xxx.cupa` 直接运行，或 `run.list` 恰有 **1** 个文件（startup 计算，写入全局 `singleProc`）。

| 场景 | `read`(socket) | write | connect/握手(open) |
|------|--------------|-------|--------------------|
| **单进程** | **阻塞**等待数据，受 `sockTimeout` 超时保护，**超时返回 0**（不再永久卡死） | 受 `sockTimeout` 写超时 | 超时 `sockTimeout` |
| **多进程** | **非阻塞**：`SetReadDeadline(now)`；有数据返回 N，**无数据返回 0 并让出时间片**（放就绪队尾） | 受 `sockTimeout` 写超时 | 超时 `sockTimeout` |

- **多进程让出时间片**：socket read 无数据时返回哨兵 `fdNoData(-6)`，dispatcher 识别后把进程放回**就绪队尾**（`enQueue`），再继续调度其它进程，本进程稍后被重新调度回来重试——既非忙自旋，也不阻塞轮转。
- **单进程超时**：`read` 设 `SetReadDeadline(now+sockTimeout)`；超时同样返回 `fdNoData`，dispatcher 转成 0 写回（无其它进程可让出，走 `insertToHead`）。防服务器 keep-alive 不关连接导致的永久阻塞。
- 不用固定短窗口（如 0.1ms）：多进程用 `deadline=now`（**真正的即刻返回**），单进程用 `sockTimeout` 绝对超时。
- 连接关闭/IO 错误 → 返回 `-5`，区别于"无数据返回 0"；读后**复位 deadline**。
- 结果依赖主机真实网络，不再完全可复现（定稿接受）。

---

## 6. 虚拟磁盘（配置项）

config.ini 新增：

| 参数 | 值 | 说明 |
|------|---|------|
| diskRoot | `disk` | 虚拟磁盘目录（相对运行目录），自动创建 `Xmips/disk/` |
| sockTimeout | `2000` | socket connect/握手/读写超时（毫秒） |

`open` 的文件路径强制限定 `diskRoot` 内：`filepath.Clean` + 拒绝绝对路径/`..` 越界（§7 错误 -1）。

---

## 7. 错误码与 RES（新增）

`#9` 返回错误码（负值）：-1 open 失败/越界，-2 非法 fd，-3 socket 连接/握手失败，-4 write 缓冲满，-5 socket 关闭/read-write 出错。哨兵 `fdNoData=-6` 仅内部使用（socket 读无数据），写回用户时统一转 0。

新增系统状态信息表项：

| RES | Level | Message |
|-----|-------|---------|
| 200 | 3 | fsdev::open: path escape / open failed |
| 201 | 3 | fsdev::read: invalid fd |
| 202 | 3 | fsdev::write: invalid fd / write failed |
| 203 | 3 | fsdev::close: invalid fd |
| 204 | 3 | fsdev::net: socket connect / io error |

---

## 8. 需要修改的 Go 文件

| 文件 | 改动 |
|------|------|
| `src/fsdev.go`（新增） | fd 表、文件/Socket 实现、路径清洗、TLS（`crypto/tls`）、阻塞/非阻塞/超时读写、哨兵 `fdNoData` |
| `src/dispatcher.go` | `swap2` 增加 `case 15/16/17/18`；`#9` 栈槽写回；`fdNoData` 多进程放队尾让出时间片 |
| `src/main.go` | 解析 `diskRoot/sockTimeout`；计算 `singleProc` 并注入 fsdev、全局常量 |
| `src/global.go` | 全局 `diskRoot/sockTimeout/singleProc` |
| `Xmips/config.ini` | 新增 §6 配置项 |
| `README.md` / `使用说明.txt` | 补系统调用表（15~18）、TLS、虚拟磁盘、阻塞/超时/让出说明 |

`assembler.go`、`interpreter.go` **无需改动上的新增 opcode/系统函数需求**；`sysfun/INT_xx.scp` **仅 `INT_10.scp`**（冒泡排序）保留，`INT_13.scp`/`INT_14.scp` 已删除（改由 dispatcher 原生执行）。

---

## 9. 验证结果

1. 文件写 → 读回：写 `disk/test.txt`，再 open 同文件 read 回显比对 ✅（`file/FILEIO.cupa`）
2. 越界防护：`open "../etc/passwd"` 拒绝（错误 -1）。
3. **HTTPS 抓取 + 写文件** ✅（`file/NETTOUR.cupa`）：`sock:ace.yakidev.top:443` 自动 TLS → 发 HTTP GET → 分块 read → 写 `disk/tour.html`，22KB 完整 HTML。
4. **单进程防死锁** ✅（`file/HTIMEOUT.cupa` + 本地静默服务器，accept 后 30s 不发数据不关闭）：程序约 4s 内退出，证明 `sockTimeout` 读超时生效。
5. 回归：`SUM`、`BUBBLE_INT`、`FILEIO` 正常；`INT 13/14` 不受影响 ✅
6. 字长 / 多进程让出时间片：待补充专项验证。

---

## 10. 待定/评审点

- [x] `sockTimeout` 默认 2000ms 是否合适（已用；单进程超时/多进程非阻塞均受其保护）。
- [ ] 非阻塞 read 的"无数据"通过 `fdNoData` 返回哨兵，dispatcher 在**多进程时放队尾让出**——仅靠 `enQueue` 在就绪队列只有 1 个进程时等于空转，可接受（多进程模式通常多进程）。
- [x] 错误码为负值、成功为 ≥0 的约定。
- [x] fd 起始 3（预留 0/1/2=stdio）。
- [ ] `read` 的 0 语义：文件=EOF，socket=暂无数据。若需文件也区分"读到 0 字节"与"EOF"，目前统一按 0 处理。