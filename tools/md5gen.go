// md5gen 为 Xmips(Go 版) 生成 MD5.cupa（ABC 汇编）
// 用法: cd tools && GO111MODULE=off go run md5gen.go
// 说明: bitMode 需设为 32；输入消息 ≤448 字节；输出 32 位小写十六进制 MD5
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// 常量辅助
func uint32Const(f float64) int {
	return int(int32(uint32(f)))
}

// 寄存器分配
const (
	A = 0 // #0
	Bv = 1 // #1
	C = 2 // #2
	D = 3 // #3
	ta = 6 // #6 临时
	tb = 7 // #7 临时
	tc = 8 // #8 临时
	td = 9 // #9 临时
)

// bufAddr BUF 数组在数据区中的起始地址（偏移）。
// 注意：Xmips 中 `MOV #r BUF` 读取的是 BUF 处的"内容"而非其地址，
// 因此取地址须用字面量（与 GREET.cupa 硬编码地址的惯例一致）。
const bufAddr = 0

func main() {
	// 每步参数
	type step struct {
		f func() []string // 生成 F/G/H/I 计算 -> 结果放 #6
		g int
		k int
		s int
	}
	var steps []step

	S1 := []int{7, 12, 17, 22}
	S2 := []int{5, 9, 14, 20}
	S3 := []int{4, 11, 16, 23}
	S4 := []int{6, 10, 15, 21}
	var S []int
	for r := 0; r < 4; r++ {
		for i := 0; i < 16; i++ {
			var tbl []int
			switch r {
			case 0:
				tbl = S1
			case 1:
				tbl = S2
			case 2:
				tbl = S3
			case 3:
				tbl = S4
			}
			S = append(S, tbl[i%4])
		}
	}
	K := make([]int, 64)
	for i := range K {
		K[i] = uint32Const(math.Abs(math.Sin(float64(i+1))) * 4294967296.0)
	}

	Fgen := func() []string {
		// F = (B&C) | (~B & D)
		return []string{
			fmt.Sprintf("MOV #%d #%d", ta, Bv),
			fmt.Sprintf("AND #%d #%d", ta, C),
			fmt.Sprintf("MOV #%d #%d", tb, Bv),
			fmt.Sprintf("NEG #%d", tb),
			fmt.Sprintf("SUB #%d 1", tb),
			fmt.Sprintf("AND #%d #%d", tb, D),
			fmt.Sprintf("OR #%d #%d", ta, tb),
		}
	}
	Ggen := func() []string {
		// G = (B&D) | (C & ~D)
		return []string{
			fmt.Sprintf("MOV #%d #%d", ta, Bv),
			fmt.Sprintf("AND #%d #%d", ta, D),
			fmt.Sprintf("MOV #%d #%d", tb, D),
			fmt.Sprintf("NEG #%d", tb),
			fmt.Sprintf("SUB #%d 1", tb),
			fmt.Sprintf("AND #%d #%d", tb, C),
			fmt.Sprintf("OR #%d #%d", ta, tb),
		}
	}
	Hgen := func() []string {
		// H = B^C^D
		return []string{
			fmt.Sprintf("MOV #%d #%d", ta, Bv),
			fmt.Sprintf("XOR #%d #%d", ta, C),
			fmt.Sprintf("XOR #%d #%d", ta, D),
		}
	}
	Igen := func() []string {
		// I = C ^ (B | ~D)
		return []string{
			fmt.Sprintf("MOV #%d #%d", tb, D),
			fmt.Sprintf("NEG #%d", tb),
			fmt.Sprintf("SUB #%d 1", tb),
			fmt.Sprintf("MOV #%d #%d", ta, Bv),
			fmt.Sprintf("OR #%d #%d", ta, tb),
			fmt.Sprintf("XOR #%d #%d", ta, C),
		}
	}

	for i := 0; i < 64; i++ {
		var fgen func() []string
		var g int
		if i < 16 {
			fgen = Fgen
			g = i
		} else if i < 32 {
			fgen = Ggen
			g = (5*i + 1) & 15
		} else if i < 48 {
			fgen = Hgen
			g = (3*i + 5) & 15
		} else {
			fgen = Igen
			g = (7 * i) & 15
		}
		steps = append(steps, step{fgen, g, K[i], S[i]})
	}

	var out []string
	emit := func(s string) {
		// 剥离行内 ; 注释（.cupa 仅支持 ~ 注释）
		if i := strings.Index(s, ";"); i >= 0 {
			s = s[:i]
		}
		out = append(out, strings.TrimRight(s, " "))
	}
	emitf := func(format string, a ...interface{}) { emit(fmt.Sprintf(format, a...)) }

	// ---------- 头部声明 ----------
	emit("~ MD5 计算器（Xmips ABC 汇编，需 bitMode=32）")
	emit("~ 输入: 标准输入（≤448 字节）；输出: 32 位小写十六进制 MD5")
	emit("DIM BUF D512")
	emit("DIM LEN D1")
	emit("DIM NBLK D1")
	emit("DIM BPTR D1")
	// 每块处理时保存/累加初始 chaining 值（MD5 要求块处理结束后累加 IV）
	emit("DIM A0 D1")
	emit("DIM B0 D1")
	emit("DIM C0 D1")
	emit("DIM D0 D1")
	for i := 0; i < 16; i++ {
		emitf("DIM M%d D1", i)
	}
	emit("")

	// ---------- 初始化 A B C D ----------
	emit("~ 初始值: A=0x67452301 B=0xefcdab89 C=0x98badcfe D=0x10325476")
	emitf("MOV #%d 1732584193", A)
	emitf("MOV #%d -271733879", Bv)
	emitf("MOV #%d -1732584194", C)
	emitf("MOV #%d 271733878", D)
	emit("MOV LEN 0")
	emit("")

	// ---------- 读入消息到 BUF ----------
	emit(": READQ")
	emitf("MOV #%d LEN", ta)
	emitf("MOV #%d 448", tb)
	emit("CMP " + rg(ta) + " " + rg(tb))
	emit("JB READ1          ; LEN < 448 继续读")
	emit("JMP PAD           ; 已到 448，停止")
	emit(": READ1")
	emitf("MOV #%d %d", ta, bufAddr)
	emitf("MOV #%d LEN", tb)
	emit("ADD " + rg(ta) + " " + rg(tb))
	emit("MOV #11 " + rg(ta))
	emit("MOV #12 1")
	emit("MOV #10 14")
	emit("INT")
	emitf("MOV #%d #13      ; 读到的标志: 1=有字符 0=EOF", ta)
	emitf("MOV #%d 0", tb)
	emit("CMP " + rg(ta) + " " + rg(tb))
	emit("JE PAD            ; EOF -> 填充")
	emitf("MOV #%d LEN", ta)
	emitf("ADD #%d 1", ta)
	emit("MOV LEN " + rg(ta))
	emit("JMP READQ")
	emit("")

	// ---------- PAD 填充 ----------
	emit(": PAD")
	emitf("MOV #%d %d", ta, bufAddr)
	emitf("MOV #%d LEN", tb)
	emit("ADD " + rg(ta) + " " + rg(tb))
	emit("MOV @#6 128       ; BUF[LEN] = 0x80")
	// nl = LEN+1
	emitf("MOV #%d LEN", ta)
	emitf("ADD #%d 1", ta)
	// t = nl & 63
	emitf("MOV #%d #%d", tc, ta)
	emitf("AND #%d 63", tc)
	// z = (56 - t) & 63
	emitf("MOV #%d 56", tb)
	emit("SUB " + rg(tb) + " " + rg(tc))
	emitf("AND #%d 63", tb)
	// 指针 = BUF + nl
	emitf("MOV #%d %d", ta, bufAddr)
	emit("ADD " + rg(ta) + " 0")
	emitf("MOV #%d LEN", tc)
	emit("INC " + rg(tc))
	emit("ADD " + rg(ta) + " " + rg(tc))
	emit("MOV #4 " + rg(ta))
	// 写 z 个 0
	emit(": ZLP")
	emitf("MOV #%d 0", ta)
	emit("CMP " + rg(tb) + " " + rg(ta))
	emit("JE ZDONE")
	emitf("MOV @#4 0")
	emit("INC #4")
	emit("SUB " + rg(tb) + " 1")
	emit("JMP ZLP")
	emit(": ZDONE")
	// bits = LEN * 8
	emitf("MOV #%d LEN", tc)
	emitf("SHL #%d 3", tc)
	// 写 8 字节长度 (低 4 字节，高 4 字节为 0)
	emitf("MOV #%d #%d", td, tc)
	emitf("AND #%d 255", td)
	emit("MOV @#4 " + rg(td))
	emit("INC #4")
	emitf("MOV #%d #%d", td, tc)
	emitf("SHR #%d 8", td)
	emitf("AND #%d 255", td)
	emit("MOV @#4 " + rg(td))
	emit("INC #4")
	emitf("MOV #%d #%d", td, tc)
	emitf("SHR #%d 16", td)
	emitf("AND #%d 255", td)
	emit("MOV @#4 " + rg(td))
	emit("INC #4")
	emitf("MOV #%d #%d", td, tc)
	emitf("SHR #%d 24", td)
	emitf("AND #%d 255", td)
	emit("MOV @#4 " + rg(td))
	emit("INC #4")
	emit("MOV " + rg(td) + " 0")
	for i := 0; i < 4; i++ {
		emit("MOV @#4 " + rg(td))
		emit("INC #4")
	}
	// total = #4 - BUF ; NBLK = total/64
	emitf("MOV #%d %d", tc, bufAddr)
	emit("SUB #4 " + rg(tc))
	emit("SHR #4 6")
	emit("MOV NBLK #4")
	emitf("MOV #%d %d", tc, bufAddr)
	emit("MOV BPTR " + rg(tc))
	emit("JMP BLK")
	emit("")

	// ---------- BLK 分块处理 ----------
	emit(": BLK")
	emitf("MOV #%d NBLK", ta)
	emitf("MOV #%d 0", tb)
	emit("CMP " + rg(ta) + " " + rg(tb))
	emit("JE DONE")
	// 装载 M0..M15
	emitf("MOV #%d BPTR", ta)
	emit("MOV #4 " + rg(ta))
	for w := 0; w < 16; w++ {
		emit("MOV #5 #4")
		emitf("ADD #5 %d", w*4)
		emitf("MOV #%d @#5", ta)
		emit("INC #5")
		emitf("MOV #%d @#5", tb)
		emit("INC #5")
		emitf("MOV #%d @#5", tc)
		emit("INC #5")
		emitf("MOV #%d @#5", td)
		emitf("SHL #%d 8", tb)
		emit("OR " + rg(ta) + " " + rg(tb))
		emitf("SHL #%d 16", tc)
		emit("OR " + rg(ta) + " " + rg(tc))
		emitf("SHL #%d 24", td)
		emit("OR " + rg(ta) + " " + rg(td))
		emitf("MOV M%d #%d", w, ta)
	}
	// 保存当前 chaining 值，块处理后将其累加回去
	emit("MOV A0 " + rg(A))
	emit("MOV B0 " + rg(Bv))
	emit("MOV C0 " + rg(C))
	emit("MOV D0 " + rg(D))
	// 64 步核心
	for _, st := range steps {
		for _, l := range st.f() {
			emit(l)
		}
		// sum = A + f + K + M[g]
		emitf("MOV #%d #%d", tc, A)
		emit("ADD " + rg(tc) + " #6")
		emitf("ADD #%d %d", tc, st.k)
		emitf("ADD #%d M%d", tc, st.g)
		emitf("ROL #%d %d", tc, st.s)
		emit("ADD " + rg(tc) + " #1")
		// 旋转状态
		emitf("MOV #%d #%d", td, D)
		emitf("MOV #%d #%d", D, C)
		emitf("MOV #%d #%d", C, Bv)
		emitf("MOV #%d #%d", Bv, tc)
		emitf("MOV #%d #%d", A, td)
	}
	emit("ADD " + rg(A) + " A0")
		emit("ADD " + rg(Bv) + " B0")
		emit("ADD " + rg(C) + " C0")
		emit("ADD " + rg(D) + " D0")
		// 前进下一块
		emitf("MOV #%d BPTR", ta)
	emitf("ADD #%d 64", ta)
	emit("MOV BPTR " + rg(ta))
	emitf("MOV #%d NBLK", ta)
	emit("SUB " + rg(ta) + " 1")
	emit("MOV NBLK " + rg(ta))
	emit("JMP BLK")
	emit("")

	// ---------- DONE: 输出 ----------
	emit(": DONE")
	words := []int{A, Bv, C, D}
	byteIdx := 0
	for _, w := range words {
		for shift := 0; shift < 32; shift += 8 {
			byteIdx++
			emitf("MOV #%d #%d", tb, w)
			if shift > 0 {
				emitf("SHR #%d %d", tb, shift)
			}
			emitf("AND #%d 255", tb)
			// 高半字节
			emitf("MOV #%d #%d", tc, tb)
			emitf("SHR #%d 4", tc)
			emitf("AND #%d 15", tc)
			emitf("MOV #%d #%d", td, tc)
			emitf("ADD #%d 48", td)
			emit("CMP " + rg(tc) + " 10")
			emitf("JB HI%d", byteIdx)
			emitf("ADD #%d 39", td)
			emitf(": HI%d", byteIdx)
			emit("MOV #13 " + rg(td))
			// 低半字节
			emitf("MOV #%d #%d", tc, tb)
			emitf("AND #%d 15", tc)
			emitf("MOV #%d #%d", td, tc)
			emitf("ADD #%d 48", td)
			emit("CMP " + rg(tc) + " 10")
			emitf("JB LO%d", byteIdx)
			emitf("ADD #%d 39", td)
			emitf(": LO%d", byteIdx)
			emit("MOV #13 " + rg(td))
		}
	}
	emit("MOV #13 10        ; 输出末尾换行")
	emit("END")

	f, err := os.Create("MD5.cupa")
	if err != nil {
		fmt.Println("创建失败:", err)
		os.Exit(1)
	}
	for _, l := range out {
		f.WriteString(l)
		f.WriteString("\n")
	}
	f.Close()
	fmt.Printf("已生成 MD5.cupa, 共 %d 行\n", len(out))
}

func rg(r int) string { return fmt.Sprintf("#%d", r) }