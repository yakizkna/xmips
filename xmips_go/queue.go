package main

// Queue 循环队列，单元为进程指针
type Queue struct {
	ID           int
	length       int
	head         int
	tail         int
	memberNumber int
	procQueue    []*Process
}

func newQueue(key, length int) *Queue {
	return &Queue{
		ID:        key,
		length:    length,
		head:      0,
		tail:      0,
		procQueue: make([]*Process, length),
	}
}

func (q *Queue) enQueue(pptr *Process) int {
	if (q.tail+1)%q.length == q.head {
		RES = 60
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d process ID:%d queue size:%d member's number:%d ", q.ID, pptr.ID, q.length-1, q.memberNumber)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	q.procQueue[q.tail] = pptr
	q.tail = (q.tail + 1) % q.length
	q.memberNumber++
	return 0
}

func (q *Queue) deQueue(pptr **Process) int {
	if q.head == q.tail {
		RES = 61
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d ", q.ID)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	*pptr = q.procQueue[q.head]
	q.head = (q.head + 1) % q.length
	q.memberNumber--
	return 0
}

// List 静态列表
type List struct {
	ID     int
	length int
	listp  []*Process
}

func newList(key, length int) *List {
	return &List{
		ID:     key,
		length: length,
		listp:  make([]*Process, length),
	}
}

func (l *List) put(t int, pptr *Process) int {
	if t < 0 || t >= l.length {
		return -1
	}
	l.listp[t] = pptr
	return 0
}

func (l *List) get(t int, pptr **Process) int {
	if t < 0 || t >= l.length {
		return -1
	}
	if l.listp[t] == nil {
		return -1
	}
	*pptr = l.listp[t]
	return 0
}

// SysList 系统函数列表，继承 List
type SysList struct {
	List
}

func newSysList(key, length int) *SysList {
	return &SysList{List: *newList(key, length)}
}

// copy 根据系统调用号 t 从表中复制系统函数进程
func (sl *SysList) copy(t int, dpptr **Process) int {
	if t < 0 || t >= sl.length {
		return -1
	}
	if sl.listp[t] == nil {
		return -1
	}
	*dpptr = copySysfun(t, 1, sl.listp[t])
	RES = 90
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d system function ID:%d ", sl.ID, t)
		systemChecker.check(RES, sysLog[:])
	}
	return 0
}

// copyProc 从源进程复制（不使用）
func (sl *SysList) copyProc(dpptr **Process, spptr *Process) int {
	*dpptr = copySysfun(0, 1, spptr)
	return 0
}
