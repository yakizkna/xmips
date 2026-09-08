package main

// 全局变量，对应 global.cpp 中的定义

var RES int // value changes when some events happen

// abendNote 系统运行状态结构
type abendNote struct {
	abendCode   int
	level       int    // Normal 0, Sub condition 1, Warning 2, Error 3
	abendReason string // abend reason
}

// checker 实例
var systemChecker = newChecker(5)

// sysLog 系统状态表
var sysLog = [100]abendNote{
	{0, 0, "SYSTEM: normal end"},
	{1, 3, "userinterface::menu: fail to open file"},
	{2, 3, "userinterface::menu: fail to create file"},
	{3, 3, "SYSTEM: the process hasnt code file in its project file"},
	{10, 3, "memory::read: memory overflow"},
	{11, 3, "memory::write: memory overflow"},
	{12, 3, "editor::editorData: illegal data"},
	{13, 3, "editor::editorCollection: illegal variable name"},
	{14, 3, "editor::editorCollection: illegal opcode"},
	{15, 3, "editor::editorCollection: too many content"},
	{16, 3, "stack::push: stack is full"},
	{17, 3, "stack::pop: stack is empty"},
	{18, 0, "memory::update: update memory size"},
	{19, 3, "editor::editorCollection: error editor num"},
	{20, 1, "interpreter::exer: interrupt occured"},
	{21, 3, "interpreter::exer: interpreter overflow"},
	{22, 1, "interpreter::exer: interpreter reaches cycleTimes"},
	{23, 0, "interpreter::exer: process execute completed"},
	{24, 3, "interpreter::exer: property set illegal"},
	{51, 3, "assembler::trans2: '@' use error"},
	{52, 3, "assembler::trans2: undeclared indentifier"},
	{53, 3, "assembler::trans2: '&' use error"},
	{54, 3, "assembler::trans2: register used in indirect addressing"},
	{55, 3, "assembler::trans2: '#' use error"},
	{56, 3, "assembler::trans2: other char occur before immediate num"},
	{57, 3, "assembler::trans2: '!' use error"},
	{58, 3, "assembler::trans2: over the max general register number(14)"},
	{60, 3, "queue::enQueue: queue is full"},
	{61, 2, "queue::deQueue: queue is empty"},
	{70, 0, "scheduler::swap: process swaps to run"},
	{71, 0, "scheduler::swap: process swaps from run to ready"},
	{72, 0, "scheduler::swap: process swaps to wait"},
	{73, 1, "scheduler::swap: a system call happened"},
	{74, 1, "scheduler::swap: a system function completed"},
	{75, 0, "scheduler::swap: process swaps from wait to ready"},
	{76, 0, "scheduler::swap: a system function swaps to ready"},
	{77, 0, "scheduler::swap: a process finished"},
	{78, 0, "scheduler::swap: a process failed"},
	{79, 1, "scheduler::swap: swap is completed"},
	{80, 3, "storage::getFile: visit way error"},
	{81, 3, "storage::getFile: file open failed"},
	{82, 3, "storage::releaseFile: file pointer is null"},
	{90, 1, "sysList::copy: a system function is copied"},
	{100, 1, "scheduler::swap: read system share"},
	{101, 1, "scheduler::swap: write system share"},
	{130, 2, "pcbList::enQueue: list is empty"},
	{131, 3, "pcbList::get: list is empty"},
	{132, 3, "pcbList::get: pcb not found"},
	{170, 0, "dispatcher::swap2: process swaps to run"},
	{171, 0, "dispatcher::swap2: process swaps from run to ready"},
	{172, 0, "dispatcher::swap2: process swaps to wait"},
	{173, 1, "dispatcher::swap2: a system call happened"},
	{174, 1, "dispatcher::swap2: a system function completed"},
	{175, 0, "dispatcher::swap2: process swaps from wait to ready"},
	{176, 0, "dispatcher::swap2: a system function swaps to ready"},
	{177, 0, "dispatcher::swap2: a process finished"},
	{178, 0, "dispatcher::swap2: a process failed"},
	{179, 1, "dispatcher::swap2: swap is completed"},
	{180, 3, "dispatcher::swap2: access violation to system share region"},
	{181, 1, "dispatcher::swap2: read system share"},
	{182, 1, "dispatcher::swap2: write system share"},
	{183, 3, "dispatcher::pcbManagement: no pcb can use now"},
	{184, 3, "dispatcher::loader: pcb dispatch failed"},
	{185, 0, "dispatcher::swap2: a process suspended"},
	{186, 0, "dispatcher::swap2: a process woken"},
	{187, 3, "dispatcher::swap2: access violation to wake up a process"},
	{188, 1, "dispatcher::swap2: insert back to the head of ready queue"},
	{189, 1, "dispatcher::swap2: interruption ban checked"},
	{190, 3, "dispatcher::swap2: over the range of wait queue numbers"},
	{191, 3, "editor::editorFromFile: initial value more than declared"},
	{192, 3, "editor::editorFromFile: illegal declaration"},
	{193, 3, "editor::editorFromFile: tag repeated"},
	{194, 3, "editor::editorFromFile: tag not find"},
	{195, 3, "editor::ASM: open file failed"},
	{196, 3, "editor::editorFromFile: indentifier repeated"},
	{197, 3, "editor::editorFromFile: immediate num cant be 1st operand in double operands instruction"},
}

var sysFunNumber = 1 // 当前使用的系统函数数量（测试用1）

var sysFunTable = [10]string{
	"bubble.scp",
	"sum.scp",
	"read.scp",
	"write.scp",
	"suspend.scp",
	"wakeup.scp",
	"P.scp",
	"V.scp",
	"semaphore.scp",
	"pause.scp",
}

var sysPath = [2]string{
	"./sysfun/",
	"./file/",
}

var runList = "run.list"

var displayMode = 0 // 0 hide the content of the process loaded to system; 1 show all information

var reportLevel = 0 // 0 report Error, 1 report Normal and Error, 2 report all

var delayMode = 0 // 0 不延时, 1 延时

var updateSysfun = 1 // 0 不重新汇编系统函数, 1 重新汇编

// procDelay 在 interpreter 的 load/store/push/pop 等操作中使用
func procDelay(mode int, timeMs int) {
	if mode == 1 {
		i := 0
		for i < timeMs {
			j := 0
			for j < timeMs {
				j++
			}
			i++
		}
	}
}
