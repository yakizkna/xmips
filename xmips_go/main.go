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

func main() {
	if config() != 0 {
		fmt.Println("system parameters config error!")
	}

	// 构建硬件/系统对象
	im := newInterpreter(101, 30, 10, 20)
	e := newEditor(103, 200)
	a := newAssembler(102)
	os_ := newDispatcher(105, 10, 30, 40)
	sys := newStorage(106, sysPath[0])
	disk := newStorage(107, sysPath[1])

	//************************************************************************************************************
	// 初始化系统函数
	{
		pptr := make([]*Process, sysFunNumber)
		// 进程模板：代码段大、数据段小、高优先级
		tp := newProcess(0, 100, 10, 40, 0)

		for i := 0; i < sysFunNumber; i++ {
			pptr[i] = newProcess(0, 100, 10, 40, 0)
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
	{
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
			return
		}
		name := scanner.Text()

		id := 50 // 用户进程起始 ID
		for name != "end" && name != "END" {
			pfr, ret2 := disk.getFile(name, 0)
			if ret2 == 0 {
				pptr := newProcess(0, 100, 40, 50, 1)
				pptr.ID = id
				fmt.Printf("process ID:%d\n", pptr.ID)

				if e.ASM(pfr, a, name) != -1 {
					Load(pptr, name)
					os_.loader(pptr)
				}
				disk.releaseFile(pfr)
			}
			id++

			if !scanner.Scan() {
				break
			}
			name = scanner.Text()
		}
		run.Close()
	}
	//*************************************************************************************************************

	// 调度执行
	os_.swap2(im)

	//*************************************************************************************************************
	// 显示已结束进程的数据段
	{
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
