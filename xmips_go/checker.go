package main

import "fmt"

// Checker 状态检查器，对应 C++ 的 checker 类
type Checker struct {
	ID int
}

func newChecker(key int) *Checker {
	return &Checker{ID: key}
}

// showLevel 判断是否显示该级别的报告
func (c *Checker) showLevel(k int, abendTable []abendNote) bool {
	i := 0
	for k != abendTable[i].abendCode {
		i++
	}
	rl := abendTable[i].level
	if reportLevel == 0 && rl != 2 && rl != 3 {
		return false
	}
	if reportLevel == 1 && rl == 1 {
		return false
	}
	return true
}

// check 检查 RES 并输出状态信息
func (c *Checker) check(k int, abendTable []abendNote) {
	i := 0
	for k != abendTable[i].abendCode {
		i++
	}
	levels := []string{"Normal", "Normal(Sub)", "Warning", "Error"}
	fmt.Printf(" %s, RES=%d, %s\n", abendTable[i].abendReason, abendTable[i].abendCode, levels[abendTable[i].level])
}

// printf 包装 fmt.Printf，方便统一
func printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
