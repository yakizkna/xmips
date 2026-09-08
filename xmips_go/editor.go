package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// tokenReader 逐行读取文件，按空白分割成 token，支持跳过行剩余部分
type tokenReader struct {
	scanner *bufio.Scanner
	tokens  []string
	pos     int
	curLine string // 当前行完整文本（用于显示注释等）
}

func newTokenReader(scanner *bufio.Scanner) *tokenReader {
	return &tokenReader{scanner: scanner}
}

// fill 读取下一行并分割成 tokens
func (tr *tokenReader) fill() bool {
	for tr.pos >= len(tr.tokens) {
		if !tr.scanner.Scan() {
			return false
		}
		tr.curLine = tr.scanner.Text()
		tr.tokens = strings.Fields(tr.curLine)
		tr.pos = 0
	}
	return true
}

// next 读取下一个 token
func (tr *tokenReader) next() (string, bool) {
	if !tr.fill() {
		return "", false
	}
	tok := tr.tokens[tr.pos]
	tr.pos++
	return tok, true
}

// peek 读取但不消费下一个 token
func (tr *tokenReader) peek() (string, bool) {
	if !tr.fill() {
		return "", false
	}
	return tr.tokens[tr.pos], true
}

// skipRestOfLine 跳过当前行剩余的 token
func (tr *tokenReader) skipRestOfLine() {
	tr.tokens = nil
	tr.pos = 0
}

// isDataToken 判断 token 是否为数组初始数据
func isDataToken(s string) bool {
	if s == "" {
		return false
	}
	if s == ";" {
		return true
	}
	if s[0] >= '0' && s[0] <= '9' || s[0] == '-' || s[0] == '+' {
		return true
	}
	return false
}

// Editor 编辑器/汇编器，对应 C++ 的 editor 类
type Editor struct {
	ID             int
	codeBufferSize int
	codeBuffer     []int
	dataBuffer     []int
}

func newEditor(key, codeSize, dataSize int) *Editor {
	return &Editor{
		ID:             key,
		codeBufferSize: codeSize,
		codeBuffer:     make([]int, codeSize),
		dataBuffer:     make([]int, dataSize),
	}
}

// editorFromFile 从文件汇编，返回代码段长度，-1 表示错误
// dataNumInMData 返回数据段中数据的个数
func (e *Editor) editorFromFile(scanner *bufio.Scanner, a *Assembler, dataNumInMData *int) int {
	tr := newTokenReader(scanner)
	var o, d, s int
	var op, od, os_ string
	var opcodeType int
	var addressing int
	var addressingUnion int
	i := 0
	j := 0
	dataCounter := 0
	state := 0

	// 清空缓冲区，防止上一次汇编的残留数据串入本次
	for k := 0; k < len(e.codeBuffer); k++ {
		e.codeBuffer[k] = 0
	}
	for k := 0; k < len(e.dataBuffer); k++ {
		e.dataBuffer[k] = 0
	}

	// 变量表/标号表/跳转表：动态扩容，支持较长程序
	var varTable []varNote
	var tagTable []varNote
	var jumpTable []varNote2

	if displayMode == 1 {
		printf("display your statememts\n")
	}
	if displayMode == 1 {
		printf("%-4d", j)
	}

	// 读取第一个 token
	var ok bool
	op, ok = tr.next()
	if !ok {
		return -1
	}
	o = a.trans(op, j)
	opcodeType = o % 10

	// 根据操作数个数读取操作数
	switch opcodeType {
	case 0:
		d = 0
		s = 0
		if displayMode == 1 {
			fmt.Printf("%s\n", op)
		}
	case 1:
		s = 0
		od, _ = tr.next()
		if displayMode == 1 {
			fmt.Printf("%s %s\n", op, od)
		}
	case 2:
		if o != 910002 {
			od, _ = tr.next()
			os_, _ = tr.next()
			if displayMode == 1 {
				fmt.Printf("%s %s %s\n", op, od, os_)
			}
		} else {
			od, _ = tr.next()
			os_, _ = tr.next()
			if displayMode == 1 {
				fmt.Printf("%s %s %s ", op, od, os_)
			}
		}
	}

	// 处理寻址方式
	if o != 910002 && o != 920001 && o != 3 && opcodeType != 0 {
		if opcodeType == 1 {
			if o != 900421 && o != 900521 && o != 900621 && o != 900721 && o != 900821 && o != 904121 {
				d = a.trans2(od, varTable, j, &addressing)
				addressingUnion = addressing
			}
		}
		if opcodeType == 2 {
			d = a.trans2(od, varTable, j, &addressing)
			if addressing == 3 {
				RES = 197
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d operand:'%s' ", e.ID, od)
					systemChecker.check(RES, sysLog[:])
				}
				state = 8
			}
			addressingUnion = addressing
			s = a.trans2(os_, varTable, j, &addressing)
			addressingUnion += addressing * 1000
		}
	}

	for o != 900000 && o != -1 && i < e.codeBufferSize {
		if o == 920001 { // 标号
			m := 0
			for m < len(tagTable) && tagTable[m].valid != 0 {
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

			tagTable = append(tagTable, varNote{valid: 1, varName: od, pos: i})

		} else if o == 900421 || o == 900521 || o == 900621 || o == 900721 || o == 900821 || o == 904121 {
			// 跳转和 call
			jumpTable = append(jumpTable, varNote2{valid: 1, varName: od, pos: i})

			e.codeBuffer[i] = o
			e.codeBuffer[i+1] = 3
			e.codeBuffer[i+3] = 0
			i += 4
			j++

		} else if o == 910002 { // DIM/LOC 声明变量
			if (od[0] >= 'a' && od[0] <= 'z') || (od[0] >= 'A' && od[0] <= 'Z') {
				t := 0
				for t < len(varTable) && varTable[t].valid != 0 {
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

				varTable = append(varTable, varNote{valid: 1, varName: od, pos: dataCounter})

				if len(os_) > 0 && os_[0] == 'D' { // 数组声明
					arrLen, _ := strconv.Atoi(os_[1:])
					dataNum := dataCounter
					dataCounter += arrLen

					// 先 peek 下一个 token，判断是否有数组初始值
					nextTok, hasNext := tr.peek()
					if hasNext && isDataToken(nextTok) {
						// 有初始值，读取直到 ';'
						var dataTokens []string
						for {
							tok, ok := tr.next()
							if !ok {
								break
							}
							// ';' 之后的部分（如 ';~ 注释'）属于注释，截断并丢弃同行剩余内容
							if idx := strings.Index(tok, ";"); idx >= 0 {
								tok = tok[:idx+1]
								dataTokens = append(dataTokens, tok)
								tr.skipRestOfLine()
								break
							}
							dataTokens = append(dataTokens, tok)
						}

						joined := strings.Join(dataTokens, " ")
						joined = strings.TrimSuffix(joined, ";")
						if strings.HasSuffix(joined, ";") {
							joined = joined[:len(joined)-1]
						}

						parts := strings.Split(joined, ",")
						for _, part := range parts {
							part = strings.TrimSpace(part)
							if part == "" {
								continue
							}
							part = strings.TrimSuffix(part, ";")
							if part == "" {
								continue
							}
							if (part[0] >= '0' && part[0] <= '9') || part[0] == '-' || part[0] == '+' {
								if dataNum+1 > dataCounter {
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
					}
				} else {
					// 单个变量赋值
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
				tr.skipRestOfLine()
				if displayMode == 1 {
					fmt.Printf("%s\n", tr.curLine)
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

		if displayMode == 1 {
			printf("%-4d", j)
		}

		// 读取下一条指令
		op, ok = tr.next()
		if !ok {
			break
		}
		o = a.trans(op, j)
		opcodeType = o % 10

		od = ""
		os_ = ""
		switch opcodeType {
		case 0:
			d = 0
			s = 0
			if displayMode == 1 {
				fmt.Printf("%s\n", op)
			}
		case 1:
			s = 0
			od, _ = tr.next()
			if displayMode == 1 {
				fmt.Printf("%s %s\n", op, od)
			}
		case 2:
			if o != 910002 {
				od, _ = tr.next()
				os_, _ = tr.next()
				if displayMode == 1 {
					fmt.Printf("%s %s %s\n", op, od, os_)
				}
			} else {
				od, _ = tr.next()
				os_, _ = tr.next()
				if displayMode == 1 {
					fmt.Printf("%s %s %s ", op, od, os_)
				}
			}
		}

		if o != 910002 && o != 920001 && o != 3 && opcodeType != 0 {
			if opcodeType == 1 {
				if o != 900421 && o != 900521 && o != 900621 && o != 900721 && o != 900821 && o != 904121 {
					d = a.trans2(od, varTable, j, &addressing)
					addressingUnion = addressing
				}
			}
			if opcodeType == 2 {
				d = a.trans2(od, varTable, j, &addressing)
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
		m := 0
		for m < len(tagTable) && tagTable[m].valid != 0 {
			n := 0
			for n < len(jumpTable) && jumpTable[n].valid != 0 {
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
		for m < len(jumpTable) && jumpTable[m].valid != 0 {
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

		if displayMode == 1 {
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
	scanner.Split(bufio.ScanLines)

	r := e.editorFromFile(scanner, a, &dataInMData)
	if r == -1 {
		return -1
	}

	// 生成 .co 文件
	tmps := s
	dotIdx := strings.LastIndex(tmps, ".")
	if dotIdx >= 0 {
		tmps = tmps[:dotIdx]
	}
	tmps += ".co"

	coFile, err := os.Create(tmps)
	if err != nil {
		return -1
	}
	defer coFile.Close()

	for k := 0; k < r; k++ {
		fmt.Fprintf(coFile, "%d ", e.codeBuffer[k])
	}
	fmt.Fprintf(coFile, "\n%d\n", dataInMData)
	for k := 0; k < dataInMData; k++ {
		fmt.Fprintf(coFile, "%d ", e.dataBuffer[k])
	}
	fmt.Fprintf(coFile, "\n")

	return r
}

// Load 从 .co 文件加载到进程的代码段和数据段
func Load(p *Process, name string) {
	// 打开 .co 文件
	coName := name
	dotIdx := strings.LastIndex(coName, ".")
	if dotIdx >= 0 {
		coName = coName[:dotIdx]
	}
	coName += ".co"

	f, err := os.Open(coName)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		return
	}
	parts := strings.Fields(scanner.Text())
	for k, part := range parts {
		if k < len(p.MCode.mem) {
			p.MCode.mem[k], _ = strconv.Atoi(part)
		}
	}

	if scanner.Scan() {
		dataCount, _ := strconv.Atoi(scanner.Text())
		if scanner.Scan() {
			dparts := strings.Fields(scanner.Text())
			for k, dp := range dparts {
				if k < len(p.MData.mem) {
					p.MData.mem[k], _ = strconv.Atoi(dp)
				}
			}
		}
		_ = dataCount
	}
}
