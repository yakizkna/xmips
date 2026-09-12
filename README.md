# 一个用软件实现的模拟指令系统及其汇编语言 —— Xmips 指令系统与 ABC 汇编语言

## 目录

- [一、概述](#一概述)
- [二、Xmips 指令系统](#二xmips-指令系统)
- [三、ABC 汇编语言](#三abc-汇编语言)
- [四、汇编过程](#四汇编过程)
- [五、执行过程](#五执行过程)
- [六、Xmips 操作实例](#六xmips-操作实例)
- [七、其他功能简介](#七其他功能简介)
- [附录 A：操作码和助记符对照表](#附录-a操作码和助记符对照表)
- [附录 B：系统运行状态信息表](#附录-b系统运行状态信息表)

---

## 一、概述

### 1.1 关于 Xmips

Xmips 是一个旨在模拟多进程调度和资源分配的程序（**早期用 C++ 实现，现已用 Go 完全重写**）。Xmips 系统组成部分目前主要包括汇编程序部分、指令解释和执行部分和进程调度部分。

Xmips 是一个单线程的控制台程序。要用它模拟多进程并发，必须为其设计一套模拟的指令系统和程序设计语言。为了实现这一目标，仿照 X86 汇编语言的形式，为 Xmips 系统设计了一套汇编语言指令，命名为 Stimulated Assembly Language By C++（以下简记为 ABC 汇编语言，名字沿用早期 C++ 实现的渊源），并为其编写了一个简单的汇编程序，用于将汇编源程序翻译成可以被系统执行的机器码。

运行在 Xmips 系统上的用户程序均是由这种汇编语言编写的，而 Xmips 的系统程序和系统函数库部分用 ABC 汇编语言编写，其余则用 Go 编写（早期为 C++），对于这部分内容，可以理解为 Xmips 系统模拟的"硬件功能"。

为了简化，Xmips 中所有的数据单元都定义为整型（int），即指令的每一个字段都是一个整型量，模拟的内存以 int 为编址单位，所有寄存器的大小也都是 int。所以目前 Xmips 系统的汇编语言是一种无类型的语言，不支持字符类型和浮点类型，一切操作都针对整型量。

### 1.2 最早的 Xmips 指令系统

最早的 Xmips 系统中没有设计寄存器，指令的执行实际上是根据代码段的内容直接修改数据段。不设寄存器的原因是因为认为在纯软件模拟的条件下不存在寄存器和主存的访问速度差别。但是在主存中增加了一个用于保存操作码和操作数地址等信息的区域。在后来的设计中，该区域被取消，其功能由系统提供的寄存器组取代。

最初开始设计的时候，ABC 汇编语言和机器指令从格式上是严格对应的：

```
操作码  目的操作数地址  源操作数地址
```

对于单操作数的指令和无操作数的指令，操作数地址缺省为 0。

这种方法的寻址方式单一，仅仅支持直接寻址。为了支持立即数寻址，只能专门增加一组针对立即数的指令，其操作码助记符形式为 `IMXXX`，如 `IMADD SUM 9`，表示将立即数 9 加到符号地址 SUM 所对应的主存单元中，即 `SUM ← (SUM)+9`。

由于某些程序中需要间接寻址，用 `@` 符号表示实际的地址为 @ 后地址单元中的值，如 `MOV SUM @I`，表示 `SUM ← ([I])`。

### 1.3 改进后的 Xmips 指令系统

在学习了组成原理指令系统后，对原来的 Xmips 指令系统和 ABC 汇编语言做了一些改进。

1. **新增寄存器**：设计了两个寄存器组 MRgst 和 GM，以及一些供系统使用的寄存器，如指令计数器 PC、状态寄存器 flag 等。
2. **指令格式与 ABC 汇编语言格式分离，增加了寻址方式字段**：ABC 汇编语言格式与原来相同，但增加了用于偏移寻址和间接寻址的操作符 `&` 和 `@`，以及寄存器标识符 `#0~#14`。指令格式由四个定长的字段构成（每个字段为一个整型量）：

```
操作码  寻址方式  目的操作数形式地址  源操作数形式地址
```

支持的寻址方式包括立即数寻址、直接寻址、间接寻址、寄存器寻址、寄存器间接寻址、变址寻址等。对于双操作数的情况，寻址方式字段由目的操作数和源操作数的寻址方式拼接而成。

3. **细分了解释执行指令的流程**：同一类数据通路的指令的取数和存数操作统一完成。操作的一般形式为：
   - `<1>` 取指令（取操作码、取寻址方式、取目的操作数形式地址、取源操作数形式地址，PC+1）
   - `<2>` 计算操作数有效地址
   - `<3>` 取操作数（可能在寄存器中、主存中、立即数或操作数就是地址）
   - `<4>` 执行运算操作
   - `<5>` 存操作数
   - `<6>` 取下一条指令

---

## 二、Xmips 指令系统

Xmips 系统中的基本数值单位是整型（int），内存单元以 int 编址，所有指令的长度和寄存器（组）的长度都只能是 int 的整数倍。

### 2.1 Xmips 中的寄存器

Xmips 中的寄存器分为两类：

**内部寄存器**（仅供系统使用，用户不可访问）：
- 指令计数器 PC
- 状态寄存器 flag
- 寄存器组 GM 中编号为 #15、#16 的寄存器（数据寄存器）
- 寄存器组 MRgst 中的所有寄存器

**通用寄存器**（用户可以使用）：
- 15 个整型通用寄存器，编号 #0~#14，全部在寄存器组 GM 中
- 其中 #1~#9 号寄存器可用于变址寻址

| 寄存器组 | 序号 | 功能说明 | 可使用者 |
|---------|------|---------|---------|
| PC | 0 | 指令计数器 | 系统 |
| flag | 0 | 状态寄存器 | 系统 |
| MRgst | 0 | 保存操作码 | 系统 |
| MRgst | 1 | 保存寻址方式（双操作数时分解为目的/源寻址方式） | 系统 |
| MRgst | 2 | 保存目的操作数形式地址和有效地址 | 系统 |
| MRgst | 3 | 保存目的操作数所在位置信息（0=主存, 1=寄存器, 2=立即数） | 系统 |
| MRgst | 4 | 保存源操作数形式地址和有效地址 | 系统 |
| MRgst | 5 | 保存源操作数所在位置信息 | 系统 |
| MRgst | 6~14 | 未定，留作扩展 | 系统 |
| GM | 0 | 通用寄存器 | 系统/用户 |
| GM | 1~9 | 通用寄存器，可作偏移寻址的变址寄存器（格式 `&R`） | 系统/用户 |
| GM | 10 | 通用寄存器，保存中断号和中断返回信息 | 系统/用户 |
| GM | 11~12 | 通用寄存器，用于系统功能调用时传参 | 系统/用户 |
| GM | 13 | **输出寄存器**，写入即输出字符到 stdout | 系统/用户 |
| GM | 14 | 仅 INT 14（输入）由 dispatcher 原生读取 stdin 时使用；**用户程序读取 #14 报错 RES=198 非法访问** | 系统 |
| GM | 15 | 数据寄存器，保存目的操作数的数值 | 系统 |
| GM | 16 | 数据寄存器，保存源操作数的数值 | 系统 |
| GM | 17~19 | 未定，留作扩展 | 系统 |

### 2.2 Xmips 指令格式

Xmips 指令共有 4 个字段，每个字段的大小为一个整型量：

```
操作码  寻址方式  目的操作数形式地址  源操作数形式地址
```

**1）操作码字段**

操作码（opcode）可以分解为三个部分：
- **操作标识（op）**：操作码的其他数字，与操作一一对应
- **数据操作类型标识（d）**：操作码的十位数字，表示存取操作数的方式
- **操作数数目标识（t）**：操作码的个位数字（双操作数 t=2，单操作数 t=1，无操作数 t=0）

例如 `ADD #2 SUM` 的操作码为 `900272`：
- op=9002（加法操作）
- d=7（取源和目的操作数的数值，结果保存到目的操作数）
- t=2（两个操作数）

**2）寻址方式字段**

用一个整型量表示寻址方式（addressing）。对于双操作数，寻址方式 = addressing_d + addressing_s × 1000。

对于单一操作数，寻址方式为一个 3 位十进制整数 ABC：
- R = A = addressing / 100（百位数表示偏移寻址的寄存器号，范围 1~9）
- type = BC = addressing mod 100（十位和个位表示寻址方式类型）

type 对应一个二进制数（S0S1S2S3）：
- S0=1 表示偏移寻址
- S1=1 表示间接寻址
- S2S3=00 表示操作数在主存
- S2S3=10 表示操作数在寄存器
- S2S3=11 表示操作数为立即数

**3）操作数形式地址字段**

- 若操作数在数据段（主存），段内偏移地址为 M，则 oa = M
- 若操作数在通用寄存器 R，则 oa = R
- 若操作数为立即数 I，则 oa = I（此时寻址方式字段为 3）

### 2.3 Xmips 的寻址方式

**1）支持的寻址方式**：
- 寄存器寻址
- 寄存器间接寻址
- 直接寻址
- 间接寻址
- 基址或变址寻址（有效地址 = 某通用寄存器值 + 指令中常量）
- 立即数寻址

**2）有效地址的计算**

ABC 汇编语言地址形式：`[&R][@]K`

| addressing | S0S1S2S3 | 汇编地址形式 | 有效地址计算 | 操作数位置 |
|-----------|----------|------------|------------|----------|
| 0 | 0000 | SUM, !0 | EA=D | 主存 |
| 2 | 0010 | #2 | EA=#D | 寄存器 |
| 3 | 0011 | 6 | 不存在 | 立即数 |
| 4 | 0100 | @ST, @!10 | EA=(D) | 主存 |
| 6 | 0110 | @#2 | EA=(#D) | 主存 |
| 8+100R | 1000 | &3SUM, &9!21 | EA=(#R)+D | 主存 |
| 12+100R | 1100 | &5@ST | EA=(#R)+(D) | 主存 |
| 14+100R | 1110 | &6@#5 | EA=(#R)+(#D) | 主存 |

### 2.4 地址空间说明

目前的 Xmips 缺少内存管理机构和地址变换机构，实际上相当于每一个进程的代码段、数据段和堆栈段都是一个物理上独立的存储器：

- 程序的逻辑地址空间就是存储器的物理地址空间
- 代码段、数据段和堆栈段分别对应三个存储器，每个段的起始地址为 0

在源代码中，用 `MCode` 表示代码段，`MData` 表示数据段，`S` 表示堆栈段。

---

## 三、ABC 汇编语言

### 3.1 一个简单的例子——SUM

```
~ 计算 SUM=1+2+...+10
DIM SUM 0
MOV #1 1
: L1
ADD SUM #1
INC #1
CMP #1 11
JB L1
END
```

汇编后生成的字符码文件：

```
900152 3002 1 1
900272 2000 0 1
900961 2 1 0
900332 3002 1 11
900521 3 4 0
900000
1 0
```

### 3.2 寻址方式

| 寻址方式 | 格式 | 说明 |
|---------|------|------|
| 寄存器寻址 | `#R` | 寄存器 #R 的内容即为操作数 |
| 寄存器间接寻址 | `@#R` | 寄存器 #R 的内容为操作数的偏移地址 |
| 直接寻址 | `S` 或 `!K` | S 为变量名；K 为数据段地址整数 |
| 间接寻址 | `@S` 或 `@!K` | 形式地址是操作数地址的地址 |
| 偏移寻址 | `&RX` | 有效地址 = (#R) + X 的形式地址 |
| 立即数寻址 | `n` | 形式地址字段就是操作数 |

组合限制：
- 立即数寻址方式中不可以有其他成分，如 `@8` 是错误的
- 偏移寻址的 X 部分不可以是寄存器寻址，如 `&5#3` 是错误的

### 3.3 机器指令语句和伪指令

**1）变量定义语句**

```
DIM S [Dn] 初值1, 初值2, ..., 初值K;
```

示例：
```
DIM SET D5 1,2,3;    ~ 定义大小为5的数组，前三个元素赋初值
DIM MAX 100           ~ 定义变量 MAX，初值为 100
```

**2）标号定义语句**

```
: L
```

**3）机器指令语句**

| 类型 | 格式 | 功能 |
|------|------|------|
| 数据传送 | `MOV OPD OPS` | OPD ← (OPS) |
| 地址传送 | `LEA OPD OPS` | OPD ← OPS |
| 加法 | `ADD OPD OPS` | OPD ← (OPD) + (OPS) |
| 加1 | `INC OPD` | OPD ← (OPD) + 1 |
| 减法 | `SUB OPD OPS` | OPD ← (OPD) - (OPS) |
| 求相反数 | `NEG OPD` | OPD ← -(OPD) |
| 比较 | `CMP OPD OPS` | (OPD) - (OPS)，结果保存到 flag |
| 乘法 | `MUL OPD OPS` | OPD ← (OPD) * (OPS) |
| 除法 | `DIV OPD OPS` | OPD ← (OPD) / (OPS)，#0 ← (OPD) % (OPS) |
| 取模 | `MOD OPD OPS` | OPD ← (OPD) % (OPS) |
| 与 | `AND OPD OPS` | OPD ← (OPD) & (OPS) |
| 或 | `OR OPD OPS` | OPD ← (OPD) \| (OPS) |
| 非 | `NOT OPD` | OPD ← ~(OPD) |
| 异或 | `XOR OPD OPS` | OPD ← (OPD) ^ (OPS) |
| 大于跳转 | `JA L` | flag > 0 时 PC=L |
| 小于跳转 | `JB L` | flag < 0 时 PC=L |
| 无条件跳转 | `JMP L` | PC=L |
| 等于跳转 | `JE L` | flag == 0 时 PC=L |
| 不等于跳转 | `JNE L` | flag != 0 时 PC=L |
| 压栈 | `PUSH OPD` | (OPD) 压栈 |
| 出栈 | `POP OPD` | OPD ← 出栈结果 |
| 寄存器组压栈 | `PUSHA` | #0~#14 压栈 |
| 寄存器组出栈 | `POPA` | #0~#14 出栈 |
| 函数调用 | `CALL FUNCTION` | 调用函数 FUNCTION |
| 函数返回 | `RET` | 返回调用函数 |
| 软中断 | `INT` | 中断号在 #10 中 |
| 中断返回 | `IRET` | 返回码在 #10 中 |
| 关中断 | `CLI` | 禁止时钟中断抢占 |
| 开中断 | `STI` | 允许时钟中断抢占 |
| 读系统共享区 | `SYSR` | #11=起始地址, #12=长度（仅系统函数） |
| 写系统共享区 | `SYSW` | #11=起始地址, #12=长度（仅系统函数） |
| 唤醒进程 | `WAKE` | #11=通信标识符, #12=等待队列号（仅系统函数） |
| 设置通信标识 | `SET` | #11=通信类型, #12=通信标识符值（仅系统函数） |
| 结束 | `END` | 汇编结束 |
| 停机 | `HALT` | 程序结束 |

**4）伪指令**

| 伪指令 | 功能 |
|--------|------|
| `~ 注释` 或 `CMNT 注释` | 注释（中间不可有分隔符） |
| `PROC FUNCTION` | 函数定义（从下一行开始，至 RET 结束） |
| `: L` | 标号定义 |
| `DIM S ...` | 变量定义 |
| `$` | 占位空语句 |

> 注意：操作符支持纯大写或纯小写，但混合使用是非法的。

### 3.4 系统功能调用简介

Xmips 系统提供了一组系统函数，用户通过软中断语句 `INT` 调用，调用前将调用号传送到寄存器 #10 中，函数参数传给相应寄存器。

示例程序 BUBBLE_INT.cupa（通过 10 号系统调用对 10 个数排序）：

```
~ 定义数组 SET
DIM SET D10 2,4,51,13,17,40,5,22,3,16;
~ 传送数组首地址到 #11
MOV #11 0
~ 传送待排序数的个数到 #12
MOV #12 10
~ 传送调用号到 #10
MOV #10 10
INT
END
```

运行结果（保存在进程的数据段中）：

```
process ID:51  memory ID:1
0   51
1   40
2   22
3   17
4   16
5   13
6   5
7   4
8   3   2
```

### 3.5 函数调用（CALL/RET）

Xmips 通过 `CALL`/`RET` 支持子程序调用（无栈帧、无局部变量，参数靠寄存器约定传递，如 `#0`=数组首地址、`#1`=长度）。`CALL 标签` 将返回地址 PC 与标志压栈后跳转，`RET` 弹栈返回。

编写含子程序的程序需注意：

- 执行入口在代码区首地址(0)，**主程序须放在文件最前**（DIM 之后）；
- 汇编器在第一个 `END` 终止，故**全文件只能有一个 `END`**（放末尾），主程序结束改用 `HALT`（等效正常结束、不终止汇编）；
- 标签须以 `: name` 前缀定义；`CALL` 可前向引用其后定义的标签。

完整示例见 [`file/BUBBLE_CALL.cupa`](dist/userfile/BUBBLE_CALL.cupa)：用两个用户子程序完成冒泡排序并输出——

```text
51 40 22 17 16 13 5 4 3 2
```

与上面 `BUBBLE_INT.cupa`（10 号系统调用）结果一致。更详细的约束见 [`USAGE.md`](USAGE.md) 的 5.1/5.2。

### 系统调用一览

| 调用号 | 功能 | 参数 | 说明 |
|--------|------|------|------|
| 10 | 冒泡排序 | #11=数组起始地址, #12=长度 | 排序后结果存回数据段 |
| 13 | 输出字符串 | #11=字符串起始地址 | **dispatcher 原生**从 #11 地址读数据段，读到 0 为止，逐字符输出到 stdout |
| 14 | 输入到缓冲区 | #11=缓冲区起始地址, #12=缓冲区大小 | **dispatcher 原生**从 stdin 逐字符读入存入缓冲区（不再经 `#14`），读到 EOF(0) 或缓冲区满结束；返回 #13=0 读到 EOF，#13=1 缓冲区满（可再次 INT 14 继续读） |
| 15 | open 文件/Socket | #11=路径串地址(0结尾), #12=模式 | `<0` 错误码；≥0 返回 fd（文件描述符） |
| 16 | read 读数据 | #11=缓冲地址, #12=最大长度, #9=fd | 0=暂无数据/EOF；N=实际读到的字节数 |
| 17 | write 写数据 | #11=缓冲地址, #12=写入长度, #9=fd | N=实际写入字节数 |
| 18 | close 关闭 | #9=fd | 0 成功；<0 错误码 |

系统调用 15~18（文件与网络 Socket）由 **dispatcher 原生执行**（无 `.scp`、不占用 PCB），返回值统一写回 `#9`。通信细节见下方「四、文件与网络」。

写入 #13（输出寄存器）即输出对应字符到 stdout，无需系统调用。INT 13 用于批量输出字符串（数据段中以 0 结尾的字符序列）。

读取 #14（输入寄存器）：**用户程序或其他系统函数读取 #14 会触发系统报错（RES=198, Error）**。用户程序只能通过 `INT 14`（dispatcher 原生）从 stdin 间接获取输入。

INT 13 输出示例：

```
~ 数据段 MSG 起始地址为 0（"hello\0"）
MOV #11 0
MOV #10 13
INT
```

INT 14 输入示例：

```
~ 数据段 NAME 起始地址为 20，大小 40
MOV #11 20
MOV #12 40
MOV #10 14
INT
MOV #0 #13      ~ #13=1 表示缓冲区满，可再次 INT 14 继续读
```

#### 文件与真实网络 Socket（INT 15~18）

`open` 的路径串前缀决定对象（`#11` 为数据段中 0 结尾字符串地址）：

| 路径前缀 | 含义 |
|---------|------|
| `sock:host:port` | TCP 客户端连接；**端口 443 自动启用 TLS** |
| `tlssock:host:port` | TCP 客户端连接，**强制 TLS**（`ServerName`=host，握手受 `sockTimeout` 约束） |
| 其它 | 文件，映射到**真实宿主文件**：`#11` 为文件路径（相对当前工作目录或绝对路径），直接读写；`sock:`/`tlssock:` 为 TCP 客户端（`sock:443` 或 `tlssock:` 自动启用 TLS） |

`#12` 模式：0=只读，1=写（截断），2=读写追加（文件有效）。返回 fd 存 `#9`（fd 从 3 起）。读写网络/文件的字节流：`INT 16` 从 `#9` 读 `#12` 字节存入 `#11` 指向的数据段；`INT 17` 把 `#11` 处 `#12` 字节写入 `#9`。Cupa 可据此实现真正的 **HTTPS 抓取 + 写文件**（例见 `userfile/NETTOUR.cupa`）。

读写阻塞语义由**单进程模式**决定（`./xmips xxx.cupa` 直跑，或 `run.list` 恰有 1 个文件）：

| 模式 | socket `read`（INT 16） |
|------|------------------------|
| 单进程 | 阻塞等待数据，受 `sockTimeout`（默认 2000ms）超时保护，**超时返回 0**，不会永久卡死 |
| 多进程 | 非阻塞（`deadline=now`）：有数据立即返回 N，**无数据返回 0 并让出时间片**（dispatcher 将进程放回就绪队尾，避免忙自旋） |

`socket write` 与 connect/握手均受 `sockTimeout` 约束。错误码（负值返回 `#9`）：`-1` open 失败/越界，`-2` 非法 fd，`-3` 连接失败，`-4` write 缓冲满，`-5` socket 关闭/读写出错。

---

## 四、汇编过程

### 4.1 从源程序到机器指令

汇编器 ASM 将字符串形式的汇编语句转换成数字形式的机器码。汇编过程中涉及四张表：

- **Opcode Table**：记录汇编操作符对应的机器码
- **Variable Table**：登记 DIM 语句定义的变量名和数据段首地址
- **Tag Table**：登记标号语句定义的标号和代码段地址
- **Jump Table**：登记跳转语句引用的标号和跳转语句地址

ASM 工作流程：

1. 查找 Opcode Table，将操作助记符转换成机器码
2. 如果是变量定义语句，将变量计入 Variable Table
3. 如果是标号语句，将标号登记到 Tag Table
4. 如果是跳转指令，将标号登记到 Jump Table
5. 分析操作数，计算寻址方式和形式地址
6. 遇到 END 语句，汇编结束
7. 汇编结束后，根据 Tag Table 和 Jump Table 修正跳转指令的目的操作数

### 4.2 SUM 程序的汇编过程

逐行汇编后生成的机器码（含标号修正和数据信息）：

```
900152 3002 1 1      ~ MOV #1 1
900272 2000 0 1      ~ ADD SUM #1
900961 2 1 0         ~ INC #1
900332 3002 1 11     ~ CMP #1 11
900521 3 4 0         ~ JB L1（4 为标号 L1 的代码段地址）
900000               ~ END
1 0                  ~ 数据段：1 个变量，初值 0
```

---

## 五、执行过程

### 5.1 执行机构

Xmips 的执行机构模拟计算机的处理器，包括：
1. 寄存器（组）：PC、flag、MRgst、GM
2. 译码机构：分析操作码，选择数据通路和运算操作
3. 有效地址计算机构
4. 访存机构：从主存取数到数据寄存器和将运算结果返回主存
5. 运算机构：对数据寄存器的内容执行运算

一条指令的执行过程：
1. 取指令，PC+1
2. 根据寻址方式和形式地址计算操作数有效地址
3. 分析操作码，选择数据通路
4. 根据数据通路将要运算的值传到数据寄存器
5. 执行运算
6. 根据数据通路选择是否把运算结果回存

### 5.2 数据通路

用一个整型量 dataLS 表示指令的数据通路（操作码的十位数字），V0V1V2 为其二进制值：

| dataLS | V0V1V2 | 数据通路 | 操作符举例 | 操作码 |
|--------|--------|---------|----------|--------|
| 7 | 111 | D ← (D) + (S) | ADD | 900272 |
| 6 | 110 | D ← (D) + 1 | INC | 900961 |
| 5 | 101 | D ← (S) | MOV | 900152 |
| 4 | 100 | D ← S | LEA / POP | 902042 / 902241 |
| 3 | 011 | (D) - (S) | CMP | 900332 |
| 2 | 010 | (D) / 堆栈 ← (D) | JMP / PUSH | 900621 / 902121 |
| 0 | 000 | 无操作数 | 其余 | XXXX00 |

- V0=1：计算结果保存到目的操作数地址
- V1=1：取目的操作数的数值到数据寄存器
- V2=1：取源操作数的数值到数据寄存器

### 5.3 程序 SUM 的执行过程

SUM 的字符码文件：

```
900152 3002 1 1    900272 2000 0 1    900961 2 1 0
900332 3002 1 11   900521 3 4 0       900000
1 0
```

执行流程：

1. **MOV #1 1**：PC=0→4, (#1)=1
2. **ADD SUM #1**：PC=4→8, MData[0]=1
3. **INC #1**：PC=8→12, (#1)=2
4. **CMP #1 11**：PC=12→16, flag=-1
5. **JB L1**：PC=16→4（跳转，因 flag<0）
6. 重复 2~5 直到 (#1)=11，此时 MData[0]=55
7. **END**：程序结束，最终结果 55 保存在 SUM 单元中

---

## 六、Xmips 操作实例

### 6.1 路径设置和参数配置

**仓库目录结构**：

```
xmips/
├── src/                    # Go 源代码（main.go、dispatcher.go、interpreter.go ...）
├── tool.sh                 # 工具脚本：./tool.sh build / clean
├── dist/                   # 运行目录（build 时同步到 ~/.xmips/，可执行文件不保留在此）
│   ├── config.ini          # 配置文件（缺省时用内置默认值）
│   ├── run.list            # 运行列表文件
│   ├── userfile/           # 用户程序目录
│   └── sysfun/             # 系统函数库目录（用户不可修改）
├── USAGE.md                # 运行与参数说明
├── IO_SYSCALL_SPEC.md      # 文件/Socket 系统调用规格
└── tools/                  # 辅助工具（如 md5gen）
```

**系统函数**：`dist/sysfun/` 下的 `.scp` 源文件统一命名为 `INT_xx.scp`，其中 `xx` 为调用号（对应 #10 中的值）。**仅 `INT 10`（冒泡排序）用 `.scp` 系统函数实现**；`INT 13/14`（输出/输入）与 `INT 15~18`（文件/Socket）均由 **dispatcher 原生执行**（无 `.scp`、不占 PCB）：

| 调用号 | 实现方式 | 功能 |
|--------|---------|------|
| 10 | `.scp` 系统函数 | 冒泡排序 |
| 13 | dispatcher 原生 | 从 #11 起始地址读字符串（0 结束），逐字符输出到 stdout |
| 14 | dispatcher 原生 | 从 stdin 逐字符读入 #11 缓冲（#12 大小）；返回 #13=0 EOF，#13=1 缓冲满 |
| 15~18 | dispatcher 原生 | 文件/Socket open / read / write / close |

**run.list 文件格式**：

```
文件名1
文件名2
...
end
```

**config.ini 配置参数**：

| 参数 | 值 | 说明 |
|------|---|------|
| delayMode | 0 | 关闭访存延时模拟 |
| delayMode | 1 | 开启访存延时模拟 |
| displayMode | 0 | 汇编过程中不显示源程序 |
| displayMode | 1 | 汇编过程中显示源程序 |
| displayMode | 2 | 运行过程中只显示标准输出，系统错误显示错误信息，其余全部屏蔽 |
| reportLevel | 0 | 显示警告和错误信息 |
| reportLevel | 1 | 显示主要运行状况 + 警告 + 错误 |
| reportLevel | 2 | 显示全部信息 |
| updateSysfun | — | 已由独立命令 `./xmips update` 取代（重建系统函数库，生成 .co） |
| codeSize | 800 | 用户进程代码区长度（int整型） |
| dataSize | 200 | 用户进程数据区长度（int整型） |
| stackSize | 50 | 用户进程栈区长度（int整型） |
| sysfunCodeSize | 800 | 系统函数代码区长度（int整型） |
| sysfunDataSize | 10 | 系统函数数据区长度（int整型，运行时指向调用者数据段） |
| sysfunStackSize | 40 | 系统函数栈区长度（int整型） |
| cycleTimes | 30 | 时间片大小：每个时间片最多执行的指令数 |
| pcbNum | 40 | PCB 数量：系统最大并发进程数 |
| bitMode | 64 | 机器字长：64=64 位整数（默认，溢出不回绕）；32=32 位整数（算术溢出按 2^32 回绕，SHL/SHR/ROL/ROR 按 32 位运算） |

config.ini 格式要求：每行 `key=value`（无空格），以 `end` 结尾；以 `;` 开头的行为注释，参数值后可跟 `;` 行内注释。

### 6.2 编辑汇编源程序

用户源程序必须建立在 `dist/userfile/` 目录下，以 `END` 语句结尾。

### 6.3 运行 xmips

编译并运行方式（详见 [`USAGE.md`](USAGE.md)）：

```
./tool.sh build       # 1. 编译；把 xmips 安装到 /usr/local/bin/xmips，dist/* 同步到 ~/.xmips/，重建系统函数库
./tool.sh clean       # 2. 卸载：rm -rf ~/.xmips；rm /usr/local/bin/xmips
```

运行（**可从任意目录直接调用**，二进制在 `/usr/local/bin/xmips`、数据在 `~/.xmips/`）：

```
xmips                      # A. 无参数 → 从 ~/.xmips/run.list 读取要运行的程序
xmips XXX.cupa             # B. 指定程序：优先当前目录，其次 ~/.xmips/userfile/ 目录
xmips /绝对/路径/XXX.cupa   # C. 指定程序的绝对/相对路径
xmips update               # D. 重建系统函数库（重汇编 ~/.xmips/sysfun/*.scp 生成 .co）
```

- `config.ini`、`run.list`、`userfile/`、`sysfun/` 均自动定位到 **`~/.xmips/`**（可被环境变量 `XMIPS_HOME` 覆盖），与当前工作目录无关，因此可在任意目录直接运行；
- 用户程序 open 的文件为**真实宿主文件**（相对当前工作目录或绝对路径），可直接读写主机任意路径；
- 可执行文件安装在 `/usr/local/bin/xmips`，已入 `PATH`，任何目录执行 `xmips XXX.cupa` 即可。

运行后系统输出依次为：
1. 系统函数库的汇编源程序
2. 用户程序的源程序
3. 进程调度信息
4. 运行结果（数据段内容）

SUM 程序的正确运行结果：

```
process ID:50  memory ID:1
0   55
1   0
2   0
...
```

---

## 七、其他功能简介

### 7.1 设计原则（自 2026-09 起）

Xmips 的定位更偏向**工具脚本**，而非完整操作系统，因此工作重心逐步移动到**单进程模式**（`./xmips XXX.cupa`，或 run.list 仅一个文件；多用 run.list 多进程调度则保留以兼容教学）。

涉及**输入输出 / 文件 / 网络**这类需要系统与底层能力介入的功能，统一由 **dispatcher 直接接管、以无 `.scp` 的「原生系统调用」方式实现**：

- 对用户程序暴露的仍是统一的 `MOV #10 调用号; INT` 接口，底层是 `.scp` 系统函数还是 dispatcher 原生，对用户透明；
- `.scp` 仅保留给纯 CPU 计算类辅助函数（当前为 `INT 10` 冒泡排序）；
- 用户程序不应直接访问通道 / 保留输入寄存器（`#14`、`#17`、`#18`），一律走系统调用。

### 7.2 系统调用实现形态

| 调用号 | 功能 | 实现 |
|--------|------|------|
| 10 | 冒泡排序 | `.scp` 系统函数 |
| 13 | 输出字符串 | dispatcher 原生 |
| 14 | 输入到缓冲区 | dispatcher 原生 |
| 15~18 | 文件 / Socket | dispatcher 原生 |

### 7.3 历史

Xmips 的核心目标原本是实现一个虚拟的支持多道程序设计和可交互式的操作系统。基本思想是虚拟化：

1. 用一个线程模拟处理机
2. 用多线程或多进程程序模拟各个设备的并行工作

目前已部分实现了虚拟化的第一个方面：当所有进程只使用处理机资源时，可以实现并发执行（伪并行，通过分时轮换）。但涉及外设操作时，由于外设和处理机的并行是真正的并行，无法用单线程程序模拟。

待完善的方向：
- 改造为多线程程序
- 建立模拟的内存管理机构
- 建立模拟的文件系统
- 建立可交互的命令行环境（类似 shell）
- 扩展数据类型支持
- 减少对实现语言（现为 Go）的依赖，逐步用 ABC 汇编语言实现更多系统功能

---

## 附录 A：操作码和助记符对照表

| 操作符 | 功能 | 机器码 |
|--------|------|--------|
| MOV | 数据传送 | 900152 |
| LEA | 地址传送 | 902042 |
| ADD | 加 | 900272 |
| INC | 加 1 | 900961 |
| SUB | 减 | 901072 |
| NEG | 求相反数 | 901961 |
| CMP | 比较 | 900332 |
| MUL | 乘 | 902072 |
| DIV | 除 | 903072 |
| MOD | 取模 | 904072 |
| AND | 与 | 905072 |
| OR | 或 | 906072 |
| NOT | 非 | 902961 |
| XOR | 异或 | 907072 |
| SHL | 逻辑左移（按机器字长） | 908072 |
| SHR | 逻辑右移（按机器字长，无符号） | 909072 |
| ROL | 循环左移（按机器字长: 32/64 位） | 911072 |
| ROR | 循环右移（按机器字长: 32/64 位） | 912072 |
| JA | 大于跳转 | 900421 |
| JB | 小于跳转 | 900521 |
| JMP | 无条件跳转 | 900621 |
| JE | 等于跳转 | 900721 |
| JNE | 不等于跳转 | 900821 |
| PUSH | 压栈 | 902121 |
| POP | 出栈 | 902241 |
| PUSHA | 寄存器组压栈 | 902300 |
| POPA | 寄存器组出栈 | 902400 |
| CALL | 函数调用 | 904121 |
| RET | 函数调用返回 | 904200 |
| INT | 软中断 | 904000 |
| IRET | 中断返回 | 904300 |
| CLI | 关中断 | 960000 |
| STI | 开中断 | 960100 |
| SYSR | 读系统共享区 | 950000 |
| SYSW | 写系统共享区 | 950100 |
| WAKE | 唤醒进程 | 950200 |
| SET | 设置通信标识 | 950300 |
| END | 结束 | 900000 |
| HALT | 停机 | 800000 |

**伪指令**：

| 操作符 | 功能 | 机器码 |
|--------|------|--------|
| DIM | 变量定义 | 910002 |
| LOC | 同 DIM（已不用） | 910002 |
| PROC | 函数定义 | 920001 |
| : | 标号 | 920001 |
| ~ | 注释 | 3 |
| CMNT | 注释 | 3 |
| $ | 占位空语句 | 0 |

---

## 附录 B：系统运行状态信息表

状态信息格式：`{RES, Level, Message}`

Level 等级：0=主要运行状况, 1=次要运行状况, 2=警告, 3=错误

| RES | Level | Message |
|-----|-------|---------|
| 0 | 0 | SYSTEM: normal end |
| 10 | 3 | memory::read: memory overflow |
| 11 | 3 | memory::write: memory overflow |
| 12 | 3 | editor::editorData: illegal data |
| 13 | 3 | editor::editorCollection: illegal variable name |
| 14 | 3 | editor::editorCollection: illegal opcode |
| 15 | 3 | editor::editorCollection: too many content |
| 16 | 3 | stack::push: stack is full |
| 17 | 3 | stack::pop: stack is empty |
| 18 | 0 | memory::update: update memory size |
| 19 | 3 | editor::editorCollection: error editor num |
| 20 | 1 | interpreter::exer: interrupt occured |
| 21 | 3 | interpreter::exer: interpreter overflow |
| 22 | 1 | interpreter::exer: interpreter reaches cycleTimes |
| 23 | 0 | interpreter::exer: process execute completed |
| 24 | 3 | interpreter::exer: property set illegal |
| 51 | 3 | assembler::trans2: '@' use error |
| 52 | 3 | assembler::trans2: undeclared indentifier |
| 53 | 3 | assembler::trans2: '&' use error |
| 54 | 3 | assembler::trans2: register used in indirect addressing |
| 55 | 3 | assembler::trans2: '#' use error |
| 56 | 3 | assembler::trans2: other char occur before immediate num |
| 57 | 3 | assembler::trans2: '!' use error |
| 58 | 3 | assembler::trans2: over the max general register number(14) |
| 60 | 3 | queue::enQueue: queue is full |
| 61 | 2 | queue::deQueue: queue is empty |
| 80 | 3 | storage::getFile: visit way error |
| 81 | 3 | storage::getFile: file open failed |
| 82 | 3 | storage::releaseFile: file pointer is null |
| 90 | 1 | sysList::copy: a system function is copied |
| 130 | 2 | pcbList::enQueue: list is empty |
| 131 | 3 | pcbList::get: list is empty |
| 132 | 3 | pcbList::get: pcb not found |
| 170 | 0 | dispatcher::swap2: process swaps to run |
| 171 | 0 | dispatcher::swap2: process swaps from run to ready |
| 172 | 0 | dispatcher::swap2: process swaps to wait |
| 173 | 1 | dispatcher::swap2: a system call happened |
| 174 | 1 | dispatcher::swap2: a system function completed |
| 175 | 0 | dispatcher::swap2: process swaps from wait to ready |
| 176 | 0 | dispatcher::swap2: a system function swaps to ready |
| 177 | 0 | dispatcher::swap2: a process finished |
| 178 | 0 | dispatcher::swap2: a process failed |
| 179 | 1 | dispatcher::swap2: swap is completed |
| 180 | 3 | dispatcher::swap2: access violation to system share region |
| 181 | 1 | dispatcher::swap2: read system share |
| 182 | 1 | dispatcher::swap2: write system share |
| 183 | 3 | dispatcher::pcbManagement: no pcb can use now |
| 184 | 3 | dispatcher::loader: pcb dispatch failed |
| 185 | 0 | dispatcher::swap2: a process suspended |
| 186 | 0 | dispatcher::swap2: a process woken |
| 187 | 3 | dispatcher::swap2: access violation to wake up a process |
| 188 | 1 | dispatcher::swap2: insert back to the head of ready queue |
| 189 | 1 | dispatcher::swap2: interruption ban checked |
| 190 | 3 | dispatcher::swap2: over the range of wait queue numbers |
| 191 | 3 | editor::editorFromFile: initial value more than declared |
| 192 | 3 | editor::editorFromFile: illegal declaration |
| 193 | 3 | editor::editorFromFile: tag repeated |
| 194 | 3 | editor::editorFromFile: tag not find |
| 195 | 3 | editor::ASM: open file failed |
| 196 | 3 | editor::editorFromFile: indentifier repeated |
| 197 | 3 | editor::editorFromFile: immediate num cant be 1st operand in double operands instruction |
