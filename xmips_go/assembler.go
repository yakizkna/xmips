package main

import "strconv"

// Assembler 汇编器，对应 C++ 的 assembler 类
type Assembler struct {
	ID int
}

func newAssembler(key int) *Assembler {
	return &Assembler{ID: key}
}

// trans 将助记符翻译为数字操作码
func (a *Assembler) trans(op string, line int) int {
	switch op {
	case "MOV", "mov":
		return 900152
	case "ADD", "add":
		return 900272
	case "SUB", "sub":
		return 901072
	case "MUL", "mul":
		return 902072
	case "DIV", "div":
		return 903072
	case "MOD", "mod":
		return 904072
	case "AND", "and":
		return 905072
	case "OR", "or":
		return 906072
	case "XOR", "xor":
		return 907072
	case "CMP", "cmp":
		return 900332
	case "JA", "ja":
		return 900421
	case "JB", "jb":
		return 900521
	case "JMP", "jmp":
		return 900621
	case "JE", "je":
		return 900721
	case "JNE", "jne":
		return 900821
	case "INC", "inc":
		return 900961
	case "NEG", "neg":
		return 901961
	case "NOT", "not":
		return 902961
	case "LEA", "lea":
		return 902042
	case "PUSH", "push":
		return 902121
	case "POP", "pop":
		return 902241
	case "PUSHA", "pusha":
		return 902300
	case "POPA", "popa":
		return 902400
	case "LOC", "loc":
		return 910002
	case "DIM", "dim":
		return 910002
	case "END", "end":
		return 900000
	case "INT", "int":
		return 904000
	case "CALL", "call":
		return 904121
	case "RET", "ret":
		return 904200
	case "IRET", "iret":
		return 904300
	case "SYSR", "sysr":
		return 950000
	case "SYSW", "sysw":
		return 950100
	case "WAKE", "wake":
		return 950200
	case "SET", "set":
		return 950300
	case "CLI", "cli":
		return 960000
	case "STI", "sti":
		return 960100
	case "PRINT", "print":
		return 970021
	case "PRINTC", "printc":
		return 970121
	case "HALT", "halt":
		return 800000
	case "~":
		return 3 // 注释
	case "CMNT", "cmnt":
		return 3
	case "$":
		return 0 // 占位
	case ":":
		return 920001 // 标号
	case "PROC", "proc":
		return 920001
	}
	return -1
}

// trans2 操作数翻译，返回操作数值，addressing 为寻址方式
// 寻址方式编码：
// & 偏移寻址，权值8，寄存器编号乘100
// @ 间接寻址，权值4
// # 寄存器寻址，权值2
// 立即数，权值3
// addressing 为各寻址方式权值之和
func (a *Assembler) trans2(op string, vp []varNote, line int, addressing *int) int {
	ptmp := []byte(op)
	idx := 0
	advisit := 0
	ptrflag := false

	*addressing = 0

	if op == "$" {
		*addressing = 0
		return 0
	}

	// 偏移寻址 &
	if idx < len(ptmp) && ptmp[idx] == '&' {
		advisit += 8
		if idx+1 < len(ptmp) && ptmp[idx+1] >= '1' && ptmp[idx+1] <= '9' {
			advisit += 100 * int(ptmp[idx+1]-'0')
			idx += 2
		} else {
			RES = 53
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
				systemChecker.check(RES, sysLog[:])
			}
			return -1
		}
	}

	// 间接寻址 @
	if idx < len(ptmp) && ptmp[idx] == '@' {
		advisit += 4
		idx++
		ptrflag = true
	}

	// 寄存器寻址 # 或 立即数 !
	if idx < len(ptmp) && (ptmp[idx] == '#' || ptmp[idx] == '!') {
		if ptmp[idx] == '#' {
			// 寄存器寻址
			if advisit%10 == 8 { // &4#2 形式非法
				RES = 54
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
					systemChecker.check(RES, sysLog[:])
				}
				return -1
			}
			if idx+1 < len(ptmp) && ptmp[idx+1] >= '0' && ptmp[idx+1] <= '9' {
				advisit += 2
				*addressing = advisit
				registerNO, _ := strconv.Atoi(string(ptmp[idx+1:]))
				// 注意：原 C++ 代码用 atoi 只取连续数字
				// 但 # 后跟的数字可能多位，这里只取一位
				// 实际上 C++ 的 atoi 会取所有连续数字
				// 但寄存器编号最大14，所以需要正确解析
				// 修正：atoi 从 ptmp+1 开始，即 # 后面的全部数字
				registerStr := string(ptmp[idx+1:])
				registerNO, _ = strconv.Atoi(registerStr)
				if registerNO > 14 {
					RES = 58
					if systemChecker.showLevel(RES, sysLog[:]) {
						printf("D%-5d row:%d, register No:%d ", a.ID, line, registerNO)
						systemChecker.check(RES, sysLog[:])
					}
					return -1
				}
				return registerNO
			} else {
				RES = 55
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
					systemChecker.check(RES, sysLog[:])
				}
				return -1
			}
		} else {
			// 立即数寻址 !
			if idx+1 < len(ptmp) && ptmp[idx+1] >= '0' && ptmp[idx+1] <= '9' {
				*addressing = advisit
				v, _ := strconv.Atoi(string(ptmp[idx+1:]))
				return v
			} else {
				RES = 57
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
					systemChecker.check(RES, sysLog[:])
				}
				return -1
			}
		}
	} else {
		// 非上述前缀，为符号地址或立即数
		remaining := string(ptmp[idx:])
		i := 0
		for i < len(vp) && vp[i].valid == 1 {
			if remaining == vp[i].varName {
				*addressing = advisit
				return vp[i].pos
			}
			i++
		}

		if ptrflag {
			// 已使用 '@' 但不是已声明的符号地址
			if len(remaining) == 0 || remaining[0] < '0' || remaining[0] > '9' {
				RES = 51
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
					systemChecker.check(RES, sysLog[:])
				}
				return -1
			} else {
				RES = 56
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
					systemChecker.check(RES, sysLog[:])
				}
				return -1
			}
		} else {
			// 普通操作数：立即数
			if len(remaining) > 0 && (remaining[0] == '+' || remaining[0] == '-' || (remaining[0] >= '0' && remaining[0] <= '9')) {
				if advisit != 0 {
					RES = 56
					if systemChecker.showLevel(RES, sysLog[:]) {
						printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
						systemChecker.check(RES, sysLog[:])
					}
					return -1
				}
				advisit += 3
				*addressing = advisit
				v, _ := strconv.Atoi(remaining)
				return v
			}

			RES = 52
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d row:%d, operand:'%s' ", a.ID, line, op)
				systemChecker.check(RES, sysLog[:])
			}
			return -1
		}
	}
}

// varNote 变量表项
type varNote struct {
	valid   int
	varName string
	pos     int
}

// varNote2 跳转表项
type varNote2 struct {
	valid   int
	varName string
	pos     int
	found   int
}

func newVarNote() varNote {
	return varNote{}
}

func newVarNote2() varNote2 {
	return varNote2{}
}
