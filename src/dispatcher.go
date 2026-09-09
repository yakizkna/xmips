package main

import (
	"bufio"
	"os"
)

// fdReg 文件/套接字句柄寄存器。
// 不用 #13 是因为 #13 是输出寄存器，用户 MOV 写入 #13 会触发 stdout；
// #9 为通用寄存器，无副作用，且在 INT 压栈/恢复范围内，跨系统调用可靠。
const fdReg = 9

var stdinReader = bufio.NewReader(os.Stdin)

// Dispatcher 调度器，对应 C++ 的 dispatcher 类
type Dispatcher struct {
	ID                   int
	pcbNum               int
	pcbSet               []PCB
	usedPcbNumber        int
	sysTem               *Process     // 当前运行的系统函数
	runb                 *PCB         // 当前运行的进程 PCB
	readyb               *PCBList     // 就绪队列
	waitb                [20]*PCBList // 等待队列
	systemShare          [200]int     // 系统共享区
	priorityCounterTable [3]int

	Finished *Queue   // 完成队列
	SysCall  *SysList // 系统函数列表
	FS       *FsDev   // 文件系统/套接字子系统（dispatcher 原生系统调用）
}

func newDispatcher(key, sysFunLen, finishLen, pcbs int) *Dispatcher {
	d := &Dispatcher{
		ID:       key,
		pcbNum:   pcbs,
		pcbSet:   make([]PCB, pcbs),
		readyb:   newPCBList(key * 10),
		Finished: newQueue(key*10+3, finishLen),
		SysCall:  newSysList(key*10+2, sysFunLen),
	}
	d.priorityCounterTable[0] = 10
	d.priorityCounterTable[1] = 5
	d.priorityCounterTable[2] = 2
	for i := 0; i < 20; i++ {
		d.waitb[i] = newPCBList(key*10 + 100 + i)
	}
	return d
}

// loader 加载进程到 PCB 并加入就绪队列
func (d *Dispatcher) loader(pptr *Process) {
	i := d.pcbManagement()
	if i != -1 {
		d.pcbSet[i].link(pptr)
		d.readyb.enQueue(&d.pcbSet[i])
	} else {
		RES = 184
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", d.ID, pptr.ID)
			systemChecker.check(RES, sysLog[:])
		}
	}
}

// loaderSysfun 加载系统函数到 PCB 并插入就绪队列头部
func (d *Dispatcher) loaderSysfun(pptr *Process) {
	i := d.pcbManagement()
	if i != -1 {
		d.pcbSet[i].link(pptr)
		d.readyb.insertToHead(&d.pcbSet[i])
	} else {
		RES = 184
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", d.ID, pptr.ID)
			systemChecker.check(RES, sysLog[:])
		}
	}
}

// swap2 调度主循环
func (d *Dispatcher) swap2(im *Interpreter) int {
	counter := 0
	r := 1
	var outPauseb *PCB

	for d.readyb.deQueue(&d.runb) != -1 {
		d.runb.state = 1 // process chosen to run

		RES = 170
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
			systemChecker.check(RES, sysLog[:])
		}

		counter = 0
		for counter < d.priorityCounterTable[d.runb.priority] && r == 1 {
			r = im.exer(d.runb.pptr)
			d.runb.exetime = d.runb.pptr.exetime
			counter++
		}

		if d.runb.pptr.update == 1 {
			d.runb.updateProc(d.runb.pptr)
			d.runb.pptr.update = 0
		}

		switch r {
		case -1: // failed
			d.runb.state = 4
			RES = 78
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
				systemChecker.check(RES, sysLog[:])
			}
			r = 1

		case 0: // finished
			d.runb.state = 3
			d.runb.pptr.writeBack(d.runb)
			d.Finished.enQueue(d.runb.pptr)
			RES = 177
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
				systemChecker.check(RES, sysLog[:])
			}
			d.pcbRelease(d.runb)
			r = 1

		case 1: // time slice arrived
			if d.runb.interrupt == 1 { // CLI 禁止中断
				RES = 189
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
				d.readyb.insertToHead(d.runb)
				RES = 188
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
				break
			}
			d.runb.state = 0
			d.readyb.enQueue(d.runb)
			RES = 171
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
				systemChecker.check(RES, sysLog[:])
			}

		case 3: // system call return (IRET)
			mode := im.GM.read(10)
			d.runb.state = 3
			RES = 174
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d system function ID:%d, caller process ID:%d ", d.ID, d.runb.ID, d.runb.callerID)
				systemChecker.check(RES, sysLog[:])
			}

			if mode != 1 { // 0=正常返回, 2=挂起返回；唤醒调用者
				outPauseb = nil
				d.waitb[d.runb.ID].deQueue(&outPauseb)
				if outPauseb != nil && outPauseb.ID == d.runb.callerID {
					// INT 14 返回：把缓冲区满/EOF 标志（系统函数 #13）写回调用者 #13 槽
					// 栈布局（INT 时压入）: GM[0..14], PC, flag，sp-4 对应 #13
					if d.runb.ID == inputReg-10 {
						sp := outPauseb.pptr.S.SP
						outPauseb.pptr.S.write(sp-4, im.GM.read(13))
					}
					if mode == 0 {
						d.readyb.insertToHead(outPauseb)
						RES = 188
						if systemChecker.showLevel(RES, sysLog[:]) {
							printf("D%-5d process ID:%d ", d.ID, outPauseb.ID)
							systemChecker.check(RES, sysLog[:])
						}
					} else { // mode == 2
						d.readyb.enQueue(outPauseb)
						outPauseb.state = 0
						RES = 175
						if systemChecker.showLevel(RES, sysLog[:]) {
							printf("D%-5d process ID:%d, pause cause:%d ", d.ID, outPauseb.ID, d.runb.ID)
							systemChecker.check(RES, sysLog[:])
						}
					}
				}
			} else { // mode == 1: suspend caller
				RES = 185
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d, suspended process ID:%d ", d.ID, d.runb.ID, d.runb.pptr.callerID)
					systemChecker.check(RES, sysLog[:])
				}
			}

			d.runb.pptr.MData.mem = nil // destroy temp sysfun copy
			// 注意：不 delete runb.pptr，因为它是由 sysList.copy 创建的临时副本
			// 但在 Go 中 GC 会处理
			d.pcbRelease(d.runb)
			r = 1

		case 5: // SYSR - read system share
			if d.runb.priority != 0 {
				RES = 180
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d, process priority:%d ", d.ID, d.runb.ID, d.runb.priority)
					systemChecker.check(RES, sysLog[:])
				}
			} else {
				address := im.GM.read(11)
				length := im.GM.read(12)
				for i := 0; i < length; i++ {
					d.runb.pptr.MData.write(d.runb.pptr.MData.mSize-length+i, d.systemShare[address+i])
				}
				RES = 181
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
				d.readyb.insertToHead(d.runb)
				RES = 188
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
			}
			r = 1

		case 6: // SYSW - write system share
			if d.runb.priority != 0 {
				RES = 180
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d, process priority:%d ", d.ID, d.runb.ID, d.runb.priority)
					systemChecker.check(RES, sysLog[:])
				}
			} else {
				address := im.GM.read(11)
				length := im.GM.read(12)
				for i := 0; i < length; i++ {
					d.systemShare[address+i] = d.runb.pptr.MData.read(d.runb.pptr.MData.mSize - length + i)
				}
				RES = 182
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
				d.readyb.insertToHead(d.runb)
				RES = 188
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
			}
			r = 1

		case 7: // WAKE - wake up process
			if d.runb.priority != 0 {
				RES = 187
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d, process priority:%d ", d.ID, d.runb.ID, d.runb.priority)
					systemChecker.check(RES, sysLog[:])
				}
			} else {
				waked := im.GM.read(11)
				where := im.GM.read(12)
				if where >= 20 || where < 0 {
					RES = 190
					if systemChecker.showLevel(RES, sysLog[:]) {
						printf("D%-5d process ID:%d requested wait queue NO:%d ", d.ID, d.runb.ID, where)
						systemChecker.check(RES, sysLog[:])
					}
				} else {
					var tmpPcbptr *PCB
					if d.waitb[where].CommunicateGet(&tmpPcbptr, waked) != -1 {
						d.readyb.enQueue(tmpPcbptr)
						RES = 186
						if systemChecker.showLevel(RES, sysLog[:]) {
							printf("D%-5d process ID:%d, waked process ID:%d ", d.ID, d.runb.ID, tmpPcbptr.ID)
							systemChecker.check(RES, sysLog[:])
						}
					}
				}
				d.readyb.insertToHead(d.runb)
				RES = 188
				if systemChecker.showLevel(RES, sysLog[:]) {
					printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
					systemChecker.check(RES, sysLog[:])
				}
			}
			r = 1

		case 15, 16, 17, 18: // 文件/套接字系统调用（dispatcher 原生，无 .scp）
			resp := d.fsSyscall(im, r)
			// 结果写回调用者栈的 fdReg(#9) 槽（INT 压栈帧内），恢复现场时回到 GM[fdReg]
			sp := d.runb.pptr.S.SP
			val := resp
			if resp == fdNoData {
				val = 0 // 用户视角：当前暂无数据可读
			}
			d.runb.pptr.S.write(sp-(17-fdReg), val)

			RES = 205
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d caller process ID:%d, file/socket call NO:%d, result:%d ", d.ID, d.runb.ID, r-10, val)
				systemChecker.check(RES, sysLog[:])
			}
			if resp == fdNoData && !singleProc {
				// socket 无可读数据（多进程非阻塞）：让出时间片，放就绪队尾，
				// 其他进程插队先运行，本进程稍后被重新调度回来重试，避免忙自旋。
				d.readyb.enQueue(d.runb)
				r = 1
				break
			}
			d.readyb.insertToHead(d.runb)
			RES = 188
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d ", d.ID, d.runb.ID)
				systemChecker.check(RES, sysLog[:])
			}
			r = 1

		case 13: // INT 13 输出字符串（dispatcher 原生，无 .scp）
			// 从调用者数据段 #11 处读到 0，逐字符输出到 stdout
			addr := im.GM.read(11)
			for {
				c := d.runb.pptr.MData.read(addr)
				if c == 0 || c < 0 {
					break
				}
				// 按原始字节输出（直接写 stdout），避免 rune 被 UTF-8 重编码导致非 ASCII 双重编码
				os.Stdout.Write([]byte{byte(c)})
				addr++
			}
			d.readyb.insertToHead(d.runb)
			r = 1

		case 14: // INT 14 输入到缓冲区（dispatcher 原生，无 .scp）
			// 从 stdin 读字符存入调用者数据段 #11 起的缓冲，最多 #12 个；
			// 返回标志写回 #13 槽：#13=0 读到 EOF，#13=1 缓冲区满（可再 INT 14 继续读）
			base := im.GM.read(11)
			size := im.GM.read(12)
			n := 0
			flag := 1 // 默认：缓冲满
			for n < size {
				b, err := stdinReader.ReadByte()
				if err != nil {
					flag = 0 // EOF
					break
				}
				d.runb.pptr.MData.write(base+n, int(b))
				n++
			}
			// 写回 #13 槽（sp-4），INT 恢复现场时回到 GM[13]
			d.runb.pptr.S.write(d.runb.pptr.S.SP-(17-13), flag)
			d.readyb.insertToHead(d.runb)
			RES = 205
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d caller process ID:%d, INT14 input, size:%d flag:%d ", d.ID, d.runb.ID, size, flag)
				systemChecker.check(RES, sysLog[:])
			}
			r = 1

		default: // system call (r >= 10)
			RES = 173
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d caller process ID:%d, system call NO:%d ", d.ID, d.runb.ID, r-10)
				systemChecker.check(RES, sysLog[:])
			}

			d.SysCall.copy(r-10, &d.sysTem)
			// 不复制 MData，直接指向调用者的数据段
			d.sysTem.MData.mem = d.runb.pptr.MData.mem
			d.sysTem.MData.mSize = d.runb.pptr.MData.mSize
			d.sysTem.callerID = d.runb.ID

			d.waitb[r-10].enQueue(d.runb)
			d.runb.state = 2

			RES = 172
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d, pause cause:%d ", d.ID, d.runb.ID, d.sysTem.ID)
				systemChecker.check(RES, sysLog[:])
			}

			d.loaderSysfun(d.sysTem)

			RES = 188
			if systemChecker.showLevel(RES, sysLog[:]) {
				printf("D%-5d process ID:%d ", d.ID, d.sysTem.ID)
				systemChecker.check(RES, sysLog[:])
			}

			r = 1
		}
	}

	RES = 179
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d interpreter ID:%d ", d.ID, im.ID)
		systemChecker.check(RES, sysLog[:])
	}
	return 0
}

// fsSyscall 文件/套接字系统调用（INT 15/16/17/18），由 dispatcher 原生执行
// 返回值统一通过 fdReg(#9) 槽写回调用者（见 swap2 的 case 15..18）
func (d *Dispatcher) fsSyscall(im *Interpreter, call int) int {
	if d.FS == nil {
		return -2
	}
	switch call {
	case 15: // open: #11=路径地址, #12=模式
		return d.FS.fsOpen(d.runb.pptr, im.GM.read(11), im.GM.read(12))
	case 16: // read: #11=缓冲地址, #12=最大长度, fdReg=fd
		return d.FS.fsRead(d.runb.pptr, im.GM.read(11), im.GM.read(fdReg), im.GM.read(12))
	case 17: // write
		return d.FS.fsWrite(d.runb.pptr, im.GM.read(11), im.GM.read(fdReg), im.GM.read(12))
	case 18: // close
		return d.FS.fsClose(im.GM.read(fdReg))
	}
	return -2
}

// pcbManagement 分配一个空闲 PCB
func (d *Dispatcher) pcbManagement() int {
	i := 0
	for i < d.pcbNum && d.pcbSet[i].valid != 0 {
		i++
	}
	if i == d.pcbNum {
		RES = 183
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d ", d.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	d.pcbSet[i].valid = 1
	d.usedPcbNumber++
	return i
}

// pcbRelease 释放 PCB
func (d *Dispatcher) pcbRelease(pblock *PCB) {
	pblock.ID = 0
	pblock.state = 0
	pblock.priority = 0
	pblock.callerID = 0
	pblock.communicate = 0
	pblock.valid = 0
	pblock.pptr = nil
	pblock.next = nil
	d.usedPcbNumber--
}
