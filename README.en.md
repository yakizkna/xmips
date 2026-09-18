> **Language / 语言**: [中文](README.md) · [English](README.en.md)

# A Software-Implemented Instruction Set and Its Assembly Language — the Xmips Instruction Set and the cupa Assembly Language

## Table of Contents

- [1. Overview](#1-overview)
- [2. The Xmips Instruction Set](#2-the-xmips-instruction-set)
- [3. The cupa Assembly Language](#3-the-cupa-assembly-language)
- [4. The Assembly Process](#4-the-assembly-process)
- [5. The Execution Process](#5-the-execution-process)
- [6. Xmips Operation Examples](#6-xmips-operation-examples)
- [7. Introduction to Other Features](#7-introduction-to-other-features)
- [Appendix A: Opcode and Mnemonic Reference Table](#appendix-a-opcode-and-mnemonic-reference-table)
- [Appendix B: System Running Status Information Table](#appendix-b-system-running-status-information-table)

---

## 1. Overview

### 1.1 About Xmips

Xmips is a program designed to simulate multiprocess scheduling and resource allocation (**it was originally implemented in C++, and has now been completely rewritten in Go**). The Xmips system currently mainly consists of an assembler part, an instruction interpretation and execution part, and a process scheduling part.

Xmips is a single-threaded console program. To use it to simulate multiprocess concurrency, a set of simulated instructions and a programming language must be designed for it. To achieve this goal, mimicking the form of X86 assembly language, a set of assembly language instructions was designed for the Xmips system, named Stimulated Assembly Language By C++ (**ABC was its historical name, i.e., the name derived from the early C++ implementation**; it has since been renamed **cupa**, with the recommended file extension `.cupa`), and a simple assembler was written for it to translate assembly source programs into machine code that can be executed by the system.

User programs running on the Xmips system are all written in this assembly language, while Xmips's system programs and system function library are written in the cupa assembly language, and the rest is written in Go (previously C++). For this part, it can be understood as the "hardware functionality" simulated by the Xmips system.

For simplicity, all data units in Xmips are defined as integers (int), i.e., every field of an instruction is an integer value, simulated memory is addressed in units of int, and the size of all registers is also int. So currently the assembly language of the Xmips system is an untyped language that does not support character types or floating-point types; all operations target integer values.

### 1.2 The Earliest Xmips Instruction Set

In the earliest Xmips system, there were no registers; instruction execution actually modified the data segment directly according to the content of the code segment. The reason for not having registers was the belief that, under purely software simulation conditions, there is no difference in access speed between registers and main memory. However, an area was added to main memory for storing information such as the opcode and operand addresses. In later designs, that area was removed and its functionality was replaced by a register set provided by the system.

When design first began, the cupa assembly language and machine instructions corresponded strictly in format:

```
操作码  目的操作数地址  源操作数地址
```

For single-operand instructions and no-operand instructions, the operand address defaults to 0.

This approach had a single addressing mode that only supported direct addressing. To support immediate addressing, a dedicated set of instructions for immediates had to be added, with opcode mnemonics of the form `IMXXX`, such as `IMADD SUM 9`, which means adding the immediate 9 to the memory unit corresponding to symbolic address SUM, i.e., `SUM ← (SUM)+9`.

Because some programs need indirect addressing, the `@` symbol denotes that the actual address is the value in the address cell after `@`, e.g., `MOV SUM @I` means `SUM ← ([I])`.

### 1.3 The Improved Xmips Instruction Set

After studying the instruction set of computer organization, some improvements were made to the original Xmips instruction set and the cupa assembly language.

1. **New registers**: two register groups, MRgst and GM, were designed, along with some registers for system use, such as the instruction counter PC and the status register flag.
2. **Instruction format separated from the cupa assembly language format, adding an addressing mode field**: the cupa assembly language format is the same as before, but the operators `&` and `@` for offset addressing and indirect addressing were added, along with register identifiers `#0~#14`. The instruction format consists of four fixed-length fields (each field is an integer value):

```
操作码  寻址方式  目的操作数形式地址  源操作数形式地址
```

The supported addressing modes include immediate addressing, direct addressing, indirect addressing, register addressing, register indirect addressing, and indexed addressing, etc. For the two-operand case, the addressing mode field is formed by concatenating the destination operand's addressing mode and the source operand's addressing mode.

3. **The process of interpreting and executing instructions was refined**: for instructions of the same class of data path, the fetch and store operations are completed uniformly. The general form of the operation is:
   - `<1>` Fetch the instruction (fetch the opcode, fetch the addressing mode, fetch the destination operand's formal address, fetch the source operand's formal address, PC+1)
   - `<2>` Compute the operand's effective address
   - `<3>` Fetch the operand (it may be in a register, in main memory, an immediate, or the operand may be an address)
   - `<4>` Execute the arithmetic operation
   - `<5>` Store the operand
   - `<6>` Fetch the next instruction

---

## 2. The Xmips Instruction Set

In the Xmips system, the basic value unit is the integer (int), memory cells are addressed in units of int, and the length of all instructions and registers (register groups) can only be an integer multiple of int.

### 2.1 Registers in Xmips

The registers in Xmips fall into two categories:

**Internal registers** (for system use only, not accessible by users):
- Instruction counter PC
- Status register flag
- Registers numbered #15 and #16 in the register group GM (data registers)
- All registers in the register group MRgst

**General-purpose registers** (usable by users):
- 15 integer general-purpose registers, numbered #0~#14, all located in the register group GM
- Among them, registers #1~#9 can be used for indexed addressing

| Register Group | Number | Function | Accessible By |
|---------|------|---------|---------|
| PC | 0 | Instruction counter | System |
| flag | 0 | Status register | System |
| MRgst | 0 | Stores the opcode | System |
| MRgst | 1 | Stores the addressing mode (for two operands, decomposed into destination/source addressing mode) | System |
| MRgst | 2 | Stores the destination operand's formal address and effective address | System |
| MRgst | 3 | Stores the location information of the destination operand (0=main memory, 1=register, 2=immediate) | System |
| MRgst | 4 | Stores the source operand's formal address and effective address | System |
| MRgst | 5 | Stores the location information of the source operand | System |
| MRgst | 6~14 | Undefined, reserved for extension | System |
| GM | 0 | General-purpose register | System/User |
| GM | 1~9 | General-purpose registers, usable as index registers for offset addressing (format `&R`) | System/User |
| GM | 10 | General-purpose register, stores the interrupt number and interrupt return information | System/User |
| GM | 11~12 | General-purpose registers, used for passing parameters during system function calls | System/User |
| GM | 13 | **Output register**: writing to it outputs a character to stdout | System/User |
| GM | 14 | Used only by INT 14 (input) where the dispatcher natively reads stdin; **a user program reading #14 raises error RES=198 illegal access** | System |
| GM | 15 | Data register, stores the destination operand's value | System |
| GM | 16 | Data register, stores the source operand's value | System |
| GM | 17~19 | Undefined, reserved for extension | System |

### 2.2 The Xmips Instruction Format

An Xmips instruction has 4 fields in total, and each field is the size of an integer value:

```
操作码  寻址方式  目的操作数形式地址  源操作数形式地址
```

**1) The opcode field**

The opcode can be decomposed into three parts:
- **Operation identifier (op)**: the other digits of the opcode, corresponding one-to-one with the operation
- **Data operation type identifier (d)**: the tens digit of the opcode, indicating how the operands are fetched and stored
- **Operand count identifier (t)**: the ones digit of the opcode (two operands t=2, one operand t=1, no operand t=0)

For example, the opcode of `ADD #2 SUM` is `900272`:
- op=9002 (addition operation)
- d=7 (fetch the values of the source and destination operands, store the result into the destination operand)
- t=2 (two operands)

**2) The addressing mode field**

An integer value is used to represent the addressing mode. For two operands, the addressing mode = addressing_d + addressing_s × 1000.

For a single operand, the addressing mode is a 3-digit decimal integer ABC:
- R = A = addressing / 100 (the hundreds digit represents the register number for offset addressing, range 1~9)
- type = BC = addressing mod 100 (the tens and ones digits represent the addressing mode type)

type corresponds to a binary number (S0S1S2S3):
- S0=1 indicates offset addressing
- S1=1 indicates indirect addressing
- S2S3=00 indicates the operand is in main memory
- S2S3=10 indicates the operand is in a register
- S2S3=11 indicates the operand is an immediate

**3) The operand formal address field**

- If the operand is in the data segment (main memory) with in-segment offset M, then oa = M
- If the operand is in general-purpose register R, then oa = R
- If the operand is an immediate I, then oa = I (in this case the addressing mode field is 3)

### 2.3 Xmips Addressing Modes

**1) Supported addressing modes**:
- Register addressing
- Register indirect addressing
- Direct addressing
- Indirect addressing
- Base or indexed addressing (effective address = some general-purpose register value + a constant in the instruction)
- Immediate addressing

**2) Computation of the effective address**

cupa assembly language address form: `[&R][@]K`

| addressing | S0S1S2S3 | assembly address form | effective address computation | operand position |
|-----------|----------|------------|------------|----------|
| 0 | 0000 | SUM, !0 | EA=D | main memory |
| 2 | 0010 | #2 | EA=#D | register |
| 3 | 0011 | 6 | does not exist | immediate |
| 4 | 0100 | @ST, @!10 | EA=(D) | main memory |
| 6 | 0110 | @#2 | EA=(#D) | main memory |
| 8+100R | 1000 | &3SUM, &9!21 | EA=(#R)+D | main memory |
| 12+100R | 1100 | &5@ST | EA=(#R)+(D) | main memory |
| 14+100R | 1110 | &6@#5 | EA=(#R)+(#D) | main memory |

### 2.4 Address Space Description

The current Xmips lacks a memory management mechanism and an address translation mechanism; in effect, each process's code segment, data segment, and stack segment are physically independent memories:

- The program's logical address space is the physical address space of the memory
- The code segment, data segment, and stack segment correspond to three memories respectively, and each segment starts at address 0

In the source code, `MCode` denotes the code segment, `MData` the data segment, and `S` the stack segment.

---

## 3. The cupa Assembly Language

### 3.1 A Simple Example — SUM

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

The character-code file generated after assembly:

```
900152 3002 1 1
900272 2000 0 1
900961 2 1 0
900332 3002 1 11
900521 3 4 0
900000
1 0
```

### 3.2 Addressing Modes

| Addressing Mode | Format | Description |
|---------|------|------|
| Register addressing | `#R` | The content of register #R is the operand |
| Register indirect addressing | `@#R` | The content of register #R is the operand's offset address |
| Direct addressing | `S` or `!K` | S is a variable name; K is an integer data-segment address |
| Indirect addressing | `@S` or `@!K` | The formal address is the address of the operand's address |
| Offset addressing | `&RX` | Effective address = formal address of (#R) + X |
| Immediate addressing | `n` | The formal address field is the operand |

Combination restrictions:
- Immediate addressing cannot contain other components; for example, `@8` is illegal
- The X part of offset addressing cannot be a register addressing; for example, `&5#3` is illegal

### 3.3 Machine Instruction Statements and Pseudo Instructions

**1) Variable definition statements**

```
DIM S [Dn] 初值1, 初值2, ..., 初值K;
```

Example:
```
DIM SET D5 1,2,3;    ~ 定义大小为5的数组，前三个元素赋初值
DIM MAX 100           ~ 定义变量 MAX，初值为 100
```

**2) Label definition statements**

```
: L
```

**3) Machine instruction statements**

| Type | Format | Function |
|------|------|------|
| Data transfer | `MOV OPD OPS` | OPD ← (OPS) |
| Address transfer | `LEA OPD OPS` | OPD ← OPS |
| Addition | `ADD OPD OPS` | OPD ← (OPD) + (OPS) |
| Increment by 1 | `INC OPD` | OPD ← (OPD) + 1 |
| Subtraction | `SUB OPD OPS` | OPD ← (OPD) - (OPS) |
| Negation | `NEG OPD` | OPD ← -(OPD) |
| Comparison | `CMP OPD OPS` | (OPD) - (OPS), the result is saved into flag |
| Multiplication | `MUL OPD OPS` | OPD ← (OPD) * (OPS) |
| Division | `DIV OPD OPS` | OPD ← (OPD) / (OPS), #0 ← (OPD) % (OPS) |
| Modulo | `MOD OPD OPS` | OPD ← (OPD) % (OPS) |
| AND | `AND OPD OPS` | OPD ← (OPD) & (OPS) |
| OR | `OR OPD OPS` | OPD ← (OPD) \| (OPS) |
| NOT | `NOT OPD` | OPD ← ~(OPD) |
| XOR | `XOR OPD OPS` | OPD ← (OPD) ^ (OPS) |
| Jump if greater | `JA L` | if flag > 0 then PC=L |
| Jump if less | `JB L` | if flag < 0 then PC=L |
| Unconditional jump | `JMP L` | PC=L |
| Jump if equal | `JE L` | if flag == 0 then PC=L |
| Jump if not equal | `JNE L` | if flag != 0 then PC=L |
| Push | `PUSH OPD` | push (OPD) |
| Pop | `POP OPD` | OPD ← pop result |
| Push register group | `PUSHA` | push #0~#14 |
| Pop register group | `POPA` | pop #0~#14 |
| Function call | `CALL FUNCTION` | call function FUNCTION |
| Function return | `RET` | return to the calling function |
| Software interrupt | `INT` | the interrupt number is in #10 |
| Interrupt return | `IRET` | the return code is in #10 |
| Disable interrupts | `CLI` | disable clock-interrupt preemption |
| Enable interrupts | `STI` | enable clock-interrupt preemption |
| Read system shared region | `SYSR` | #11=starting address, #12=length (system functions only) |
| Write system shared region | `SYSW` | #11=starting address, #12=length (system functions only) |
| Wake up a process | `WAKE` | #11=communication identifier, #12=waiting-queue number (system functions only) |
| Set communication identifier | `SET` | #11=communication type, #12=communication identifier value (system functions only) |
| End | `END` | end of assembly |
| Halt | `HALT` | end of program |

**4) Pseudo instructions**

| Pseudo Instruction | Function |
|--------|------|
| `~ comment` or `CMNT comment` | comment (no separators allowed in between) |
| `PROC FUNCTION` | function definition (starts from the next line, ends at RET) |
| `: L` | label definition |
| `DIM S ...` | variable definition |
| `$` | placeholder empty statement |

> Note: operators support all-uppercase or all-lowercase, but mixing case is illegal.

### 3.4 Introduction to System Function Calls

The Xmips system provides a set of system functions that users invoke through the software-interrupt statement `INT`. Before the call, the call number is transferred into register #10, and the function arguments are passed to the corresponding registers.

Example program BUBBLE_INT.cupa (sorts 10 numbers through system call 10):

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

Running result (stored in the process's data segment):

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

### 3.5 Function Calls (CALL/RET)

Xmips supports subroutine calls through `CALL`/`RET` (no stack frames, no local variables; arguments are passed via register conventions, such as `#0`=array start address, `#1`=length). `CALL label` pushes the return address PC and the flag onto the stack and then jumps; `RET` pops the stack and returns.

When writing a program containing subroutines, note:

- The execution entry is at the first address of the code region (0); **the main program must be placed at the very beginning of the file** (after DIM);
- The assembler terminates at the first `END`, so **there can be only one `END` in the entire file** (placed at the end); the main program should instead end with `HALT` (equivalent to a normal termination without ending assembly);
- Labels must be defined with a `: name` prefix; `CALL` may forward-reference a label defined later.

A complete example is in [`file/BUBBLE_CALL.cupa`](dist/userfile/BUBBLE_CALL.cupa): it uses two user subroutines to perform a bubble sort and outputs the result —

```text
51 40 22 17 16 13 5 4 3 2
```

This is consistent with the result of `BUBBLE_INT.cupa` (system call 10) above. More detailed constraints can be found in section 5.1/5.2 of [`USAGE.md`](doc/USAGE.en.md).

### System Call Reference

| Call Number | Function | Arguments | Description |
|--------|------|------|------|
| 10 | Bubble sort | #11=array start address, #12=length | stores the sorted result back into the data segment |
| 13 | Output a string | #11=string start address | **dispatcher native**: reads the data segment from address #11 until it reads 0, outputting characters one by one to stdout |
| 14 | Input into a buffer | #11=buffer start address, #12=buffer size | **dispatcher native**: reads characters one by one from stdin into the buffer (no longer via `#14`), ending at EOF(0) or a full buffer; returns #13=0 for EOF, #13=1 for a full buffer (may `INT 14` again to continue reading) |
| 15 | open file/Socket | #11=pending string address (0-terminated), #12=mode | `<0` error code; ≥0 returns fd (file descriptor) |
| 16 | read data | #11=buffer address, #12=maximum length, #9=fd | 0=no data yet/EOF; N=number of bytes actually read |
| 17 | write data | #11=buffer address, #12=write length, #9=fd | N=number of bytes actually written |
| 18 | close | #9=fd | 0 success; <0 error code |

System calls 15~18 (files and network Sockets) are **natively executed by the dispatcher** (no `.scp`, no PCB used), and the return value is uniformly written back to `#9`. For communication details, see "4. Files and Networking" below.

Writing to #13 (the output register) outputs the corresponding character to stdout without a system call. INT 13 is used to output a string in bulk (a character sequence in the data segment terminated by 0).

Reading #14 (the input register): **user programs or other system functions reading #14 trigger a system error (RES=198, Error)**. User programs can only obtain input from stdin indirectly through `INT 14` (dispatcher native).

INT 13 output example:

```
~ 数据段 MSG 起始地址为 0（"hello\0"）
MOV #11 0
MOV #10 13
INT
```

INT 14 input example:

```
~ 数据段 NAME 起始地址为 20，大小 40
MOV #11 20
MOV #12 40
MOV #10 14
INT
MOV #0 #13      ~ #13=1 表示缓冲区满，可再次 INT 14 继续读
```

#### Files and Real Network Sockets (INT 15~18)

The prefix of the path string of `open` determines the object (`#11` is the address of a 0-terminated string in the data segment):

| Path Prefix | Meaning |
|---------|------|
| `sock:host:port` | TCP client connection; **port 443 automatically enables TLS** |
| `tlssock:host:port` | TCP client connection with **forced TLS** (`ServerName`=host; the handshake is subject to `sockTimeout`) |
| Other | a file, mapped to a **real host file**: `#11` is the file path (relative to the current working directory or an absolute path), read/written directly; `sock:`/`tlssock:` are TCP clients (`sock:443` or `tlssock:` automatically enable TLS) |

`#12` modes: 0=read-only, 1=write (truncate), 2=read/write append (valid for files). The returned fd is stored in `#9` (fd starts from 3). To read/write network/file byte streams: `INT 16` reads `#12` bytes from `#9` into the data segment pointed to by `#11`; `INT 17` writes `#12` bytes at `#11` to `#9`. Cupa can thereby implement real **HTTPS fetching + file writing** (see the example `userfile/NETTOUR.cupa`).

Read/write blocking semantics are determined by the **single-process mode** (running `./xmips xxx.cupa` directly, or when `run.list` contains exactly 1 file):

| Mode | socket `read` (INT 16) |
|------|------------------------|
| Single-process | blocks waiting for data, protected by the `sockTimeout` (default 2000ms) timeout; **on timeout, returns 0** and never hangs forever |
| Multiprocess | non-blocking (`deadline=now`): if data is present, returns N immediately; **if no data, returns 0 and yields the time slice** (the dispatcher puts the process at the tail of the ready queue to avoid busy spinning) |

`socket write` and connect/handshake are subject to `sockTimeout`. Error codes (returned as negative values in `#9`): `-1` open failed/out of range, `-2` illegal fd, `-3` connection failed, `-4` write buffer full, `-5` socket closed/read-write error.

---

## 4. The Assembly Process

### 4.1 From Source Program to Machine Instruction

The assembler ASM translates string-form assembly statements into numeric machine code. Four tables are involved in the assembly process:

- **Opcode Table**: records the machine code corresponding to each assembly operator
- **Variable Table**: registers the variable names and data-segment start addresses defined by DIM statements
- **Tag Table**: registers the labels defined by label statements and their code-segment addresses
- **Jump Table**: registers the labels referenced by jump statements and the jump statements' addresses

ASM workflow:

1. Look up the Opcode Table and translate operation mnemonics into machine code
2. If it is a variable definition statement, add the variable to the Variable Table
3. If it is a label statement, register the label into the Tag Table
4. If it is a jump instruction, register the label into the Jump Table
5. Analyze the operands and compute the addressing mode and formal address
6. On encountering the END statement, assembly terminates
7. After assembly, correct the destination operands of jump instructions according to the Tag Table and Jump Table

### 4.2 The Assembly Process of the SUM Program

The machine code generated by assembling each line (after label correction and data information):

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

## 5. The Execution Process

### 5.1 The Execution Mechanism

The execution mechanism of Xmips simulates a computer's processor, including:
1. Register (groups): PC, flag, MRgst, GM
2. The decoding mechanism: analyzes the opcode and selects the data path and arithmetic operation
3. The effective address computation mechanism
4. The memory access mechanism: fetches data from main memory into the data registers and returns the computation results to main memory
5. The arithmetic mechanism: performs arithmetic on the contents of the data registers

The execution process of one instruction:
1. Fetch the instruction, PC+1
2. Compute the operand's effective address according to the addressing mode and formal address
3. Analyze the opcode and select the data path
4. Transfer the value to be operated on into the data register according to the data path
5. Execute the operation
6. According to the data path, select whether to write the computation result back

### 5.2 The Data Path

An integer dataLS is used to represent the instruction's data path (the tens digit of the opcode), and V0V1V2 is its binary value:

| dataLS | V0V1V2 | Data Path | Example Operator | Opcode |
|--------|--------|---------|----------|--------|
| 7 | 111 | D ← (D) + (S) | ADD | 900272 |
| 6 | 110 | D ← (D) + 1 | INC | 900961 |
| 5 | 101 | D ← (S) | MOV | 900152 |
| 4 | 100 | D ← S | LEA / POP | 902042 / 902241 |
| 3 | 011 | (D) - (S) | CMP | 900332 |
| 2 | 010 | (D) / stack ← (D) | JMP / PUSH | 900621 / 902121 |
| 0 | 000 | no operand | others | XXXX00 |

- V0=1: the computation result is saved to the destination operand's address
- V1=1: fetch the value of the destination operand into the data register
- V2=1: fetch the value of the source operand into the data register

### 5.3 The Execution Process of the SUM Program

SUM's character-code file:

```
900152 3002 1 1    900272 2000 0 1    900961 2 1 0
900332 3002 1 11   900521 3 4 0       900000
1 0
```

Execution flow:

1. **MOV #1 1**: PC=0→4, (#1)=1
2. **ADD SUM #1**: PC=4→8, MData[0]=1
3. **INC #1**: PC=8→12, (#1)=2
4. **CMP #1 11**: PC=12→16, flag=-1
5. **JB L1**: PC=16→4 (jump, because flag<0)
6. Repeat steps 2~5 until (#1)=11, at which point MData[0]=55
7. **END**: the program ends; the final result 55 is stored in the SUM unit

---

## 6. Xmips Operation Examples

### 6.1 Path Settings and Parameter Configuration

**Repository directory structure**:

```
xmips/
├── src/                    # Go 源代码（main.go、dispatcher.go、interpreter.go ...）
├── tool.sh                 # 工具脚本：./tool.sh build / clean
├── dist/                   # 运行目录（build 时同步到 ~/.xmips/，可执行文件不保留在此）
│   ├── config.ini          # 配置文件（缺省时用内置默认值）
│   ├── run.list            # 运行列表文件
│   ├── userfile/           # 用户程序目录
│   └── sysfun/             # 系统函数库目录（用户不可修改）
├── doc/
│   ├── USAGE.md            # 运行与参数说明
│   └── IO_SYSCALL_SPEC.md  # 文件/Socket 系统调用规格
└── tools/                  # 辅助工具（如 md5gen）
```

**System functions**: the `.scp` source files under `dist/sysfun/` are uniformly named `INT_xx.scp`, where `xx` is the call number (corresponding to the value in #10). **Only `INT 10` (bubble sort) is implemented with a `.scp` system function**; `INT 13/14` (output/input) and `INT 15~18` (file/Socket) are all **natively executed by the dispatcher** (no `.scp`, no PCB):

| Call Number | Implementation | Function |
|--------|---------|------|
| 10 | `.scp` system function | Bubble sort |
| 13 | dispatcher native | reads a string (0-terminated) starting from address #11 and outputs each character to stdout |
| 14 | dispatcher native | reads characters one by one from stdin into the #11 buffer (of #12 size); returns #13=0 EOF, #13=1 buffer full |
| 15~18 | dispatcher native | file/Socket open / read / write / close |

**run.list file format**:

```
文件名1
文件名2
...
end
```

**config.ini configuration parameters**:

| Parameter | Value | Description |
|------|---|------|
| delayMode | 0 | disables memory-access delay simulation |
| delayMode | 1 | enables memory-access delay simulation |
| displayMode | 0 | does not display the source program during assembly |
| displayMode | 1 | displays the source program during assembly |
| displayMode | 2 | during running, only displays standard output; system errors show error messages; everything else is suppressed |
| reportLevel | 0 | displays warnings and error messages |
| reportLevel | 1 | displays main running status + warnings + errors |
| reportLevel | 2 | displays all messages |
| updateSysfun | — | superseded by the standalone command `./xmips update` (rebuilds the system function library, generating .co) |
| codeSize | 800 | length of the user process's code region (in int integers) |
| dataSize | 200 | length of the user process's data region (in int integers) |
| stackSize | 50 | length of the user process's stack region (in int integers) |
| sysfunCodeSize | 800 | length of the system function's code region (in int integers) |
| sysfunDataSize | 10 | length of the system function's data region (in int integers; at runtime it points to the caller's data segment) |
| sysfunStackSize | 40 | length of the system function's stack region (in int integers) |
| cycleTimes | 30 | time-slice size: the maximum number of instructions executed per time slice |
| pcbNum | 40 | PCB count: the system's maximum number of concurrent processes |
| bitMode | 64 | machine word length: 64=64-bit integer (default, overflow does not wrap); 32=32-bit integer (arithmetic overflow wraps by 2^32, and SHL/SHR/ROL/ROR operate on 32 bits) |

config.ini format requirements: each line is `key=value` (no spaces), ending with `end`; lines beginning with `;` are comments, and a `;` inline comment may follow the parameter value.

### 6.2 Editing the Assembly Source Program

User source programs must be created under the `dist/userfile/` directory and end with an `END` statement.

### 6.3 Running xmips

Compile and run procedures (see [`USAGE.md`](doc/USAGE.en.md) for details):

```
./tool.sh build       # 1. 编译；把 xmips 安装到 /usr/local/bin/xmips，dist/* 同步到 ~/.xmips/，重建系统函数库
./tool.sh clean       # 2. 卸载：rm -rf ~/.xmips；rm /usr/local/bin/xmips
```

Running (**can be invoked from any directory**; the binary is at `/usr/local/bin/xmips` and data is at `~/.xmips/`):

```
xmips                      # A. 无参数 → 从 ~/.xmips/run.list 读取要运行的程序
xmips XXX.cupa             # B. 指定程序：优先当前目录，其次 ~/.xmips/userfile/ 目录
xmips /绝对/路径/XXX.cupa   # C. 指定程序的绝对/相对路径
xmips update               # D. 重建系统函数库（重汇编 ~/.xmips/sysfun/*.scp 生成 .co）
```

- `config.ini`, `run.list`, `userfile/`, `sysfun/` are all automatically located under **`~/.xmips/`** (overridable by the environment variable `XMIPS_HOME`), regardless of the current working directory, so it can be run from any directory;
- Files opened by a user program are **real host files** (relative to the current working directory or absolute paths), and any host path can be read/written directly;
- The executable is installed at `/usr/local/bin/xmips` and is already on `PATH`, so running `xmips XXX.cupa` from any directory suffices.

After running, the system output appears in the following order:
1. The assembly source programs of the system function library
2. The source program of the user program
3. Process scheduling information
4. The running result (data segment contents)

Correct running result of the SUM program:

```
process ID:50  memory ID:1
0   55
1   0
2   0
...
```

---

## 7. Introduction to Other Features

### 7.1 Design Principles (since 2026-09)

Xmips is positioned more as a **tool script** than a full operating system, so the focus of the work has gradually moved to **single-process mode** (`./xmips XXX.cupa`, or a run.list with only one file; the multi-process scheduling via a multi-file run.list is retained for teaching compatibility).

For features that require system and low-level capabilities to intervene — involving **input/output, files, and networking** — they are uniformly **taken over directly by the dispatcher and implemented as "native system calls" without `.scp`**:

- To user programs, the exposed interface remains the uniform `MOV #10 callNumber; INT`; whether the underlying implementation is a `.scp` system function or dispatcher native is transparent to users;
- `.scp` is reserved only for pure CPU-computational helper functions (currently `INT 10` bubble sort);
- User programs should not directly access channels / reserved input registers (`#14`, `#17`, `#18`); they should always go through system calls.

### 7.2 System Call Implementation Forms

| Call Number | Function | Implementation |
|--------|------|------|
| 10 | Bubble sort | `.scp` system function |
| 13 | Output a string | dispatcher native |
| 14 | Input into a buffer | dispatcher native |
| 15~18 | File / Socket | dispatcher native |

### 7.3 History

The core goal of Xmips was originally to implement a virtual operating system supporting multiprogramming and interaction. The basic idea is virtualization:

1. Use one thread to simulate the processor
2. Use multithreaded or multiprocess programs to simulate the parallel operation of the various devices

Currently the first aspect of virtualization is partially implemented: when all processes only use the processor resource, concurrent execution can be achieved (pseudo-parallelism, through time-slice rotation). However, when peripheral operations are involved, because the parallelism between peripherals and the processor is genuine parallelism, it cannot be simulated with a single-threaded program.

Directions to be improved:
- Rework it into a multithreaded program
- Build a simulated memory management mechanism
- Build a simulated file system
- Build an interactive command-line environment (similar to a shell)
- Extend data type support
- Reduce the reliance on the implementation language (currently Go) and gradually implement more system functions in the cupa assembly language

---

## Appendix A: Opcode and Mnemonic Reference Table

| Operator | Function | Machine Code |
|--------|------|--------|
| MOV | Data transfer | 900152 |
| LEA | Address transfer | 902042 |
| ADD | Add | 900272 |
| INC | Increment by 1 | 900961 |
| SUB | Subtract | 901072 |
| NEG | Negation | 901961 |
| CMP | Compare | 900332 |
| MUL | Multiply | 902072 |
| DIV | Divide | 903072 |
| MOD | Modulo | 904072 |
| AND | Bitwise AND | 905072 |
| OR | Bitwise OR | 906072 |
| NOT | Bitwise NOT | 902961 |
| XOR | Bitwise XOR | 907072 |
| SHL | Logical shift left (by machine word length) | 908072 |
| SHR | Logical shift right (by machine word length, unsigned) | 909072 |
| ROL | Rotate left (by machine word length: 32/64 bits) | 911072 |
| ROR | Rotate right (by machine word length: 32/64 bits) | 912072 |
| JA | Jump if greater | 900421 |
| JB | Jump if less | 900521 |
| JMP | Unconditional jump | 900621 |
| JE | Jump if equal | 900721 |
| JNE | Jump if not equal | 900821 |
| PUSH | Push onto stack | 902121 |
| POP | Pop from stack | 902241 |
| PUSHA | Push register group | 902300 |
| POPA | Pop register group | 902400 |
| CALL | Function call | 904121 |
| RET | Function call return | 904200 |
| INT | Software interrupt | 904000 |
| IRET | Interrupt return | 904300 |
| CLI | Disable interrupts | 960000 |
| STI | Enable interrupts | 960100 |
| SYSR | Read system shared region | 950000 |
| SYSW | Write system shared region | 950100 |
| WAKE | Wake up a process | 950200 |
| SET | Set communication identifier | 950300 |
| END | End | 900000 |
| HALT | Halt | 800000 |

**Pseudo instructions**:

| Operator | Function | Machine Code |
|--------|------|--------|
| DIM | Variable definition | 910002 |
| LOC | Same as DIM (no longer used) | 910002 |
| PROC | Function definition | 920001 |
| : | Label | 920001 |
| ~ | Comment | 3 |
| CMNT | Comment | 3 |
| $ | Placeholder empty statement | 0 |

---

## Appendix B: System Running Status Information Table

Status information format: `{RES, Level, Message}`

Level grades: 0=main running status, 1=secondary running status, 2=warning, 3=error

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