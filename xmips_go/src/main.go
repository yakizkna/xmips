package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// config 读取系统配置文件 config.ini
func config() int {
	f, err := os.Open("config.ini")
	if err != nil {
		return -1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		para := scanner.Text()
		if para == "end" || para == "END" {
			break
		}

		// 解析 "key=value" 格式，无空格
		idx := strings.Index(para, "=")
		if idx < 0 {
			continue
		}
		key := para[:idx]
		val := para[idx+1:]
		r := atoiSafe(val)

		switch key {
		case "delayMode":
			delayMode = r
		case "displayMode":
			displayMode = r
		case "reportLevel":
			reportLevel = r
		case "updateSysfun":
			updateSysfun = r
		case "codeSize":
			if r > 0 {
				codeSize = r
			}
		case "dataSize":
			if r > 0 {
				dataSize = r
			}
		case "stackSize":
			if r > 0 {
				stackSize = r
			}
		case "sysfunCodeSize":
			if r > 0 {
				sysfunCodeSize = r
			}
		case "sysfunDataSize":
			if r > 0 {
				sysfunDataSize = r
			}
		case "sysfunStackSize":
			if r > 0 {
				sysfunStackSize = r
			}
		case "cycleTimes":
			if r > 0 {
				cycleTimes = r
			}
		case "pcbNum":
			if r > 0 {
				pcbNum = r
			}
		case "bitMode":
			if r == 32 || r == 64 {
				wordBits = r
			}
		case "diskRoot":
			// 字符串值需剥离行内注释与两端空白
			v := strings.SplitN(val, ";", 2)[0]
			v = strings.TrimSpace(v)
			if v != "" {
				diskRoot = v
			}
		case "sockTimeout":
			if r > 0 {
				sockTimeout = r
			}
		}
	}
	return 0
}

func atoiSafe(s string) int {
	n := 0
	negative := false
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	if negative {
		n = -n
	}
	return n
}

// osOpen 直接按完整路径打开文件（用于命令行指定的任意路径程序）
func osOpen(path string) (*os.File, int) {
	f, err := os.Open(path)
	if err != nil {
		return nil, -1
	}
	return f, 0
}

func main() {
	if config() != 0 {
		fmt.Println("system parameters config error!")
	}

	// 构建硬件/系统对象
	im := newInterpreter(101, cycleTimes, 10, 20)
	e := newEditor(103, codeSize, dataSize)
	a := newAssembler(102)
	os_ := newDispatcher(105, 10, 30, pcbNum)
	os_.FS = newFsDev() // 注入文件系统/套接字子系统
	sys := newStorage(106, sysPath[0])
	disk := newStorage(107, sysPath[1])

	//************************************************************************************************************
	// 初始化系统函数
	{
		pptr := make([]*Process, sysFunNumber)
		// 进程模板：代码段大、数据段小、高优先级
		tp := newProcess(0, sysfunCodeSize, sysfunDataSize, sysfunStackSize, 0)

		for i := 0; i < sysFunNumber; i++ {
			if sysFunTable[i] == "" {
				continue
			}
			pptr[i] = newProcess(0, sysfunCodeSize, sysfunDataSize, sysfunStackSize, 0)
			pptr[i].copy(tp)
			pptr[i].ID = i

			if updateSysfun == 1 { // 重新汇编系统函数
				fp, ret := sys.getFile(sysFunTable[i], 0)
				if ret == 0 {
					if e.ASM(fp, a, sysFunTable[i]) != -1 {
						Load(pptr[i], sysFunTable[i])
						os_.SysCall.put(i, pptr[i])
					}
					sys.releaseFile(fp)
				}
			} else {
				Load(pptr[i], sysFunTable[i])
				os_.SysCall.put(i, pptr[i])
			}
		}
	}
	//************************************************************************************************************

	//************************************************************************************************************
	// 加载用户程序
	// 支持两种方式：
	//   1) ./xmips            -> 从 run.list 读取要运行的程序
	//   2) ./xmips prog.cupa  -> 直接用命令行参数指定程序，忽略 run.list
	{
		// 从 file 目录加载用户程序（run.list 场景）
		loadFromDisk := func(name string, id int) int {
			pfr, ret2 := disk.getFile(name, 0)
			if ret2 != 0 {
				fmt.Printf("cannot open %s\n", name)
				return ret2
			}
			defer disk.releaseFile(pfr)
			pptr := newProcess(0, codeSize, dataSize, stackSize, 1)
			pptr.ID = id
			if displayMode != 2 {
				fmt.Printf("process ID:%d\n", pptr.ID)
			}
			if e.ASM(pfr, a, name) != -1 {
				Load(pptr, name)
				os_.loader(pptr)
			}
			return 0
		}

		if len(os.Args) >= 2 { // 命令行指定程序（可直接传 file 目录内文件名或任意路径）
			singleProc = true // 单进程模式：网络 read 用阻塞式
			prog := os.Args[1]
			// 若参数带路径分隔符或文件不存在于 file 目录，则当作直接路径打开
			var pfr *os.File
			var ret int
			if strings.ContainsAny(prog, `/\`) {
				pfr, ret = osOpen(prog)
			} else {
				pfr, ret = disk.getFile(prog, 0)
			}
			if ret != 0 {
				fmt.Printf("cannot open %s\n", prog)
				return
			}
			pptr := newProcess(0, codeSize, dataSize, stackSize, 1)
			pptr.ID = 50
			if displayMode != 2 {
				fmt.Printf("process ID:%d\n", pptr.ID)
			}
			if e.ASM(pfr, a, prog) != -1 {
				Load(pptr, prog)
				os_.loader(pptr)
			}
			pfr.Close()
		} else { // 从 run.list 读取
			run, ret := disk.getFile(runList, 0)
			if ret != 0 {
				fmt.Println("cannot open run.list")
				return
			}

			scanner := bufio.NewScanner(run)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
			scanner.Split(bufio.ScanWords)

			// 读取第一个程序名
			if !scanner.Scan() {
				run.Close()
				return
			}
			name := scanner.Text()

			id := 50 // 用户进程起始 ID
			loaded := 0
			for name != "end" && name != "END" {
				loadFromDisk(name, id)
				id++
				loaded++
				if !scanner.Scan() {
					break
				}
				name = scanner.Text()
			}
			singleProc = loaded <= 1 // 仅 1 个文件视为单进程模式
			run.Close()
		}
	}
	//*************************************************************************************************************

	// 调度执行
	os_.swap2(im)

	//*************************************************************************************************************
	// 显示已结束进程的数据段
	{
		if displayMode == 2 {
			// displayMode=2：不显示任何额外内容，直接结束
			return
		}
		var pptr *Process
		r := os_.Finished.deQueue(&pptr)

		fmt.Print("\ndisplay data?(y/n): ")
		reader := bufio.NewReader(os.Stdin)
		c, _ := reader.ReadByte()
		fmt.Println()

		for r != -1 {
			fmt.Printf("process ID:%d\n", pptr.ID)
			if c == 'y' || c == 'Y' {
				fmt.Printf("memory ID:%d\n", pptr.MData.ID)
				for i := 0; i < pptr.MData.mSize; i++ {
					fmt.Printf("%-4d%d\n", i, pptr.MData.read(i))
				}
			}
			fmt.Printf("execute times:%d\n", pptr.exetime)
			r = os_.Finished.deQueue(&pptr)
		}
	}

	fmt.Println("\nPress any key to exit.")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
}
