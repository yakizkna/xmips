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
	{198, 3, "interpreter::exer: illegal access to input register #14 (only INT 14 can read it)"},
	{200, 3, "fsdev::open: path escape / open failed"},
	{201, 3, "fsdev::read: invalid fd"},
	{202, 3, "fsdev::write: invalid fd / write failed"},
	{203, 3, "fsdev::close: invalid fd"},
	{204, 3, "fsdev::net: socket connect / io error"},
	{205, 1, "dispatcher::swap2: a file/socket system call"},
}

var sysFunNumber = 5 // 系统函数数量（索引 0-4 → INT 10-14）

var sysFunTable = [10]string{
	"INT_10.scp", // index 0 → INT 10: 冒泡排序（.scp 系统函数）
	"",           // index 1 → INT 11: (保留)
	"",           // index 2 → INT 12: (保留)
	"",           // index 3 → INT 13: 输出字符串（dispatcher 原生，无 .scp）
	"",           // index 4 → INT 14: 输入缓冲（dispatcher 原生，无 .scp）
	"",
	"",
	"",
	"",
	"",
}

// runDir xmips 运行根目录（可执行文件所在目录）。config.ini、sysfun/file 目录、
// 虚拟磁盘 diskRoot 均以此为基准，使 `xmips XXX.cupa` 可从任意目录直接运行
var runDir = ""

var sysPath = [2]string{
	"./sysfun/",
	"./file/",
}

var runList = "run.list"

var displayMode = 0 // 0 hide the content of the process loaded to system; 1 show all information

var reportLevel = 0 // 0 report Error, 1 report Normal and Error, 2 report all

var delayMode = 0 // 0 不延时, 1 延时

var updateSysfun = 0 // 碳由历史保留：普通运行直接 Load .co；重建系统库请用 `./xmips update`

// config.ini 可配置的系统参数（默认值与代码原硬编码一致）
var codeSize = 800       // 用户进程代码区长度（字）
var dataSize = 200       // 用户进程数据区长度（字）
var stackSize = 50       // 用户进程栈区长度（字）
var sysfunCodeSize = 800 // 系统函数代码区长度（字）
var sysfunDataSize = 10  // 系统函数数据区长度（字，运行时指向调用者数据段）
var sysfunStackSize = 40 // 系统函数栈区长度（字）
var cycleTimes = 30      // 时间片大小：每个时间片最多执行的指令数
var pcbNum = 40          // PCB 数量：系统最大并发进程数

// wordBits 机器字长（位）：32 或 64。决定通用寄存器与内存中整数的有效位数与进位行为
var wordBits = 64

// diskRoot 为 config.ini 中的预留配置项。文件系统 open 目前直接读写真实宿主文件
// （相对当前工作目录或绝对路径），不再强制限定在 disk/ 目录内
var diskRoot = "disk"

// sockTimeout socket connect/write 超时（毫秒）
var sockTimeout = 2000

// dumpEnabled 是否开启指令执行 dump 日志（config.ini 中  dump=1 开启）
var dumpEnabled = 0

// singleProc 是否为单进程模式（./xmips xxx.cupa 直跑，或 run.list 仅 1 个文件）。
// 单进程模式下网络 read 采用阻塞式，多进程模式下用非阻塞，避免卡死时间片轮转
var singleProc = false

// truncWord 将任意 int 截断为当前机器字长，返回带符号表示
// - 64 位：Go int 本身即 64 位带符号，直接返回
// - 32 位：取低 32 位并按 32 位带符号解释（等价于 C 的 int 溢出回绕）
func truncWord(v int) int {
	if wordBits == 32 {
		return int(int32(uint32(v)))
	}
	return v
}

// 位移与循环移位，均按当前机器字长（wordBits）运算
func shiftL(v, c int) int {
	if wordBits == 32 {
		return int(int32(uint32(v) << uint(c&31)))
	}
	return int(uint64(v) << uint(c&63))
}

func shiftR(v, c int) int {
	if wordBits == 32 {
		return int(int32(uint32(v) >> uint(c&31)))
	}
	return int(uint64(v) >> uint(c&63))
}

func rotL(v, c int) int {
	if wordBits == 32 {
		k := uint(c & 31)
		u := uint32(v)
		if k == 0 {
			return int(int32(u))
		}
		return int(int32(u<<k | u>>(32-k)))
	}
	k := uint(c & 63)
	u := uint64(v)
	if k == 0 {
		return int(u)
	}
	return int(u<<k | u>>(64-k))
}

func rotR(v, c int) int {
	if wordBits == 32 {
		k := uint(c & 31)
		u := uint32(v)
		if k == 0 {
			return int(int32(u))
		}
		return int(int32(u>>k | u<<(32-k)))
	}
	k := uint(c & 63)
	u := uint64(v)
	if k == 0 {
		return int(u)
	}
	return int(u>>k | u<<(64-k))
}

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
