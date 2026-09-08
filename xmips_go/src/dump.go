package main

import (
	"fmt"
	"os"
	"strings"
)

// dump 日志：config.ini 中 dump=1 时开启，把每个时钟周期（每条指令执行并写回后）
// 发生变化的寄存器（#0~#16）与数据段内存单元输出到运行目录下的 dump.log。

var dumpFile *os.File
var dumpPrevRegs []int  // 上一周期寄存器快照，用于求差
var dumpPrevData []int  // 上一周期数据段快照

// initDump 开启 dump 日志（在运行目录 log/ 下创建 dump.log，覆盖旧文件）
func initDump() {
	if dumpEnabled == 0 {
		return
	}
	dir := runDir + "/log"
	_ = os.MkdirAll(dir, 0o755)
	f, err := os.Create(dir + "/dump.log")
	if err != nil {
		return
	}
	dumpFile = f
	dumpPrevRegs = nil
	dumpPrevData = nil
}

// closeDump 关闭 dump 日志文件
func closeDump() {
	if dumpFile != nil {
		dumpFile.Close()
		dumpFile = nil
	}
}

// doDump 在每个时钟周期后记录与上一周期相比有变化的寄存器与数据段内存
func doDump(im *Interpreter, proc *Process) {
	if dumpFile == nil {
		return
	}
	if len(dumpPrevRegs) < 17 {
		dumpPrevRegs = make([]int, 17)
	}
	if len(dumpPrevData) < proc.MData.mSize {
		dumpPrevData = make([]int, proc.MData.mSize)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[cycle=%d PID=%d PC=%d]", proc.exetime, proc.ID, im.PC))
	changed := false

	// 寄存器 #0..#16 变化
	for i := 0; i <= 16; i++ {
		if dumpPrevRegs[i] != im.GM.mem[i] {
			fmt.Fprintf(&sb, " #%d=%d", i, im.GM.mem[i])
			dumpPrevRegs[i] = im.GM.mem[i]
			changed = true
		}
	}

	// 数据段内存变化
	for i := 0; i < proc.MData.mSize; i++ {
		if dumpPrevData[i] != proc.MData.mem[i] {
			fmt.Fprintf(&sb, " m%d=%d", i, proc.MData.mem[i])
			dumpPrevData[i] = proc.MData.mem[i]
			changed = true
		}
	}

	if changed {
		fmt.Fprintln(dumpFile, sb.String())
	}
}