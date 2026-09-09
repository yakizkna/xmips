package main

import "fmt"

// I/O 寄存器编号（占用 GM 中的预留寄存器，通过系统调用访问）
const outputReg = 13 // #13: 输出寄存器
const inputReg = 14  // #14: 输入寄存器

// Interpreter 解释执行器，对应 C++ 的 interpreter 类
type Interpreter struct {
	ID         int
	cycleTimes int
	MRgst      *Memory // 内部寄存器组，用于寄存器编码和临时存储
	GM         *Memory // 通用寄存器组（0~16，其中15=目的操作数，16=源操作数）
	GMNumber   int
	flag       int // 状态标志寄存器
	PC         int // 指令计数器
}

func newInterpreter(key, cc, regNum, GMNum int) *Interpreter {
	return &Interpreter{
		ID:         key,
		cycleTimes: cc,
		MRgst:      newMemory(10*key, regNum),
		GM:         newMemory(10*key+1, GMNum),
		GMNumber:   GMNum,
	}
}

func defaultInterpreter() *Interpreter {
	return newInterpreter(101, 30, 10, 20)
}

// load 从 M2[is] 加载到 M1[id]
func (im *Interpreter) load(m1 *Memory, id int, m2 *Memory, is int) {
	procDelay(delayMode, 100)
	m1.write(id, m2.read(is))
}

// store 从 M2[is] 存储到 M1[id]
func (im *Interpreter) store(m1 *Memory, id int, m2 *Memory, is int) {
	procDelay(delayMode, 100)
	m1.write(id, m2.read(is))
}

func (im *Interpreter) push(vs int, s *Stack) {
	procDelay(delayMode, 50)
	s.push(vs)
}

func (im *Interpreter) pop(s *Stack) int {
	procDelay(delayMode, 50)
	return s.pop()
}

// isInt14 当前进程是否为 INT 14 输入系统函数
// 系统函数进程 ID = sysFunTable 索引 = 系统调用号-10（INT 14 → ID 4），且 callerID 指向调用者
func (im *Interpreter) isInt14(proc *Process) bool {
	return proc.callerID != 0 && proc.ID == inputReg-10
}

// readGM 读取通用寄存器
// 输入寄存器 #14 只允许 INT 14 系统函数读取（从 stdin 读一个字符，EOF 返回 0）；
// 其他进程读 #14 属于非法访问，报错并返回 0
func (im *Interpreter) readGM(n int, proc *Process) int {
	if n == inputReg {
		if !im.isInt14(proc) {
			RES = 198
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d, illegal access to input register #14 ", im.ID, proc.ID)
				systemChecker.check(RES, sysLog[:])
			}
			return 0
		}
		b, err := stdinReader.ReadByte()
		if err != nil {
			return 0 // EOF
		}
		return int(b)
	}
	return truncWord(im.GM.read(n))
}

// effectAddressing 计算有效地址
// addcode: 寻址方式编码, fadd: 形式地址, proc: 进程
// RorM: 0=在内存中, 1=在寄存器中, 2=立即数
func (im *Interpreter) effectAddressing(addcode, fadd int, proc *Process, RorM *int) int {
	R := addcode / 100 // 偏移寻址时使用的通用寄存器编号
	typ := addcode % 100
	*RorM = 0

	var EA int
	switch typ {
	case 0:
		EA = fadd
	case 2:
		EA = fadd
		*RorM = 1
	case 3:
		EA = fadd
		*RorM = 2
	case 4:
		EA = proc.MData.read(fadd)
	case 6:
		EA = im.readGM(fadd, proc)
	case 8:
		EA = im.readGM(R, proc) + fadd
	case 12:
		EA = im.readGM(R, proc) + proc.MData.read(fadd)
	case 14:
		EA = im.readGM(R, proc) + im.readGM(fadd, proc)
	default:
		printf("error!")
		return -1
	}
	return EA
}

// exer 执行进程的指令，返回执行状态
// 0: 正常结束, 1: 时间片到, -1: 失败, 3: 系统函数返回(IRET),
// 5: SYSR, 6: SYSW, 7: WAKE, >=10: 系统调用号
func (im *Interpreter) exer(proc *Process) int {
	i := 0
	var ppk int

	if proc.exetime == 0 {
		im.PC = 0
		proc.exetime = 1
	} else {
		// 恢复现场
		im.flag = im.pop(proc.S)
		im.PC = im.pop(proc.S)
		for ppk = 14; ppk >= 0; ppk-- {
			im.GM.mem[ppk] = im.pop(proc.S)
		}
	}

	im.load(im.MRgst, 0, proc.MCode, im.PC)
	im.PC++
	proc.exetime++

	for i < im.cycleTimes && im.MRgst.read(0) != 900000 && im.MRgst.read(0) != 800000 && im.PC < proc.MCode.mSize {
		var addVisit_s, addVisit_d int
		var opType int
		var EA_s, EA_d int
		var RorM_s, RorM_d int
		var dataLS int

		opType = im.MRgst.read(0) % 10
		dataLS = (im.MRgst.read(0) % 100) / 10

		im.load(im.MRgst, 1, proc.MCode, im.PC)
		im.PC++
		im.load(im.MRgst, 2, proc.MCode, im.PC)
		im.PC++
		im.load(im.MRgst, 4, proc.MCode, im.PC)
		im.PC++

		addVisit_s = im.MRgst.read(1) % 1000
		addVisit_d = im.MRgst.read(1) / 1000

		if opType == 0 { // 无操作数
			im.MRgst.write(3, -1)
		} else if opType == 1 { // 单操作数
			EA_s = im.effectAddressing(addVisit_s, im.MRgst.read(2), proc, &RorM_s)
			im.MRgst.write(2, EA_s)
			im.MRgst.write(3, RorM_s)
		} else if opType == 2 { // 双操作数
			EA_s = im.effectAddressing(addVisit_s, im.MRgst.read(2), proc, &RorM_s)
			im.MRgst.write(2, EA_s)
			im.MRgst.write(3, RorM_s)

			EA_d = im.effectAddressing(addVisit_d, im.MRgst.read(4), proc, &RorM_d)
			im.MRgst.write(4, EA_d)
			im.MRgst.write(5, RorM_d)
		}

		tmp_dataLS := dataLS

		// 源操作数取值到 GM[16]
		if tmp_dataLS%2 == 1 && im.MRgst.read(5) == 0 {
			im.load(im.GM, 16, proc.MData, im.MRgst.read(4))
		}
		tmp_dataLS >>= 1

		// 目的操作数取值到 GM[15]
		if tmp_dataLS%2 == 1 && im.MRgst.read(3) == 0 {
			im.load(im.GM, 15, proc.MData, im.MRgst.read(2))
		}
		tmp_dataLS >>= 1

		// 寄存器操作数直接取
		if im.MRgst.read(3) == 1 {
			im.GM.mem[15] = im.readGM(im.MRgst.read(2), proc)
		}
		if im.MRgst.read(5) == 1 {
			im.GM.mem[16] = im.readGM(im.MRgst.read(4), proc)
		}
		if im.MRgst.read(3) == 2 {
			im.GM.mem[15] = im.MRgst.read(2)
		}
		if im.MRgst.read(5) == 2 {
			im.GM.mem[16] = im.MRgst.read(4)
		}

		switch im.MRgst.read(0) {
		case 900152: // MOV
			im.GM.mem[15] = im.GM.mem[16]
		case 902042: // LEA
			im.GM.write(15, im.MRgst.read(4))
		case 900272: // ADD
			im.GM.mem[15] += im.GM.mem[16]
		case 901072: // SUB
			im.GM.mem[15] -= im.GM.mem[16]
		case 902072: // MUL
			im.GM.mem[15] *= im.GM.mem[16]
		case 903072: // DIV
			tmp := im.GM.mem[15]
			im.GM.mem[15] /= im.GM.mem[16]
			im.GM.mem[0] = truncWord(tmp % im.GM.mem[16]) // 余数
		case 904072: // MOD
			im.GM.mem[15] %= im.GM.mem[16]
		case 905072: // AND
			im.GM.mem[15] &= im.GM.mem[16]
		case 906072: // OR
			im.GM.mem[15] |= im.GM.mem[16]
		case 907072: // XOR
			im.GM.mem[15] ^= im.GM.mem[16]
		case 908072: // SHL
			im.GM.mem[15] = shiftL(im.GM.mem[15], im.GM.mem[16])
		case 909072: // SHR
			im.GM.mem[15] = shiftR(im.GM.mem[15], im.GM.mem[16])
		case 911072: // ROL
			im.GM.mem[15] = rotL(im.GM.mem[15], im.GM.mem[16])
		case 912072: // ROR
			im.GM.mem[15] = rotR(im.GM.mem[15], im.GM.mem[16])
		case 900332: // CMP
			r := im.GM.mem[15] - im.GM.mem[16]
			if r > 0 {
				im.flag = 1
			} else if r < 0 {
				im.flag = -1
			} else {
				im.flag = 0
			}
		case 900421: // JA
			if im.flag > 0 {
				im.PC = im.GM.mem[15]
				im.flag = 0
			}
		case 900521: // JB
			if im.flag < 0 {
				im.PC = im.GM.mem[15]
				im.flag = 0
			}
		case 900621: // JMP
			im.PC = im.GM.mem[15]
			im.flag = 0
		case 900721: // JE
			if im.flag == 0 {
				im.PC = im.GM.mem[15]
				im.flag = 0
			}
		case 900821: // JNE
			if im.flag != 0 {
				im.PC = im.GM.mem[15]
				im.flag = 0
			}
		case 900961: // INC
			im.GM.mem[15]++
		case 901961: // NEG
			im.GM.mem[15] = -im.GM.mem[15]
		case 902961: // NOT
			im.GM.mem[15] = ^im.GM.mem[15]
		case 902121: // PUSH
			im.push(im.GM.mem[15], proc.S)
		case 902300: // PUSHA
			for ppk = 0; ppk < 15; ppk++ {
				im.push(im.GM.mem[ppk], proc.S)
			}
		case 902241: // POP
			im.GM.mem[15] = im.pop(proc.S)
		case 902400: // POPA
			for ppk = 14; ppk >= 0; ppk-- {
				im.GM.mem[ppk] = im.pop(proc.S)
			}
		case 904000: // INT
			for ppk = 0; ppk < 15; ppk++ {
				im.push(im.GM.mem[ppk], proc.S)
			}
			im.push(im.PC, proc.S)
			im.push(im.flag, proc.S)

			RES = 20
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d, instruction row:%d ", im.ID, proc.ID, im.PC/4-1)
				systemChecker.check(RES, sysLog[:])
			}
			return im.GM.mem[10] // 返回系统调用号
		case 904121: // CALL
			im.push(im.PC, proc.S)
			im.push(im.flag, proc.S)
			im.PC = im.GM.mem[15]
			im.flag = 0
		case 904200: // RET
			im.flag = im.pop(proc.S)
			im.PC = im.pop(proc.S)
		case 904300: // IRET
			return 3
		case 950000: // SYSR
			for ppk = 0; ppk < 15; ppk++ {
				im.push(im.GM.mem[ppk], proc.S)
			}
			im.push(im.PC, proc.S)
			im.push(im.flag, proc.S)
			return 5
		case 950100: // SYSW
			for ppk = 0; ppk < 15; ppk++ {
				im.push(im.GM.mem[ppk], proc.S)
			}
			im.push(im.PC, proc.S)
			im.push(im.flag, proc.S)
			return 6
		case 950200: // WAKE
			for ppk = 0; ppk < 15; ppk++ {
				im.push(im.GM.mem[ppk], proc.S)
			}
			im.push(im.PC, proc.S)
			im.push(im.flag, proc.S)
			return 7
		case 950300: // SET
			switch im.GM.read(11) {
			case 0:
				proc.communicate = im.GM.read(12)
			default:
				RES = 24
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d, property NO:%d ", im.ID, proc.ID, im.GM.read(11))
					systemChecker.check(RES, sysLog[:])
				}
			}
			proc.update = 1
		case 960000: // CLI
			proc.interrupt = 1
			proc.update = 1
		case 960100: // STI
			proc.interrupt = 0
			proc.update = 1
		}

		// 写回目的操作数
		if tmp_dataLS == 1 && im.MRgst.read(3) == 0 {
			im.store(proc.MData, im.MRgst.read(2), im.GM, 15)
		}
		if tmp_dataLS == 1 && im.MRgst.read(3) == 1 {
			im.GM.write(im.MRgst.read(2), truncWord(im.GM.mem[15]))
			// 写入输出寄存器 #13：立即按字符输出到 stdout
			// （INT 14 系统函数内 #13 用作返回标志，不触发输出）
			if im.MRgst.read(2) == outputReg && !im.isInt14(proc) {
				fmt.Printf("%c", im.GM.mem[15])
			}
		}

		im.load(im.MRgst, 0, proc.MCode, im.PC)
		im.PC++
		proc.exetime++
		i++

		// dump 日志：记录本周期有变化的寄存器与数据段内存
		if dumpEnabled == 1 {
			doDump(im, proc)
		}
	}

	if i == im.cycleTimes { // 时间片到
		im.PC--
		for ppk = 0; ppk < 15; ppk++ {
			im.push(im.GM.mem[ppk], proc.S)
		}
		im.push(im.PC, proc.S)
		im.push(im.flag, proc.S)

		RES = 22
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", im.ID, proc.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return 1
	} else if im.PC == proc.MCode.mSize { // 代码段溢出
		RES = 21
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", im.ID, proc.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	} else { // END 900000 或 HALT 800000，正常结束
		RES = 23
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", im.ID, proc.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return 0
	}
}
