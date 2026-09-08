package main

// Process 进程类，对应 C++ 的 process 类
type Process struct {
	ID          int
	state       int // 0 ready, 1 run, 2 wait, 3 finished, 4 failed
	exetime     int
	priority    int
	callerID    int
	communicate int
	interrupt   int // 0 allow, 1 dont allow
	update      int

	MCode *Memory
	MData *Memory
	S     *Stack
}

func newProcess(pid, csz, dsz, ssz, prty int) *Process {
	return &Process{
		ID:          pid,
		priority:    prty,
		state:       0,
		exetime:     0,
		callerID:    0,
		communicate: 0,
		update:      0,
		interrupt:   0,
		MCode:       newMemory(10*pid, csz),
		MData:       newMemory(10*pid+1, dsz),
		S:           newStack(10*pid+3, ssz),
	}
}

// defaultProcess 创建默认模板进程
func defaultProcess(pid, csz, dsz, ssz, prty int) *Process {
	return newProcess(pid, csz, dsz, ssz, prty)
}

// copy 从 sproc 复制进程属性（不复制内存指针，单独分配）
func (p *Process) copy(sproc *Process) {
	*p = *sproc
	p.MCode = newMemory(sproc.MCode.ID, sproc.MCode.mSize)
	p.MData = newMemory(sproc.MData.ID, sproc.MData.mSize)
	p.S = newStack(sproc.S.ID, sproc.S.mSize)

	for i := 0; i < sproc.MCode.mSize; i++ {
		p.MCode.mem[i] = sproc.MCode.mem[i]
	}
	for i := 0; i < sproc.MData.mSize; i++ {
		p.MData.mem[i] = sproc.MData.mem[i]
	}
	for i := 0; i < sproc.S.mSize; i++ {
		p.S.mem[i] = sproc.S.mem[i]
	}
}

// copySysfun 系统函数复制构造（way=1时不分配MData）
func copySysfun(pid, way int, sproc *Process) *Process {
	p := &Process{}
	*p = *sproc
	p.ID = pid

	p.MCode = newMemory(sproc.MCode.ID, sproc.MCode.mSize)
	if way != 1 {
		p.MData = newMemory(sproc.MData.ID, sproc.MData.mSize)
	} else {
		p.MData = &Memory{ID: sproc.MData.ID, mSize: 0}
	}
	p.S = newStack(sproc.S.ID, sproc.S.mSize)

	for i := 0; i < sproc.MCode.mSize; i++ {
		p.MCode.mem[i] = sproc.MCode.mem[i]
	}
	if way != 1 {
		for i := 0; i < sproc.MData.mSize; i++ {
			p.MData.mem[i] = sproc.MData.mem[i]
		}
	}
	for i := 0; i < sproc.S.mSize; i++ {
		p.S.mem[i] = sproc.S.mem[i]
	}

	return p
}

// writeBack 将 pcb 内容写回进程
func (p *Process) writeBack(block *PCB) {
	p.state = block.state
	p.exetime = block.exetime
}

// PCB 进程控制块
type PCB struct {
	ID          int
	state       int
	priority    int
	exetime     int
	callerID    int
	communicate int
	interrupt   int
	valid       int // 0 free, 1 used
	pptr        *Process
	next        *PCB
}

func newBlankPCB() *PCB {
	return &PCB{ID: -1, callerID: -1}
}

// link 将进程链接到 pcb
func (pcb *PCB) link(proc *Process) {
	pcb.ID = proc.ID
	pcb.state = proc.ID
	pcb.priority = proc.priority
	pcb.exetime = proc.exetime
	pcb.callerID = proc.callerID
	pcb.communicate = proc.communicate
	pcb.interrupt = proc.interrupt
	pcb.pptr = proc
	pcb.next = nil
}

// update 更新 pcb 属性
func (pcb *PCB) updateProc(pptr *Process) {
	pcb.ID = pptr.ID
	pcb.state = pptr.ID
	pcb.priority = pptr.priority
	pcb.exetime = pptr.exetime
	pcb.callerID = pptr.callerID
	pcb.communicate = pptr.communicate
	pcb.interrupt = pptr.interrupt
}

// release 释放 pcb
func (pcb *PCB) release() {
	pcb.ID = 0
	pcb.state = 0
	pcb.priority = 0
	pcb.callerID = 0
	pcb.communicate = 0
	pcb.interrupt = 0
	pcb.valid = 0
	pcb.pptr = nil
	pcb.next = nil
}

// PCBList 进程控制块链表（动态链表）
type PCBList struct {
	ID    int
	head  *PCB // blank head
	tail  *PCB
	blank PCB
	len   int
}

func newPCBList(id int) *PCBList {
	l := &PCBList{ID: id}
	l.head = &l.blank
	l.tail = &l.blank
	l.len = 0
	return l
}

// insertToHead 插入到链表头部
func (l *PCBList) insertToHead(pblock *PCB) int {
	pblock.next = l.head.next
	l.head.next = pblock
	l.len++
	return 0
}

// enQueue 插入到链表尾部
func (l *PCBList) enQueue(pblock *PCB) int {
	l.tail.next = pblock
	l.tail = l.tail.next
	pblock.next = nil
	l.len++
	return 0
}

// deQueue 从链表头部取出
func (l *PCBList) deQueue(pblock **PCB) int {
	if l.len == 0 {
		RES = 130
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d pcb list:%d ", l.ID, l.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	*pblock = l.head.next
	l.head.next = l.head.next.next
	if l.len == 1 {
		l.tail = l.head
	}
	l.len--
	return 0
}

// IDGet 根据 ID 获取 pcb
func (l *PCBList) IDGet(pblock **PCB, id int) int {
	if l.len == 0 {
		RES = 131
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d pcb's process ID:%d ", l.ID, (*pblock).ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	current := l.head.next
	pri := l.head
	for current != nil && current.ID != id {
		pri = current
		current = current.next
	}
	if current == nil {
		if displayMode != 2 {
			printf("not find pcb!\n")
		}
		return -1
	}
	*pblock = current
	pri.next = current.next
	if current == l.tail {
		l.tail = pri
	}
	l.len--
	return 0
}

// CommunicateGet 根据 communicate 获取 pcb
func (l *PCBList) CommunicateGet(pblock **PCB, com int) int {
	if l.len == 0 {
		RES = 131
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d ", l.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	current := l.head.next
	pri := l.head
	for current != nil && current.communicate != com {
		pri = current
		current = current.next
	}
	if current == nil {
		RES = 132
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d searched pcb communicate:%d ", l.ID, com)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	*pblock = current
	pri.next = current.next
	if current == l.tail {
		l.tail = pri
	}
	l.len--
	return 0
}
