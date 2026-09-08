package main

// Memory 存储器类，对应 C++ 的 memory 类
type Memory struct {
	ID    int
	mSize int
	mem   []int
}

func newMemory(id int, m int) *Memory {
	return &Memory{
		ID:    id,
		mSize: m,
		mem:   make([]int, m),
	}
}

func (m *Memory) update(size int) {
	m.mSize = size
	m.mem = make([]int, size)
	RES = 18
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d new size:%d ", m.ID, size)
		systemChecker.check(RES, sysLog[:])
	}
}

func (m *Memory) read(i int) int {
	if i >= m.mSize || i < 0 {
		RES = 10
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d memory size:%d ,read address:%d ", m.ID, m.mSize, i)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	return m.mem[i]
}

func (m *Memory) write(i, k int) int {
	if i >= m.mSize || i < 0 {
		RES = 11
		if systemChecker.showLevel(RES, sysLog[:]) {
			printf("D%-5d memory size:%d ,write address:%d ", m.ID, m.mSize, i)
			systemChecker.check(RES, sysLog[:])
		}
		return -1
	}
	m.mem[i] = k
	return 0
}

// Stack 栈类，对应 C++ 的 stack 类（继承 memory）
type Stack struct {
	Memory
	SP int
}

func newStack(id int, m int) *Stack {
	return &Stack{
		Memory: Memory{ID: id, mSize: m, mem: make([]int, m)},
		SP:     0,
	}
}

func (s *Stack) push(i int) int {
	if s.SP < s.mSize {
		s.write(s.SP, i)
		s.SP++
		return 0
	}
	RES = 16
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d stack size:%d, SP:%d ", s.ID, s.mSize, s.SP)
		systemChecker.check(RES, sysLog[:])
	}
	return -1
}

func (s *Stack) pop() int {
	if s.SP > 0 {
		s.SP--
		return s.read(s.SP)
	}
	RES = 17
	if systemChecker.showLevel(RES, sysLog[:]) {
		printf("D%-5d stack size:%d, SP:%d ", s.ID, s.mSize, s.SP)
		systemChecker.check(RES, sysLog[:])
	}
	return -1
}
