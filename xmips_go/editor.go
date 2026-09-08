package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Editor 编辑器/汇编器，对应 C++ 的 editor 类
type Editor struct {
	ID             int
	codeBufferSize int
	codeBuffer     []int
	dataBuffer     []int
}

func newEditor(key, codeSize int) *Editor {
	return &Editor{
		ID:             key,
		codeBufferSize: codeSize,
		codeBuffer:     make([]int, 200),
		dataBuffer:     make([]int, 100),
	}
}

// editorFromFile 从文件汇编，返回代码段长度，-1 表示错误
// dataNumInMData 返回数据段中数据的个数
func (e *Editor) editorFromFile(scanner *bufio.Scanner, a *Assembler, dataNumInMData *int) int {
	var o, d, s int // opcode, destination operand, source operand
	var op, od, os_ string
	var comment string
	var opcodeType int   // 0,1,2 operands
	var addressing int   // addressing way
	var addressingUnion int
	i := 0 // 实际行号（代码段位置）
	j := 0 // 显示行号
	dataCounter := 0
	state := 0

	var v int    // 变量数
	var tagv int // 标号数
	var jmpv int // 跳转数

	var varTable [20]varNote
	var tagTable [20]varNote
	var jumpTable [20]varNote2

	if displayMode != 0 {
		printf("display your statememts\n")
	}
	if displayMode != 0 {
		printf("%-4d", j)
	}

	// 读取第一个 token
	if !scanner.Scan() {
		return -1
	}
	op = scanner.Text()
	o = a.trans(op, j)
	opcodeType = o % 10

	// 根据操作数个数读取操作数
	switch opcodeType {
	case 0:
		d = 0
		s = 0
		if displayMode != 0 {
			fmt.Printf("%s\n", op)
		}
	case 1:
		s = 0
		if scanner.Scan() {
			od = scanner.Text()
		}
		if displayMode != 0 {
			fmt.Printf("%s %s\n", op, od)
		}
	case 2:
		if o != 910002 { // 非 dim 语句
			if scanner.Scan() {
				od = scanner.Text()
			}
			if scanner.Scan() {
				os_ = scanner.Text()
			}
			if displayMode != 0 {
				fmt.Printf("%s %s %s\n", op, od, os_)
			}
		} else {
			if scanner.Scan() {
				od = scanner.Text()
			}
			if scanner.Scan() {
				os_ = scanner.Text()
			}
			if displayMode != 0 {
				fmt.Printf("%s %s %s ", op, od, os_)
			}
		}
	}

	// 处理寻址方式
	if o != 910002 && o != 920001 && o != 3 && opcodeType != 0 {
		if opcodeType == 1 {
			if o != 900421 && o != 900521 && o != 900621 && o != 900721 && o != 900821 && o != 904121 {
				d = a.trans2(od, varTable[:], j, &addressing)
				addressingUnion = addressing
			}
		}
		if opcodeType == 2 {
			d = a.trans2(od, varTable[:], j, &addressing)
			if addressing == 3 {
				RES = 197
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d operand:'%s' ", e.ID, od)
					systemChecker.check(RES, sysLog[:])
				}
				state = 8
			}
			addressingUnion = addressing
			s = a.trans2(os_, varTable[:], j, &addressing)
			addressingUnion += addressing * 1000
		}
	}

	for o != 900000 && o != -1 && i < e.codeBufferSize {
		if o == 920001 { // 标号
			m := 0
			for tagTable[m].valid != 0 {
				if od == tagTable[m].varName {
					RES = 193
					if systemChecker.showLevel(RES, sysLog[:]) {
						printf("D%-5d repeated tag:%s ", e.ID, od)
						systemChecker.check(RES, sysLog[:])
					}
					state = 1
					break
				}
				m++
			}

			tagTable[tagv].varName = od
			tagTable[tagv].pos = i
			tagTable[tagv].valid = 1
			tagv++

		} else if o == 900421 || o == 900521 || o == 900621 || o == 900721 || o == 900821 || o == 904121 {
			// 跳转和 call
			jumpTable[jmpv].varName = od
			jumpTable[jmpv].pos = i
			jumpTable[jmpv].valid = 1
			jmpv++

			e.codeBuffer[i] = o
			e.codeBuffer[i+1] = 3 // 立即数寻址
			e.codeBuffer[i+3] = 0
			i += 4
			j++

		} else if o == 910002 { // DIM/LOC 声明变量
			if (od[0] >= 'a' && od[0] <= 'z') || (od[0] >= 'A' && od[0] <= 'Z') {
				t := 0
				for varTable[t].valid != 0 {
					if od == varTable[t].varName {
						RES = 196
						if systemChecker.showLevel(RES, sysLog[:]) {
							printf("D%-5d repeated indentifier:%s ", e.ID, od)
							systemChecker.check(RES, sysLog[:])
						}
						state = 2
						break
					}
					t++
				}

				varTable[v].varName = od
				varTable[v].pos = dataCounter
				varTable[v].valid = 1
				v++

				if len(os_) > 0 && os_[0] == 'D' { // 数组声明
					arrLen, _ := strconv.Atoi(os_[1:])
					dataNum := dataCounter
					dataCounter += arrLen

					// C++ 用 fgetc 逐字符读取直到 ';'
					// ScanWords 模式下，读取 token 直到遇到含 ';' 的
					var dataTokens []string
					for scanner.Scan() {
						token := scanner.Text()
						dataTokens = append(dataTokens, token)
						if strings.HasSuffix(token, ";") || token == ";" {
							break
						}
					}

					// 将所有 token 拼接，按逗号分隔解析
					joined := strings.Join(dataTokens, " ")
					// 去掉分号
					joined = strings.TrimSuffix(joined, ";")
					if strings.HasSuffix(joined, ";") {
						joined = joined[:len(joined)-1]
					}
					fmt.Printf("%s;\n", joined)

					// 按逗号分割
					parts := strings.Split(joined, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if part == "" {
							continue
						}
						// 去掉可能残留的分号
						part = strings.TrimSuffix(part, ";")
						if part == "" {
							continue
						}
						if (part[0] >= '0' && part[0] <= '9') || part[0] == '-' || part[0] == '+' {
							if dataNum+1 > dataCounter {
								fmt.Println()
								RES = 191
								if systemChecker.showLevel(RES, sysLog[:]) {
									printf("D%-5d ", e.ID)
									systemChecker.check(RES, sysLog[:])
								}
								state = 3
								break
							}
							k, _ := strconv.Atoi(part)
							e.dataBuffer[dataNum] = k
							dataNum++
						}
					}
				} else {
					// 单个变量赋值
					fmt.Println()
					if len(os_) > 0 && ((os_[0] >= '0' && os_[0] <= '9') || os_[0] == '-' || os_[0] == '+') {
						k, _ := strconv.Atoi(os_)
						e.dataBuffer[dataCounter] = k
						dataCounter++
					} else {
						RES = 192
						if systemChecker.showLevel(RES, sysLog[:]) {
							printf("D%-5d ", e.ID)
							systemChecker.check(RES, sysLog[:])
						}
						state = 4
					}
				}

			} else {
				RES = 13
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d variable name:\"%s\" ", e.ID, od)
					systemChecker.check(RES, sysLog[:])
				}
				state = 5
			}
		} else {
			// 普通指令
			if o == 3 { // 注释
				if scanner.Scan() {
					comment = scanner.Text()
				}
				if displayMode != 0 {
					fmt.Printf("%s %s\n", op, comment)
				}
			} else {
				if o == 800000 {
					addressingUnion = 0
					d = 0
					s = 0
				}

				e.codeBuffer[i] = o
				e.codeBuffer[i+1] = addressingUnion
				e.codeBuffer[i+2] = d
				e.codeBuffer[i+3] = s

				i += 4
				j++
			}
		}

		if displayMode != 0 {
			printf("%-4d", j)
		}

		// 读取下一条指令
		if !scanner.Scan() {
			break
		}
		op = scanner.Text()
		o = a.trans(op, j)
		opcodeType = o % 10

		od = ""
		os_ = ""
		switch opcodeType {
		case 0:
			d = 0
			s = 0
			if displayMode != 0 {
				fmt.Printf("%s\n", op)
			}
		case 1:
			s = 0
			if scanner.Scan() {
				od = scanner.Text()
			}
			if displayMode != 0 {
				fmt.Printf("%s %s\n", op, od)
			}
		case 2:
			if o != 910002 {
				if scanner.Scan() {
					od = scanner.Text()
				}
				if scanner.Scan() {
					os_ = scanner.Text()
				}
				if displayMode != 0 {
					fmt.Printf("%s %s %s\n", op, od, os_)
				}
			} else {
				if scanner.Scan() {
					od = scanner.Text()
				}
				if scanner.Scan() {
					os_ = scanner.Text()
				}
				if displayMode != 0 {
					fmt.Printf("%s %s %s ", op, od, os_)
				}
			}
		}

		if o != 910002 && o != 920001 && o != 3 && opcodeType != 0 {
			if opcodeType == 1 {
				if o != 900421 && o != 900521 && o != 900621 && o != 900721 && o != 900821 && o != 904121 {
					d = a.trans2(od, varTable[:], j, &addressing)
					addressingUnion = addressing
				}
			}
			if opcodeType == 2 {
				d = a.trans2(od, varTable[:], j, &addressing)
				if addressing == 3 {
					RES = 197
					if systemChecker.showLevel(RES, sysLog[:]) {
						printf("D%-5d operand:'%s' ", e.ID, od)
						systemChecker.check(RES, sysLog[:])
					}
					state = 8
				}
				addressingUnion = addressing
				s = a.trans2(os_, varTable[:], j, &addressing)
				addressingUnion += addressing * 1000
			}
		}
	}

	if o == 900000 && state == 0 {
		// 正常结束，修正跳转表
		m := 0
		for tagTable[m].valid != 0 {
			n := 0
			for jumpTable[n].valid != 0 {
				if jumpTable[n].found == 0 {
					if jumpTable[n].varName == tagTable[m].varName {
						e.codeBuffer[jumpTable[n].pos+2] = tagTable[m].pos
						jumpTable[n].found = 1
					}
				}
				n++
			}
			m++
		}

		m = 0
		for jumpTable[m].valid != 0 {
			if jumpTable[m].found == 0 {
				RES = 194
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d unfound tag:%s ", e.ID, jumpTable[m].varName)
					systemChecker.check(RES, sysLog[:])
				}
				state = 7
				break
			}
			m++
		}

		if displayMode != 0 {
			fmt.Println("end edit")
		}
		e.codeBuffer[i] = o
		i++
		*dataNumInMData = dataCounter
		return i
	}

	if o == -1 {
		RES = 14
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d opcode:\"%s\" ", e.ID, op)
			systemChecker.check(RES, sysLog[:])
		}
	} else if state == 0 {
		RES = 15
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d numbers of lines:%d ", e.ID, i/4)
			systemChecker.check(RES, sysLog[:])
		}
	}
	return -1
}

// ASM 汇编入口，生成 .co 文件
func (e *Editor) ASM(f *os.File, a *Assembler, s string) int {
	var dataInMData int
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	scanner.Split(bufio.ScanWords)

	r := e.editorFromFile(scanner, a, &dataInMData)
	if r == -1 {
		return -1
	}

	// 生成 .co 文件
	tmps := s
	// 找到最后一个 '.' 替换扩展名
	dotIdx := strings.LastIndex(tmps, ".")
	if dotIdx >= 0 {
		tmps = tmps[:dotIdx]
	}
	tmps += ".co"

	fsave, err := os.Create(tmps)
	if err != nil {
		RES = 195
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d ", e.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	defer fsave.Close()

	for i := 0; i < r; i++ {
		fmt.Fprintf(fsave, "%d ", e.codeBuffer[i])
	}

	fmt.Fprintf(fsave, "%d ", dataInMData)

	for i := 0; i < dataInMData; i++ {
		fmt.Fprintf(fsave, "%d ", e.dataBuffer[i])
	}

	return 0
}

// Load 从 .co 文件加载代码和数据到进程
func Load(pptr *Process, s string) int {
	dotIdx := strings.LastIndex(s, ".")
	tmps := s
	if dotIdx >= 0 {
		tmps = s[:dotIdx]
	}
	tmps += ".co"

	fo, err := os.Open(tmps)
	if err != nil {
		return -1
	}
	defer fo.Close()

	scanner := bufio.NewScanner(fo)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	scanner.Split(bufio.ScanWords)

	i := 0
	for scanner.Scan() {
		cw, _ := strconv.Atoi(scanner.Text())
		if cw == 900000 {
			pptr.MCode.write(i, cw)
			i++
			break
		}
		pptr.MCode.write(i, cw)
		i++
	}

	// 读取数据段长度
	if !scanner.Scan() {
		return 0
	}
	length, _ := strconv.Atoi(scanner.Text())

	for i := 0; i < length; i++ {
		if scanner.Scan() {
			cw, _ := strconv.Atoi(scanner.Text())
			pptr.MData.write(i, cw)
		}
	}

	return 0
}
