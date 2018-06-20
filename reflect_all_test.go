/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	. "github.com/badu/reflect"
	"go/ast"
	"go/token"
	"io"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

const (
	testPackageName = "reflect_test"
)

type (
	pair struct {
		i interface{}
		s string
	}

	integer int
	T       struct {
		a int
		b float64
		c string
		d *int
	}

	Basic struct {
		x int
		y float32
	}

	NotBasic Basic

	DeepEqualTest struct {
		a, b interface{}
		eq   bool
	}
	big struct {
		a, b, c, d, e int64
	}

	self struct{}

	Loop  *Loop
	Loopy interface{}

	Recursive struct {
		x int
		r *Recursive
	}

	_Complex struct {
		a int
		b [3]*_Complex
		c *string
		d map[float64]float64
	}

	UnexpT struct {
		m map[int]int
	}

	two [2]uintptr

	emptyStruct struct{}

	nonEmptyStruct struct {
		member int
	}

	// Embedding via pointer.

	Tm1 struct {
		Tm2
	}

	Tm2 struct {
		*Tm3
	}

	Tm3 struct {
		*Tm4
	}

	Tm4 struct {
	}

	Tinter interface {
		M(int, byte) (byte, int)
	}

	Tsmallv byte

	Tsmallp byte

	Twordv uintptr

	Twordp uintptr

	Tbigv [2]uintptr

	Tbigp [2]uintptr

	Point struct {
		x, y int
	}

	T1 struct {
		a string
		int
	}

	FTest struct {
		s     interface{}
		name  string
		index []int
		value int
	}

	D1 struct {
		d int
	}
	D2 struct {
		d int
	}

	S0 struct {
		A, B, C int
		D1
		D2
	}

	S1 struct {
		B int
		S0
	}

	S2 struct {
		A int
		*S1
	}

	S1x struct {
		S1
	}

	S1y struct {
		S1
	}

	S3 struct {
		S1x
		S2
		D, E int
		*S1y
	}

	S4 struct {
		*S4
		A int
	}

	// The X in S6 and S7 annihilate, but they also block the X in S8.S9.
	S5 struct {
		S6
		S7
		S8
	}

	S6 struct {
		X int
	}

	S7 S6

	S8 struct {
		S9
	}

	S9 struct {
		X int
		Y int
	}

	// The X in S11.S6 and S12.S6 annihilate, but they also block the X in S13.S8.S9.
	S10 struct {
		S11
		S12
		S13
	}

	S11 struct {
		S6
	}

	S12 struct {
		S6
	}

	S13 struct {
		S8
	}

	// The X in S15.S11.S1 and S16.S11.S1 annihilate.
	S14 struct {
		S15
		S16
	}

	S15 struct {
		S11
	}

	S16 struct {
		S11
	}

	inner struct {
		x int
	}

	outer struct {
		y int
		inner
	}

	unexpI interface {
		f() (int32, int8)
	}
	unexp struct{}

	InnerInt struct {
		X int
	}
	FuncDDD  func(...interface{}) error
	OuterInt struct {
		Y int
		InnerInt
	}

	Private struct {
		x int
		y **int
		Z int
	}

	private struct {
		Z int
		z int
		S string
		A [1]Private
		T []Private
	}

	Public struct {
		X int
		Y **int
		private
	}

	timp int

	Empty    struct{}
	MyStruct struct {
		x int `some:"tag"`
	}
	MyString string
	MyBytes  []byte
	MyRunes  []int32
	MyFunc   func()
	MyByte   byte

	ComparableStruct struct {
		X int
	}

	NonComparableStruct struct {
		X int
		Y map[string]int
	}
	R0 struct {
		*R1
		*R2
		*R3
		*R4
	}

	R1 struct {
		*R5
		*R6
		*R7
		*R8
	}

	R2 R1
	R3 R1
	R4 R1

	R5 struct {
		*R9
		*R10
		*R11
		*R12
	}

	R6 R5
	R7 R5
	R8 R5

	R9 struct {
		*R13
		*R14
		*R15
		*R16
	}

	R10 R9
	R11 R9
	R12 R9

	R13 struct {
		*R17
		*R18
		*R19
		*R20
	}

	R14 R13
	R15 R13
	R16 R13

	R17 struct {
		*R21
		*R22
		*R23
		*R24
	}

	R18 R17
	R19 R17
	R20 R17

	R21 struct {
		X int
	}

	R22 R21
	R23 R21
	R24 R21

	S struct {
		i1 int64
		i2 int64
	}

	Outer struct {
		*Inner
		R io.Reader
	}

	Inner struct {
		X  *Outer
		P1 uintptr
		P2 uintptr
	}

	Impl struct{}

	UnExportedFirst int

	// Issue 18635 (method version).
	KeepMethodLive struct{}

	funcLayoutTest struct {
		rcvr, t                  *RType
		size, argsize, retOffset uintptr
		stack                    []byte // pointer bitmap: 1 is pointer, 0 is scalar (or uninitialized)
		gc                       []byte
	}

	Tint int

	Tint2 = Tint

	Talias1 struct {
		byte
		uint8
		int
		int32
		rune
	}

	Talias2 struct {
		Tint
		Tint2
	}

	XM struct{ _ bool }

	TheNameOfThisTypeIsExactly255BytesLongSoWhenTheCompilerPrependsTheReflectTestPackageNameAndExtraStarTheLinkerRuntimeAndReflectPackagesWillHaveToCorrectlyDecodeTheSecondLengthByte0123456789_0123456789_0123456789_0123456789_0123456789_012345678 int

	nameTest struct {
		v    interface{}
		want string
	}

	embed struct {
		EmbedWithUnexpMeth
	}
	NonExportedFirst int
	MyBuffer         bytes.Buffer
	notAnExpr        struct{}
	notASTExpr       interface {
		Pos() token.Pos
		End() token.Pos
		exprNode()
	}
	IntPtr  *int
	IntPtr1 *int
)

func (t4 Tm4) M(x int, b byte) (byte, int) { return b, x + 40 }
func (v Tsmallv) M(x int, b byte) (byte, int) {
	fmt.Printf("V Tsmallv.M called with : %d %c\n", x, b)
	return b, x + int(v)
}
func (p *Tsmallp) M(x int, b byte) (byte, int) {
	fmt.Printf("P Tsmallp.M called with : %d %c\n", x, b)
	return b, x + int(*p)
}
func (v Twordv) M(x int, b byte) (byte, int)  { return b, x + int(v) }
func (p *Twordp) M(x int, b byte) (byte, int) { return b, x + int(*p) }
func (v Tbigv) M(x int, b byte) (byte, int)   { return b, x + int(v[0]) + int(v[1]) }
func (p *Tbigp) M(x int, b byte) (byte, int)  { return b, x + int(p[0]) + int(p[1]) }
func (p Point) AnotherMethod(scale int) int {
	println("AnotherMethod called with scale = " + I2A(scale, -1))
	return -1
}
func (p Point) Dist(scale int) int {
	println("Dist called with scale = " + I2A(scale, -1))
	return p.x*p.x*scale + p.y*p.y*scale
}
func (p Point) GCMethod(k int) int {
	println("GCMethod called with k = " + I2A(k, -1))
	runtime.GC()
	return k + p.x
}
func (p Point) NoArgs() { println("NoArgs called.") }
func (p Point) TotalDist(points ...Point) int {
	println("TotalDist called.")
	tot := 0
	for _, q := range points {
		dx := q.x - p.x
		dy := q.y - p.y
		tot += dx*dx + dy*dy // Should call Sqrt, but it's just a test.

	}
	return tot
}

func (*inner) M()               {}
func (*outer) M()               {}
func (*unexp) f() (int32, int8) { return 7, 7 }
func (*unexp) g() (int64, int8) { return 8, 8 }
func (i *InnerInt) M() int {
	println("InnerInt.M Called")
	return i.X
}
func (i InnerInt) MN() int {
	println("InnerInt.MN Called")
	return i.X
}
func (f FuncDDD) M()                  {}
func (p *Private) m()                 {}
func (p private) P()                  {}
func (p Public) M()                   {}
func (t timp) W()                     {}
func (t timp) Y()                     {}
func (t timp) w()                     {}
func (t timp) y()                     {}
func (i UnExportedFirst) ΦExported()  {}
func (i UnExportedFirst) unexported() {}
func (Impl) F()                       {}
func (pi *Inner) M() {
	// Clear references to pi so that the only way the
	// garbage collection will find the pointer is in the
	// argument frame, typed as a *Outer.
	pi.X.Inner = nil

	// Set up an interface value that will cause a crash.
	// P1 = 1 is a non-zero, so the interface looks non-nil.
	// P2 = pi ensures that the data word points into the
	// allocated heap; if not the collection skips the interface
	// value as irrelevant, without dereferencing P1.
	pi.P1 = 1
	pi.P2 = uintptr(unsafe.Pointer(pi))
}
func (k KeepMethodLive) Method1(i int) {
	clobber()
	if i > 0 {
		ToStruct(ReflectOn(k)).MethodByName("Method2").Interface().(func(i int))(i - 1)
	}
}
func (k KeepMethodLive) Method2(i int) {
	clobber()
	ToStruct(ReflectOn(k)).MethodByName("Method1").Interface().(func(i int))(i)
}
func (*XM) String() string {
	return ""
}
func (notAnExpr) Pos() token.Pos {
	return token.NoPos
}
func (notAnExpr) End() token.Pos {
	return token.NoPos
}
func (notAnExpr) exprNode()                      {}
func (i NonExportedFirst) ΦExported(name string) {}
func (i NonExportedFirst) nonexported() int {
	panic("wrong")
}

var (
	sink interface{}
	// Simple functions for DeepEqual tests.
	fn1 func()             // nil.
	fn2 func()             // nil.
	fn3 = func() { fn1() } // Not nil.

	loop1, loop2   Loop
	loopy1, loopy2 Loopy

	deepEqualTests = []DeepEqualTest{
		// Equalities
		{nil, nil, true},
		{1, 1, true},
		{int32(1), int32(1), true},
		{0.5, 0.5, true},
		{float32(0.5), float32(0.5), true},
		{"hello", "hello", true},
		{make([]int, 10), make([]int, 10), true},
		{&[3]int{1, 2, 3}, &[3]int{1, 2, 3}, true},
		{Basic{1, 0.5}, Basic{1, 0.5}, true},
		{error(nil), error(nil), true},
		{map[int]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}, true},
		{fn1, fn2, true},

		// Inequalities
		{1, 2, false},
		{int32(1), int32(2), false},
		{0.5, 0.6, false},
		{float32(0.5), float32(0.6), false},
		{"hello", "hey", false},
		{make([]int, 10), make([]int, 11), false},
		{&[3]int{1, 2, 3}, &[3]int{1, 2, 4}, false},
		{Basic{1, 0.5}, Basic{1, 0.6}, false},
		{Basic{1, 0}, Basic{2, 0}, false},
		{map[int]string{1: "one", 3: "two"}, map[int]string{2: "two", 1: "one"}, false},
		{map[int]string{1: "one", 2: "txo"}, map[int]string{2: "two", 1: "one"}, false},
		{map[int]string{1: "one"}, map[int]string{2: "two", 1: "one"}, false},
		{map[int]string{2: "two", 1: "one"}, map[int]string{1: "one"}, false},
		{nil, 1, false},
		{1, nil, false},
		{fn1, fn3, false},
		{fn3, fn3, false},
		{[][]int{{1}}, [][]int{{2}}, false},
		{math.NaN(), math.NaN(), false},
		{&[1]float64{math.NaN()}, &[1]float64{math.NaN()}, false},
		{&[1]float64{math.NaN()}, self{}, true},
		{[]float64{math.NaN()}, []float64{math.NaN()}, false},
		{[]float64{math.NaN()}, self{}, true},
		{map[float64]float64{math.NaN(): 1}, map[float64]float64{1: 2}, false},
		{map[float64]float64{math.NaN(): 1}, self{}, true},

		// Nil vs empty: not the same.
		{[]int{}, []int(nil), false},
		{[]int{}, []int{}, true},
		{[]int(nil), []int(nil), true},
		{map[int]int{}, map[int]int(nil), false},
		{map[int]int{}, map[int]int{}, true},
		{map[int]int(nil), map[int]int(nil), true},

		// Mismatched types
		{1, 1.0, false},
		{int32(1), int64(1), false},
		{0.5, "hello", false},
		{[]int{1, 2, 3}, [3]int{1, 2, 3}, false},
		{&[3]interface{}{1, 2, 4}, &[3]interface{}{1, 2, "s"}, false},
		{Basic{1, 0.5}, NotBasic{1, 0.5}, false},
		{map[uint]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}, false},

		// Possible loops.
		{&loop1, &loop1, true},
		{&loop1, &loop2, true},
		{&loopy1, &loopy1, true},
		{&loopy1, &loopy2, true},
	}

	typeTests = []pair{
		{struct{ x int }{}, "int"},
		{struct{ x int8 }{}, "int8"},
		{struct{ x int16 }{}, "int16"},
		{struct{ x int32 }{}, "int32"},
		{struct{ x int64 }{}, "int64"},
		{struct{ x uint }{}, "uint"},
		{struct{ x uint8 }{}, "uint8"},
		{struct{ x uint16 }{}, "uint16"},
		{struct{ x uint32 }{}, "uint32"},
		{struct{ x uint64 }{}, "uint64"},
		{struct{ x float32 }{}, "float32"},
		{struct{ x float64 }{}, "float64"},
		{struct{ x int8 }{}, "int8"},
		{struct{ x **int8 }{}, "**int8"},
		{struct{ x **integer }{}, "**" + testPackageName + ".integer"},
		{struct{ x [32]int32 }{}, "[32]int32"},
		{struct{ x []int8 }{}, "[]int8"},
		{struct{ x map[string]int32 }{}, "map[string]int32"},
		{struct{ x chan<- string }{}, "chan<- string"},
		{struct {
			x struct {
				c chan *int32
				d float32
			}
		}{},
			"struct { c chan *int32; d float32 }",
		},
		{struct{ x func(a int8, b int32) }{}, "func(int8, int32)"},
		{struct {
			x struct {
				c func(chan *integer, *int8)
			}
		}{},
			"struct { c func(chan *" + testPackageName + ".integer, *int8) }",
		},
		{struct {
			x struct {
				a int8
				b int32
			}
		}{},
			"struct { a int8; b int32 }",
		},
		{struct {
			x struct {
				a int8
				b int8
				c int32
			}
		}{},
			"struct { a int8; b int8; c int32 }",
		},
		{struct {
			x struct {
				a int8
				b int8
				c int8
				d int32
			}
		}{},
			"struct { a int8; b int8; c int8; d int32 }",
		},
		{struct {
			x struct {
				a int8
				b int8
				c int8
				d int8
				e int32
			}
		}{},
			"struct { a int8; b int8; c int8; d int8; e int32 }",
		},
		{struct {
			x struct {
				a int8
				b int8
				c int8
				d int8
				e int8
				f int32
			}
		}{},
			"struct { a int8; b int8; c int8; d int8; e int8; f int32 }",
		},
		{struct {
			x struct {
				a int8 `reflect:"hi there"`
			}
		}{},
			`struct { a int8 "reflect:\"hi there\"" }`,
		},
		{struct {
			x struct {
				a int8 `reflect:"hi \x00there\t\n\"\\"`
			}
		}{},
			`struct { a int8 "reflect:\"hi \\x00there\\t\\n\\\"\\\\\"" }`,
		},
		{struct {
			x struct {
				f func(args ...int)
			}
		}{},
			"struct { f func(...int) }",
		},
		{struct {
			x interface {
				a(func(func(int) int) func(func(int)) int)
				b()
			}
		}{},
			"interface { " + testPackageName + ".a(func(func(int) int) func(func(int)) int); " + testPackageName + ".b() }",
		},
	}
	valueTests = []pair{
		{new(int), "132"},
		{new(int8), "8"},
		{new(int16), "16"},
		{new(int32), "32"},
		{new(int64), "64"},
		{new(uint), "132"},
		{new(uint8), "8"},
		{new(uint16), "16"},
		{new(uint32), "32"},
		{new(uint64), "64"},
		{new(float32), "256.25"},
		{new(float64), "512.125"},
		{new(complex64), "532.125+10i"},
		{new(complex128), "564.25+1i"},
		{new(string), "stringy cheese"},
		{new(bool), "true"},
		{new(*int8), "*int8(0)"},
		{new(**int8), "**int8(0)"},
		{new([5]int32), "[5]int32{0, 0, 0, 0, 0}"},
		{new(**integer), "**" + testPackageName + ".integer(0)"},
		{new(map[string]int32), "map[string]int32{<can't iterate on maps>}"},
		{new(chan<- string), "chan<- string"},
		{new(func(a int8, b int32)), "func(int8, int32)(0)"},
		{new(struct {
			c chan *int32
			d float32
		}),
			"struct { c chan *int32; d float32 }{chan *int32, 0}",
		},
		{new(struct{ c func(chan *integer, *int8) }),
			"struct { c func(chan *" + testPackageName + ".integer, *int8) }{func(chan *" + testPackageName + ".integer, *int8)(0)}",
		},
		{new(struct {
			a int8
			b int32
		}),
			"struct { a int8; b int32 }{0, 0}",
		},
		{new(struct {
			a int8
			b int8
			c int32
		}),
			"struct { a int8; b int8; c int32 }{0, 0, 0}",
		},
	}

	_i = 7

	valueToStringTests = []pair{
		{123, "123"},
		{123.5, "123.5"},
		{byte(123), "123"},
		{"abc", "abc"},
		{T{123, 456.75, "hello", &_i}, "" + testPackageName + ".T{123, 456.75, hello, *int(&7)}"},
		{new(chan *T), "*chan *" + testPackageName + ".T(&chan *" + testPackageName + ".T)"},
		{[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}"},
		{&[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "*[10]int(&[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})"},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}"},
		{&[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "*[]int(&[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})"},
	}
	appendTests = []struct {
		orig, extra []int
	}{
		{make([]int, 2, 4), []int{22}},
		{make([]int, 2, 4), []int{22, 33, 44}},
	}

	unexpi unexpI = new(unexp)

	tagGetTests = []struct {
		Tag   string
		Key   string
		Value string
	}{
		{`protobuf:"PB(1,2)"`, `protobuf`, `PB(1,2)`},
		{`protobuf:"PB(1,2)"`, `foo`, ``},
		{`protobuf:"PB(1,2)"`, `rotobuf`, ``},
		{`protobuf:"PB(1,2)" json:"name"`, `json`, `name`},
		{`protobuf:"PB(1,2)" json:"name"`, `protobuf`, `PB(1,2)`},
		{`k0:"values contain spaces" k1:"and\ttabs"`, "k0", "values contain spaces"},
		{`k0:"values contain spaces" k1:"and\ttabs"`, "k1", "and\ttabs"},
	}

	convertTests = []struct {
		in  Value
		out Value
	}{
		// numbers
		{ReflectOn(int8(1)), ReflectOn(int8(1))},
		{ReflectOn(int8(2)), ReflectOn(uint8(2))},
		{ReflectOn(uint8(3)), ReflectOn(int8(3))},
		{ReflectOn(int8(4)), ReflectOn(int16(4))},
		{ReflectOn(int16(5)), ReflectOn(int8(5))},
		{ReflectOn(int8(6)), ReflectOn(uint16(6))},
		{ReflectOn(uint16(7)), ReflectOn(int8(7))},
		{ReflectOn(int8(8)), ReflectOn(int32(8))},
		{ReflectOn(int32(9)), ReflectOn(int8(9))},
		{ReflectOn(int8(10)), ReflectOn(uint32(10))},
		{ReflectOn(uint32(11)), ReflectOn(int8(11))},
		{ReflectOn(int8(12)), ReflectOn(int64(12))},
		{ReflectOn(int64(13)), ReflectOn(int8(13))},
		{ReflectOn(int8(14)), ReflectOn(uint64(14))},
		{ReflectOn(uint64(15)), ReflectOn(int8(15))},
		{ReflectOn(int8(16)), ReflectOn(int(16))},
		{ReflectOn(int(17)), ReflectOn(int8(17))},
		{ReflectOn(int8(18)), ReflectOn(uint(18))},
		{ReflectOn(uint(19)), ReflectOn(int8(19))},
		{ReflectOn(int8(20)), ReflectOn(uintptr(20))},
		{ReflectOn(uintptr(21)), ReflectOn(int8(21))},
		{ReflectOn(int8(22)), ReflectOn(float32(22))},
		{ReflectOn(float32(23)), ReflectOn(int8(23))},
		{ReflectOn(int8(24)), ReflectOn(float64(24))},
		{ReflectOn(float64(25)), ReflectOn(int8(25))},
		{ReflectOn(uint8(26)), ReflectOn(uint8(26))},
		{ReflectOn(uint8(27)), ReflectOn(int16(27))},
		{ReflectOn(int16(28)), ReflectOn(uint8(28))},
		{ReflectOn(uint8(29)), ReflectOn(uint16(29))},
		{ReflectOn(uint16(30)), ReflectOn(uint8(30))},
		{ReflectOn(uint8(31)), ReflectOn(int32(31))},
		{ReflectOn(int32(32)), ReflectOn(uint8(32))},
		{ReflectOn(uint8(33)), ReflectOn(uint32(33))},
		{ReflectOn(uint32(34)), ReflectOn(uint8(34))},
		{ReflectOn(uint8(35)), ReflectOn(int64(35))},
		{ReflectOn(int64(36)), ReflectOn(uint8(36))},
		{ReflectOn(uint8(37)), ReflectOn(uint64(37))},
		{ReflectOn(uint64(38)), ReflectOn(uint8(38))},
		{ReflectOn(uint8(39)), ReflectOn(int(39))},
		{ReflectOn(int(40)), ReflectOn(uint8(40))},
		{ReflectOn(uint8(41)), ReflectOn(uint(41))},
		{ReflectOn(uint(42)), ReflectOn(uint8(42))},
		{ReflectOn(uint8(43)), ReflectOn(uintptr(43))},
		{ReflectOn(uintptr(44)), ReflectOn(uint8(44))},
		{ReflectOn(uint8(45)), ReflectOn(float32(45))},
		{ReflectOn(float32(46)), ReflectOn(uint8(46))},
		{ReflectOn(uint8(47)), ReflectOn(float64(47))},
		{ReflectOn(float64(48)), ReflectOn(uint8(48))},
		{ReflectOn(int16(49)), ReflectOn(int16(49))},
		{ReflectOn(int16(50)), ReflectOn(uint16(50))},
		{ReflectOn(uint16(51)), ReflectOn(int16(51))},
		{ReflectOn(int16(52)), ReflectOn(int32(52))},
		{ReflectOn(int32(53)), ReflectOn(int16(53))},
		{ReflectOn(int16(54)), ReflectOn(uint32(54))},
		{ReflectOn(uint32(55)), ReflectOn(int16(55))},
		{ReflectOn(int16(56)), ReflectOn(int64(56))},
		{ReflectOn(int64(57)), ReflectOn(int16(57))},
		{ReflectOn(int16(58)), ReflectOn(uint64(58))},
		{ReflectOn(uint64(59)), ReflectOn(int16(59))},
		{ReflectOn(int16(60)), ReflectOn(int(60))},
		{ReflectOn(int(61)), ReflectOn(int16(61))},
		{ReflectOn(int16(62)), ReflectOn(uint(62))},
		{ReflectOn(uint(63)), ReflectOn(int16(63))},
		{ReflectOn(int16(64)), ReflectOn(uintptr(64))},
		{ReflectOn(uintptr(65)), ReflectOn(int16(65))},
		{ReflectOn(int16(66)), ReflectOn(float32(66))},
		{ReflectOn(float32(67)), ReflectOn(int16(67))},
		{ReflectOn(int16(68)), ReflectOn(float64(68))},
		{ReflectOn(float64(69)), ReflectOn(int16(69))},
		{ReflectOn(uint16(70)), ReflectOn(uint16(70))},
		{ReflectOn(uint16(71)), ReflectOn(int32(71))},
		{ReflectOn(int32(72)), ReflectOn(uint16(72))},
		{ReflectOn(uint16(73)), ReflectOn(uint32(73))},
		{ReflectOn(uint32(74)), ReflectOn(uint16(74))},
		{ReflectOn(uint16(75)), ReflectOn(int64(75))},
		{ReflectOn(int64(76)), ReflectOn(uint16(76))},
		{ReflectOn(uint16(77)), ReflectOn(uint64(77))},
		{ReflectOn(uint64(78)), ReflectOn(uint16(78))},
		{ReflectOn(uint16(79)), ReflectOn(int(79))},
		{ReflectOn(int(80)), ReflectOn(uint16(80))},
		{ReflectOn(uint16(81)), ReflectOn(uint(81))},
		{ReflectOn(uint(82)), ReflectOn(uint16(82))},
		{ReflectOn(uint16(83)), ReflectOn(uintptr(83))},
		{ReflectOn(uintptr(84)), ReflectOn(uint16(84))},
		{ReflectOn(uint16(85)), ReflectOn(float32(85))},
		{ReflectOn(float32(86)), ReflectOn(uint16(86))},
		{ReflectOn(uint16(87)), ReflectOn(float64(87))},
		{ReflectOn(float64(88)), ReflectOn(uint16(88))},
		{ReflectOn(int32(89)), ReflectOn(int32(89))},
		{ReflectOn(int32(90)), ReflectOn(uint32(90))},
		{ReflectOn(uint32(91)), ReflectOn(int32(91))},
		{ReflectOn(int32(92)), ReflectOn(int64(92))},
		{ReflectOn(int64(93)), ReflectOn(int32(93))},
		{ReflectOn(int32(94)), ReflectOn(uint64(94))},
		{ReflectOn(uint64(95)), ReflectOn(int32(95))},
		{ReflectOn(int32(96)), ReflectOn(int(96))},
		{ReflectOn(int(97)), ReflectOn(int32(97))},
		{ReflectOn(int32(98)), ReflectOn(uint(98))},
		{ReflectOn(uint(99)), ReflectOn(int32(99))},
		{ReflectOn(int32(100)), ReflectOn(uintptr(100))},
		{ReflectOn(uintptr(101)), ReflectOn(int32(101))},
		{ReflectOn(int32(102)), ReflectOn(float32(102))},
		{ReflectOn(float32(103)), ReflectOn(int32(103))},
		{ReflectOn(int32(104)), ReflectOn(float64(104))},
		{ReflectOn(float64(105)), ReflectOn(int32(105))},
		{ReflectOn(uint32(106)), ReflectOn(uint32(106))},
		{ReflectOn(uint32(107)), ReflectOn(int64(107))},
		{ReflectOn(int64(108)), ReflectOn(uint32(108))},
		{ReflectOn(uint32(109)), ReflectOn(uint64(109))},
		{ReflectOn(uint64(110)), ReflectOn(uint32(110))},
		{ReflectOn(uint32(111)), ReflectOn(int(111))},
		{ReflectOn(int(112)), ReflectOn(uint32(112))},
		{ReflectOn(uint32(113)), ReflectOn(uint(113))},
		{ReflectOn(uint(114)), ReflectOn(uint32(114))},
		{ReflectOn(uint32(115)), ReflectOn(uintptr(115))},
		{ReflectOn(uintptr(116)), ReflectOn(uint32(116))},
		{ReflectOn(uint32(117)), ReflectOn(float32(117))},
		{ReflectOn(float32(118)), ReflectOn(uint32(118))},
		{ReflectOn(uint32(119)), ReflectOn(float64(119))},
		{ReflectOn(float64(120)), ReflectOn(uint32(120))},
		{ReflectOn(int64(121)), ReflectOn(int64(121))},
		{ReflectOn(int64(122)), ReflectOn(uint64(122))},
		{ReflectOn(uint64(123)), ReflectOn(int64(123))},
		{ReflectOn(int64(124)), ReflectOn(int(124))},
		{ReflectOn(int(125)), ReflectOn(int64(125))},
		{ReflectOn(int64(126)), ReflectOn(uint(126))},
		{ReflectOn(uint(127)), ReflectOn(int64(127))},
		{ReflectOn(int64(128)), ReflectOn(uintptr(128))},
		{ReflectOn(uintptr(129)), ReflectOn(int64(129))},
		{ReflectOn(int64(130)), ReflectOn(float32(130))},
		{ReflectOn(float32(131)), ReflectOn(int64(131))},
		{ReflectOn(int64(132)), ReflectOn(float64(132))},
		{ReflectOn(float64(133)), ReflectOn(int64(133))},
		{ReflectOn(uint64(134)), ReflectOn(uint64(134))},
		{ReflectOn(uint64(135)), ReflectOn(int(135))},
		{ReflectOn(int(136)), ReflectOn(uint64(136))},
		{ReflectOn(uint64(137)), ReflectOn(uint(137))},
		{ReflectOn(uint(138)), ReflectOn(uint64(138))},
		{ReflectOn(uint64(139)), ReflectOn(uintptr(139))},
		{ReflectOn(uintptr(140)), ReflectOn(uint64(140))},
		{ReflectOn(uint64(141)), ReflectOn(float32(141))},
		{ReflectOn(float32(142)), ReflectOn(uint64(142))},
		{ReflectOn(uint64(143)), ReflectOn(float64(143))},
		{ReflectOn(float64(144)), ReflectOn(uint64(144))},
		{ReflectOn(int(145)), ReflectOn(int(145))},
		{ReflectOn(int(146)), ReflectOn(uint(146))},
		{ReflectOn(uint(147)), ReflectOn(int(147))},
		{ReflectOn(int(148)), ReflectOn(uintptr(148))},
		{ReflectOn(uintptr(149)), ReflectOn(int(149))},
		{ReflectOn(int(150)), ReflectOn(float32(150))},
		{ReflectOn(float32(151)), ReflectOn(int(151))},
		{ReflectOn(int(152)), ReflectOn(float64(152))},
		{ReflectOn(float64(153)), ReflectOn(int(153))},
		{ReflectOn(uint(154)), ReflectOn(uint(154))},
		{ReflectOn(uint(155)), ReflectOn(uintptr(155))},
		{ReflectOn(uintptr(156)), ReflectOn(uint(156))},
		{ReflectOn(uint(157)), ReflectOn(float32(157))},
		{ReflectOn(float32(158)), ReflectOn(uint(158))},
		{ReflectOn(uint(159)), ReflectOn(float64(159))},
		{ReflectOn(float64(160)), ReflectOn(uint(160))},
		{ReflectOn(uintptr(161)), ReflectOn(uintptr(161))},
		{ReflectOn(uintptr(162)), ReflectOn(float32(162))},
		{ReflectOn(float32(163)), ReflectOn(uintptr(163))},
		{ReflectOn(uintptr(164)), ReflectOn(float64(164))},
		{ReflectOn(float64(165)), ReflectOn(uintptr(165))},
		{ReflectOn(float32(166)), ReflectOn(float32(166))},
		{ReflectOn(float32(167)), ReflectOn(float64(167))},
		{ReflectOn(float64(168)), ReflectOn(float32(168))},
		{ReflectOn(float64(169)), ReflectOn(float64(169))},

		// truncation
		{ReflectOn(float64(1.5)), ReflectOn(int(1))},

		// complex
		{ReflectOn(complex64(1i)), ReflectOn(complex64(1i))},
		{ReflectOn(complex64(2i)), ReflectOn(complex128(2i))},
		{ReflectOn(complex128(3i)), ReflectOn(complex64(3i))},
		{ReflectOn(complex128(4i)), ReflectOn(complex128(4i))},

		// string
		{ReflectOn(string("hello")), ReflectOn(string("hello"))},
		{ReflectOn(string("bytes1")), ReflectOn([]byte("bytes1"))},
		{ReflectOn([]byte("bytes2")), ReflectOn(string("bytes2"))},
		{ReflectOn([]byte("bytes3")), ReflectOn([]byte("bytes3"))},
		{ReflectOn(string("runes♝")), ReflectOn([]rune("runes♝"))},
		{ReflectOn([]rune("runes♕")), ReflectOn(string("runes♕"))},
		{ReflectOn([]rune("runes🙈🙉🙊")), ReflectOn([]rune("runes🙈🙉🙊"))},
		{ReflectOn(int('a')), ReflectOn(string("a"))},
		{ReflectOn(int8('a')), ReflectOn(string("a"))},
		{ReflectOn(int16('a')), ReflectOn(string("a"))},
		{ReflectOn(int32('a')), ReflectOn(string("a"))},
		{ReflectOn(int64('a')), ReflectOn(string("a"))},
		{ReflectOn(uint('a')), ReflectOn(string("a"))},
		{ReflectOn(uint8('a')), ReflectOn(string("a"))},
		{ReflectOn(uint16('a')), ReflectOn(string("a"))},
		{ReflectOn(uint32('a')), ReflectOn(string("a"))},
		{ReflectOn(uint64('a')), ReflectOn(string("a"))},
		{ReflectOn(uintptr('a')), ReflectOn(string("a"))},
		{ReflectOn(int(-1)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(int8(-2)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(int16(-3)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(int32(-4)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(int64(-5)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(uint(0x110001)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(uint32(0x110002)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(uint64(0x110003)), ReflectOn(string("\uFFFD"))},
		{ReflectOn(uintptr(0x110004)), ReflectOn(string("\uFFFD"))},

		// named string
		{ReflectOn(MyString("hello")), ReflectOn(string("hello"))},
		{ReflectOn(string("hello")), ReflectOn(MyString("hello"))},
		{ReflectOn(string("hello")), ReflectOn(string("hello"))},
		{ReflectOn(MyString("hello")), ReflectOn(MyString("hello"))},
		{ReflectOn(MyString("bytes1")), ReflectOn([]byte("bytes1"))},
		{ReflectOn([]byte("bytes2")), ReflectOn(MyString("bytes2"))},
		{ReflectOn([]byte("bytes3")), ReflectOn([]byte("bytes3"))},
		{ReflectOn(MyString("runes♝")), ReflectOn([]rune("runes♝"))},
		{ReflectOn([]rune("runes♕")), ReflectOn(MyString("runes♕"))},
		{ReflectOn([]rune("runes🙈🙉🙊")), ReflectOn([]rune("runes🙈🙉🙊"))},
		{ReflectOn([]rune("runes🙈🙉🙊")), ReflectOn(MyRunes("runes🙈🙉🙊"))},
		{ReflectOn(MyRunes("runes🙈🙉🙊")), ReflectOn([]rune("runes🙈🙉🙊"))},
		{ReflectOn(int('a')), ReflectOn(MyString("a"))},
		{ReflectOn(int8('a')), ReflectOn(MyString("a"))},
		{ReflectOn(int16('a')), ReflectOn(MyString("a"))},
		{ReflectOn(int32('a')), ReflectOn(MyString("a"))},
		{ReflectOn(int64('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uint('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uint8('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uint16('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uint32('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uint64('a')), ReflectOn(MyString("a"))},
		{ReflectOn(uintptr('a')), ReflectOn(MyString("a"))},
		{ReflectOn(int(-1)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(int8(-2)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(int16(-3)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(int32(-4)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(int64(-5)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(uint(0x110001)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(uint32(0x110002)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(uint64(0x110003)), ReflectOn(MyString("\uFFFD"))},
		{ReflectOn(uintptr(0x110004)), ReflectOn(MyString("\uFFFD"))},

		// named []byte
		{ReflectOn(string("bytes1")), ReflectOn(MyBytes("bytes1"))},
		{ReflectOn(MyBytes("bytes2")), ReflectOn(string("bytes2"))},
		{ReflectOn(MyBytes("bytes3")), ReflectOn(MyBytes("bytes3"))},
		{ReflectOn(MyString("bytes1")), ReflectOn(MyBytes("bytes1"))},
		{ReflectOn(MyBytes("bytes2")), ReflectOn(MyString("bytes2"))},

		// named []rune
		{ReflectOn(string("runes♝")), ReflectOn(MyRunes("runes♝"))},
		{ReflectOn(MyRunes("runes♕")), ReflectOn(string("runes♕"))},
		{ReflectOn(MyRunes("runes🙈🙉🙊")), ReflectOn(MyRunes("runes🙈🙉🙊"))},
		{ReflectOn(MyString("runes♝")), ReflectOn(MyRunes("runes♝"))},
		{ReflectOn(MyRunes("runes♕")), ReflectOn(MyString("runes♕"))},

		// named types and equal underlying types
		{ReflectOn(new(int)), ReflectOn(new(integer))},
		{ReflectOn(new(integer)), ReflectOn(new(int))},
		{ReflectOn(Empty{}), ReflectOn(struct{}{})},
		{ReflectOn(new(Empty)), ReflectOn(new(struct{}))},
		{ReflectOn(struct{}{}), ReflectOn(Empty{})},
		{ReflectOn(new(struct{})), ReflectOn(new(Empty))},
		{ReflectOn(Empty{}), ReflectOn(Empty{})},
		{ReflectOn(MyBytes{}), ReflectOn([]byte{})},
		{ReflectOn([]byte{}), ReflectOn(MyBytes{})},
		{ReflectOn((func())(nil)), ReflectOn(MyFunc(nil))},
		{ReflectOn((MyFunc)(nil)), ReflectOn((func())(nil))},

		// structs with different tags
		{ReflectOn(struct {
			x int `some:"foo"`
		}{}), ReflectOn(struct {
			x int `some:"bar"`
		}{})},

		{ReflectOn(struct {
			x int `some:"bar"`
		}{}), ReflectOn(struct {
			x int `some:"foo"`
		}{})},

		{ReflectOn(MyStruct{}), ReflectOn(struct {
			x int `some:"foo"`
		}{})},

		{ReflectOn(struct {
			x int `some:"foo"`
		}{}), ReflectOn(MyStruct{})},

		{ReflectOn(MyStruct{}), ReflectOn(struct {
			x int `some:"bar"`
		}{})},

		{ReflectOn(struct {
			x int `some:"bar"`
		}{}), ReflectOn(MyStruct{})},

		// can convert *byte and *MyByte
		{ReflectOn((*byte)(nil)), ReflectOn((*MyByte)(nil))},
		{ReflectOn((*MyByte)(nil)), ReflectOn((*byte)(nil))},

		// cannot convert mismatched array sizes
		{ReflectOn([2]byte{}), ReflectOn([2]byte{})},
		{ReflectOn([3]byte{}), ReflectOn([3]byte{})},

		// cannot convert other instances
		{ReflectOn((**byte)(nil)), ReflectOn((**byte)(nil))},
		{ReflectOn((**MyByte)(nil)), ReflectOn((**MyByte)(nil))},
		{ReflectOn(([]byte)(nil)), ReflectOn(([]byte)(nil))},
		{ReflectOn(([]MyByte)(nil)), ReflectOn(([]MyByte)(nil))},
		{ReflectOn((map[int]byte)(nil)), ReflectOn((map[int]byte)(nil))},
		{ReflectOn((map[int]MyByte)(nil)), ReflectOn((map[int]MyByte)(nil))},
		{ReflectOn((map[byte]int)(nil)), ReflectOn((map[byte]int)(nil))},
		{ReflectOn((map[MyByte]int)(nil)), ReflectOn((map[MyByte]int)(nil))},
		{ReflectOn([2]byte{}), ReflectOn([2]byte{})},
		{ReflectOn([2]MyByte{}), ReflectOn([2]MyByte{})},

		// other
		{ReflectOn((***int)(nil)), ReflectOn((***int)(nil))},
		{ReflectOn((***byte)(nil)), ReflectOn((***byte)(nil))},
		{ReflectOn((***int32)(nil)), ReflectOn((***int32)(nil))},
		{ReflectOn((***int64)(nil)), ReflectOn((***int64)(nil))},
		{ReflectOn((map[int]bool)(nil)), ReflectOn((map[int]bool)(nil))},
		{ReflectOn((map[int]byte)(nil)), ReflectOn((map[int]byte)(nil))},
		{ReflectOn((map[uint]bool)(nil)), ReflectOn((map[uint]bool)(nil))},
		{ReflectOn([]uint(nil)), ReflectOn([]uint(nil))},
		{ReflectOn([]int(nil)), ReflectOn([]int(nil))},
		{ReflectOn(new(interface{})), ReflectOn(new(interface{}))},
		{ReflectOn(new(io.Reader)), ReflectOn(new(io.Reader))},
		{ReflectOn(new(io.Writer)), ReflectOn(new(io.Writer))},

		// interfaces
		{ReflectOn(int(1)), EmptyInterfaceV(int(1))},
		{ReflectOn(string("hello")), EmptyInterfaceV(string("hello"))},
		{ReflectOn(new(bytes.Buffer)), ReaderV(new(bytes.Buffer))},
		{ReadWriterV(new(bytes.Buffer)), ReaderV(new(bytes.Buffer))},
		{ReflectOn(new(bytes.Buffer)), ReadWriterV(new(bytes.Buffer))},
	}

	comparableTests = []struct {
		typ *RType
		ok  bool
	}{
		{ReflectOn(1).Type, true},
		{ReflectOn("hello").Type, true},
		{ReflectOn(new(byte)).Type, true},
		{ReflectOn((func())(nil)).Type, false},
		{ReflectOn([]byte{}).Type, false},
		{ReflectOn(map[string]int{}).Type, false},
		{ReflectOn(make(chan int)).Type, true},
		{ReflectOn(1.5).Type, true},
		{ReflectOn(false).Type, true},
		{ReflectOn(1i).Type, true},
		{ReflectOn(ComparableStruct{}).Type, true},
		{ReflectOn(NonComparableStruct{}).Type, false},
		{ReflectOn([10]map[string]int{}).Type, false},
		{ReflectOn([10]string{}).Type, true},
		{ReflectOn(new(interface{})).Type.Deref(), true},
	}
	funcLayoutTests []funcLayoutTest
	nameTests       = []nameTest{
		{(*int32)(nil), "int32"},
		{(*D1)(nil), "D1"},
		{(*[]D1)(nil), ""},
		{(*chan D1)(nil), ""},
		{(*func() D1)(nil), ""},
		{(*interface{})(nil), ""},
		{(*interface {
			F()
		})(nil), ""},
		{(*TheNameOfThisTypeIsExactly255BytesLongSoWhenTheCompilerPrependsTheReflectTestPackageNameAndExtraStarTheLinkerRuntimeAndReflectPackagesWillHaveToCorrectlyDecodeTheSecondLengthByte0123456789_0123456789_0123456789_0123456789_0123456789_012345678)(nil), "TheNameOfThisTypeIsExactly255BytesLongSoWhenTheCompilerPrependsTheReflectTestPackageNameAndExtraStarTheLinkerRuntimeAndReflectPackagesWillHaveToCorrectlyDecodeTheSecondLengthByte0123456789_0123456789_0123456789_0123456789_0123456789_012345678"},
	}

	implementsTests = []struct {
		x interface{}
		t interface{}
		b bool
	}{
		{new(*bytes.Buffer), new(io.Reader), true},
		{new(bytes.Buffer), new(io.Reader), false},
		{new(*bytes.Buffer), new(io.ReaderAt), false},
		{new(*ast.Ident), new(ast.Expr), true},
		{new(*notAnExpr), new(ast.Expr), false},
		{new(*ast.Ident), new(notASTExpr), false},
		{new(notASTExpr), new(ast.Expr), false},
		{new(ast.Expr), new(notASTExpr), false},
		{new(*notAnExpr), new(notASTExpr), true},
	}

	assignableTests = []struct {
		x interface{}
		t interface{}
		b bool
	}{
		{new(*int), new(IntPtr), true},
		{new(IntPtr), new(*int), true},
		{new(IntPtr), new(IntPtr1), false},
		// test runs implementsTests too
	}
)

func naclpad() []byte {
	if runtime.GOARCH == "amd64p32" {
		return lit(0)
	}
	return nil
}

func rep(n int, b []byte) []byte { return bytes.Repeat(b, n) }
func join(b ...[]byte) []byte    { return bytes.Join(b, nil) }
func lit(x ...byte) []byte       { return x }

// clobber tries to clobber unreachable memory.
func clobber() {
	runtime.GC()
	for i := 1; i < 32; i++ {
		for j := 0; j < 10; j++ {
			obj := make([]*byte, i)
			sink = obj
		}
	}
	runtime.GC()
}

func verifyGCBits(t *testing.T, typ *RType, bits []byte) {
	heapBits := GCBits(New(typ).Interface())
	if !bytes.Equal(heapBits, bits) {
		t.Errorf("heapBits incorrect for %v\nhave %v\nwant %v", typ, heapBits, bits)
	}
}

func verifyGCBitsSlice(t *testing.T, typ *RType, cap int, bits []byte) {
	// Creating a slice causes the runtime to repeat a bitmap, which exercises a different path from making the compiler repeat a bitmap for a small array or executing a repeat in a GC program.
	val := MakeSlice(typ, 0, cap)
	data := NewAt(ArrayOf(typ, cap), unsafe.Pointer(val.Pointer()))
	heapBits := GCBits(data.Interface())
	// Repeat the bitmap for the slice size, trimming scalars in
	// the last element.
	bits = rep(cap, bits)
	for len(bits) > 2 && bits[len(bits)-1] == 0 {
		bits = bits[:len(bits)-1]
	}
	if len(bits) == 2 && bits[0] == 0 && bits[1] == 0 {
		bits = bits[:0]
	}
	if !bytes.Equal(heapBits, bits) {
		t.Errorf("heapBits incorrect for make(%v, 0, %v)\nhave %v\nwant %v", typ, cap, heapBits, bits)
	}
}

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
func methodName() string {
	pc, _, line, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name() + " line : " + strconv.Itoa(line)
}

func checkSameType(t *testing.T, x, y interface{}) {
	if ReflectOn(x).Type != ReflectOn(y).Type {
		t.Errorf(methodName()+" did not find preexisting type for %s (vs %s)", ReflectOn(x).Type, ReflectOn(y).Type)
	}
}

func EmptyInterfaceV(x interface{}) Value {
	return ReflectOnPtr(&x)
}

func ReaderV(x io.Reader) Value {
	return ReflectOnPtr(&x)
}

func ReadWriterV(x io.ReadWriter) Value {
	return ReflectOnPtr(&x)
}

func isNonNil(x interface{}) {
	if x == nil {
		panic("nil interface")
	}
}

func isValid(v Value) {
	if !v.IsValid() {
		panic("zero Value")
	}
}

func noAlloc(t *testing.T, n int, f func(int)) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping; GOMAXPROCS>1")
	}
	i := -1
	allocs := testing.AllocsPerRun(n, func() {
		f(i)
		i++
	})
	if allocs > 0 {
		t.Errorf("%d iterations: got %v mallocs, want 0", n, allocs)
	}
}

func returnEmpty() emptyStruct {
	return emptyStruct{}
}

func takesEmpty(e emptyStruct) {
}

func returnNonEmpty(i int) nonEmptyStruct {
	return nonEmptyStruct{member: i}
}

func takesNonEmpty(n nonEmptyStruct) int {
	return n.member
}

func sameInts(x, y []int) bool {
	if len(x) != len(y) {
		return false
	}
	for i, xx := range x {
		if xx != y[i] {
			return false
		}
	}
	return true
}

func testType(t *testing.T, i int, typ *RType, want string) {
	s := typ.String()
	if s != want {
		t.Errorf("#%d: have %#q, want %#q", i, s, want)
	}
}

func assert(t *testing.T, s, want string) {
	if s != want {
		t.Errorf("have %#q want %#q", s, want)
	}
}

// Difficult test for function call because of
// implicit padding between arguments.
func dummy(b byte, c int, d byte, e two, f byte, g float32, h byte) (i byte, j int, k byte, l two, m byte, n float32, o byte) {
	return b, c, d, e, f, g, h
}

func TestBool(t *testing.T) {
	v := ReflectOn(true)
	if v.Bool().Get() != true {
		t.Error("ReflectOn(true).Bool() = false")
	}
}
func TestTypes(t *testing.T) {
	for i, tt := range typeTests {
		testType(t, i, ToStruct(ReflectOn(tt.i)).Field(0).Type, tt.s)
	}
}

func TestSet(t *testing.T) {
	for i, tt := range valueTests {
		v := ReflectOnPtr(tt.i)
		switch v.Kind() {
		case Int:
			v.Int().Set(132)
		case Int8:
			v.Int().Set(8)
		case Int16:
			v.Int().Set(16)
		case Int32:
			v.Int().Set(32)
		case Int64:
			v.Int().Set(64)
		case Uint:
			v.Uint().Set(132)
		case Uint8:
			v.Uint().Set(8)
		case Uint16:
			v.Uint().Set(16)
		case Uint32:
			v.Uint().Set(32)
		case Uint64:
			v.Uint().Set(64)
		case Float32:
			v.Float().Set(256.25)
		case Float64:
			v.Float().Set(512.125)
		case Complex64:
			v.Complex().Set(532.125 + 10i)
		case Complex128:
			v.Complex().Set(564.25 + 1i)
		case String:
			v.String().Set("stringy cheese")
		case Bool:
			v.Bool().Set(true)
		}
		s := ValueToString(v)
		if s != tt.s {
			t.Errorf("#%d: have %#q, want %#q", i, s, tt.s)
		}
	}
}

func TestSetValue(t *testing.T) {
	for i, tt := range valueTests {
		v := ReflectOnPtr(tt.i)
		switch v.Kind() {
		case Int:
			v.Set(ReflectOn(int(132)))
		case Int8:
			v.Set(ReflectOn(int8(8)))
		case Int16:
			v.Set(ReflectOn(int16(16)))
		case Int32:
			v.Set(ReflectOn(int32(32)))
		case Int64:
			v.Set(ReflectOn(int64(64)))
		case Uint:
			v.Set(ReflectOn(uint(132)))
		case Uint8:
			v.Set(ReflectOn(uint8(8)))
		case Uint16:
			v.Set(ReflectOn(uint16(16)))
		case Uint32:
			v.Set(ReflectOn(uint32(32)))
		case Uint64:
			v.Set(ReflectOn(uint64(64)))
		case Float32:
			v.Set(ReflectOn(float32(256.25)))
		case Float64:
			v.Set(ReflectOn(512.125))
		case Complex64:
			v.Set(ReflectOn(complex64(532.125 + 10i)))
		case Complex128:
			v.Set(ReflectOn(complex128(564.25 + 1i)))
		case String:
			v.Set(ReflectOn("stringy cheese"))
		case Bool:
			v.Set(ReflectOn(true))
		}
		s := ValueToString(v)
		if s != tt.s {
			t.Errorf("#%d: have %#q, want %#q", i, s, tt.s)
		}
	}
}

func TestCanSetField(t *testing.T) {
	type embed struct{ x, X int }
	type Embed struct{ x, X int }
	type S1 struct {
		embed
		x, X int
	}
	type S2 struct {
		*embed
		x, X int
	}
	type S3 struct {
		Embed
		x, X int
	}
	type S4 struct {
		*Embed
		x, X int
	}

	type testCase struct {
		index  []int
		canSet bool
	}
	tests := []struct {
		val   Value
		cases []testCase
	}{{
		val: ReflectOn(&S1{}),
		cases: []testCase{
			{[]int{0}, false},
			{[]int{0, 0}, false},
			{[]int{0, 1}, true},
			{[]int{1}, false},
			{[]int{2}, true},
		},
	}, {
		val: ReflectOn(&S2{embed: &embed{}}),
		cases: []testCase{
			{[]int{0}, false},
			{[]int{0, 0}, false},
			{[]int{0, 1}, true},
			{[]int{1}, false},
			{[]int{2}, true},
		},
	}, {
		val: ReflectOn(&S3{}),
		cases: []testCase{
			{[]int{0}, true},
			{[]int{0, 0}, false},
			{[]int{0, 1}, true},
			{[]int{1}, false},
			{[]int{2}, true},
		},
	}, {
		val: ReflectOn(&S4{Embed: &Embed{}}),
		cases: []testCase{
			{[]int{0}, true},
			{[]int{0, 0}, false},
			{[]int{0, 1}, true},
			{[]int{1}, false},
			{[]int{2}, true},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.val.Type.Name(), func(t *testing.T) {
			for _, testCase := range tt.cases {
				f := tt.val
				for _, i := range testCase.index {
					if f.Kind() == Ptr {
						f = f.Deref()
					}
					f = ToStruct(f).Field(i)
				}
				if got := f.CanSet(); got != testCase.canSet {
					t.Errorf("CanSet() = %v, want %v", got, testCase.canSet)
				}
			}
		})
	}
}

func TestValueToString(t *testing.T) {
	for i, test := range valueToStringTests {
		s := ValueToString(ReflectOn(test.i))
		if s != test.s {
			t.Errorf("#%d: have %#q, want %#q", i, s, test.s)
		}
	}
}

func TestArrayElemSet(t *testing.T) {
	v := ToSlice(ReflectOnPtr(&[10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
	v.Index(4).Int().Set(123)
	s := ValueToString(v.Value)
	const want = "[10]int{1, 2, 3, 4, 123, 6, 7, 8, 9, 10}"
	if s != want {
		t.Errorf("[10]int: have %#q want %#q", s, want)
	}

	v = ToSlice(ReflectOn([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
	v.Index(4).Int().Set(123)
	s = ValueToString(v.Value)
	const want1 = "[]int{1, 2, 3, 4, 123, 6, 7, 8, 9, 10}"
	if s != want1 {
		t.Errorf("[]int: have %#q want %#q", s, want1)
	}
}

func TestAll(t *testing.T) {
	testType(t, 1, ReflectOn((int8)(0)).Type, "int8")
	testType(t, 2, ReflectOn((*int8)(nil)).Type.Deref(), "int8")

	typ := ReflectOn((*struct {
		c chan *int32
		d float32
	})(nil)).Type
	testType(t, 3, typ, "*struct { c chan *int32; d float32 }")
	etyp := typ.Deref()
	testType(t, 4, etyp, "struct { c chan *int32; d float32 }")

	typ = ReflectOn([32]int32{}).Type
	testType(t, 7, typ, "[32]int32")
	testType(t, 8, typ.ConvToArray().ElemType, "int32")

	typ = ReflectOn((map[string]*int32)(nil)).Type
	testType(t, 9, typ, "map[string]*int32")
	mtyp := typ
	testType(t, 10, mtyp.ConvToMap().KeyType, "string")
	testType(t, 11, mtyp.ConvToMap().ElemType, "*int32")

	// make sure tag strings are not part of element type
	typ = ToStruct(ReflectOn(struct {
		d []uint32 `reflect:"TAG"`
	}{})).Field(0).Type
	testType(t, 14, typ, "[]uint32")
}

func TestInterfaceGet(t *testing.T) {
	var inter struct {
		E interface{}
	}
	inter.E = 123.456
	v1 := ReflectOnPtr(&inter)
	v2 := ToStruct(v1).Field(0)
	assert(t, v2.Type.String(), "interface {}")
	i2 := v2.Interface()
	v3 := ReflectOn(i2)
	assert(t, v3.Type.String(), "float64")
}

func TestInterfaceValue(t *testing.T) {
	var inter struct {
		E interface{}
	}
	inter.E = 123.456
	v1 := ReflectOnPtr(&inter)
	v2 := ToStruct(v1).Field(0)
	assert(t, v2.Type.String(), "interface {}")
	v3 := v2.Iface()
	assert(t, v3.Type.String(), "float64")

	i3 := v2.Interface()
	if _, ok := i3.(float64); !ok {
		t.Error("v2.Interface() did not return float64, got ", ReflectOn(i3))
	}
}

func TestFunctionValue(t *testing.T) {
	var x interface{} = func() {}
	v := ReflectOn(x)
	if fmt.Sprint(v.Interface()) != fmt.Sprint(x) {
		t.Errorf("TestFunction returned wrong pointer")
	}
	assert(t, v.Type.String(), "func()")
}

func TestPtrPointTo(t *testing.T) {
	var ip *int32
	var i int32 = 1234
	vip := ReflectOnPtr(&ip)
	vi := ReflectOnPtr(&i)
	vip.Set(vi.Addr())
	if *ip != 1234 {
		t.Errorf("got %d, want 1234", *ip)
	}

	ip = nil
	vp := ReflectOnPtr(&ip)
	vp.Set(Zero(vp.Type))
	if ip != nil {
		t.Errorf("got non-nil (%p), want nil", ip)
	}
}

func TestPtrSetNil(t *testing.T) {
	var i int32 = 1234
	ip := &i
	vip := ReflectOnPtr(&ip)
	vip.Set(Zero(vip.Type))
	if ip != nil {
		t.Errorf("got non-nil (%d), want nil", *ip)
	}
}

func TestMapSetNil(t *testing.T) {
	m := make(map[string]int)
	vm := ReflectOnPtr(&m)
	vm.Set(Zero(vm.Type))
	if m != nil {
		t.Errorf("got non-nil (%p), want nil", m)
	}
}

func TestAppend(t *testing.T) {
	for i, test := range appendTests {
		origLen, extraLen := len(test.orig), len(test.extra)
		want := append(test.orig, test.extra...)
		// Convert extra from []int to []Value.
		e0 := make([]Value, len(test.extra))
		for j, e := range test.extra {
			e0[j] = ReflectOn(e)
		}
		// Convert extra from []int to *Iterable.
		e1 := ToSlice(ReflectOn(test.extra))
		// Test Append.
		a0 := ToSlice(ReflectOn(test.orig))
		have0 := a0.Append(e0...).Interface().([]int)
		if !sameInts(have0, want) {
			t.Errorf("Append #%d: have %v, want %v (%p %p)", i, have0, want, test.orig, have0)
		}
		// Check that the orig and extra slices were not modified.
		if len(test.orig) != origLen {
			t.Errorf("Append #%d origLen: have %v, want %v", i, len(test.orig), origLen)
		}
		if len(test.extra) != extraLen {
			t.Errorf("Append #%d extraLen: have %v, want %v", i, len(test.extra), extraLen)
		}

		// Test AppendSlice.
		a1 := ToSlice(ReflectOn(test.orig))
		have1 := a1.AppendWithSlice(e1).Interface().([]int)
		if !sameInts(have1, want) {
			t.Errorf("AppendSlice #%d: have %v, want %v", i, have1, want)
		}
		// Check that the orig and extra slices were not modified.
		if len(test.orig) != origLen {
			t.Errorf("AppendSlice #%d origLen: have %v, want %v", i, len(test.orig), origLen)
		}
		if len(test.extra) != extraLen {
			t.Errorf("AppendSlice #%d extraLen: have %v, want %v", i, len(test.extra), extraLen)
		}
	}
}

func TestCopy(t *testing.T) {
	a := []int{1, 2, 3, 4, 10, 9, 8, 7}
	b := []int{11, 22, 33, 44, 1010, 99, 88, 77, 66, 55, 44}
	c := []int{11, 22, 33, 44, 1010, 99, 88, 77, 66, 55, 44}
	for i := 0; i < len(b); i++ {
		if b[i] != c[i] {
			t.Errorf("b != c before test")
		}
	}
	a1 := a
	b1 := b
	aa := ToSlice(ReflectOnPtr(&a1))
	ab := ToSlice(ReflectOnPtr(&b1))
	for tocopy := 1; tocopy <= 7; tocopy++ {
		aa.SetLen(tocopy)
		Copy(ab, aa)
		aa.SetLen(8)
		for i := 0; i < tocopy; i++ {
			if a[i] != b[i] {
				t.Errorf("(i) tocopy=%d a[%d]=%d, b[%d]=%d",
					tocopy, i, a[i], i, b[i])
			}
		}
		for i := tocopy; i < len(b); i++ {
			if b[i] != c[i] {
				if i < len(a) {
					t.Errorf("(ii) tocopy=%d a[%d]=%d, b[%d]=%d, c[%d]=%d",
						tocopy, i, a[i], i, b[i], i, c[i])
				} else {
					t.Errorf("(iii) tocopy=%d b[%d]=%d, c[%d]=%d",
						tocopy, i, b[i], i, c[i])
				}
			} else {
				// t.Logf("tocopy=%d elem %d is okay\n", tocopy, i)
			}
		}
	}
}

func TestCopyString(t *testing.T) {
	t.Run("Slice", func(t *testing.T) {
		s := bytes.Repeat([]byte{'_'}, 8)
		val := ToSlice(ReflectOn(s))

		n, ok := Copy(val, ToSlice(ReflectOn("")))
		if ok {
			if expecting := []byte("________"); n != 0 || !bytes.Equal(s, expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 0, s = %s", n, s, expecting)
			}
		} else {
			t.Error("Copy failed")
		}

		n, ok = Copy(val, ToSlice(ReflectOn("hello")))
		if ok {
			if expecting := []byte("hello___"); n != 5 || !bytes.Equal(s, expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 5, s = %s", n, s, expecting)
			}
		} else {
			t.Error("Copy failed")
		}

		n, ok = Copy(val, ToSlice(ReflectOn("helloworld")))
		if ok {
			if expecting := []byte("hellowor"); n != 8 || !bytes.Equal(s, expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 8, s = %s", n, s, expecting)
			}
		} else {
			t.Error("Copy failed")
		}
	})
	t.Run("Array", func(t *testing.T) {
		s := [...]byte{'_', '_', '_', '_', '_', '_', '_', '_'}
		val := ToSlice(ReflectOnPtr(&s))

		n, ok := Copy(val, ToSlice(ReflectOn("")))
		if ok {
			if expecting := []byte("________"); n != 0 || !bytes.Equal(s[:], expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 0, s = %s", n, s[:], expecting)
			}
		} else {
			t.Error("Copy failed")
		}

		n, ok = Copy(val, ToSlice(ReflectOn("hello")))
		if ok {
			if expecting := []byte("hello___"); n != 5 || !bytes.Equal(s[:], expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 5, s = %s", n, s[:], expecting)
			}
		} else {
			t.Error("Copy failed")
		}
		n, ok = Copy(val, ToSlice(ReflectOn("helloworld")))
		if ok {
			if expecting := []byte("hellowor"); n != 8 || !bytes.Equal(s[:], expecting) {
				t.Errorf("got n = %d, s = %s, expecting n = 8, s = %s", n, s[:], expecting)
			}
		} else {
			t.Error("Copy failed")
		}
	})
}

func TestCopyArray(t *testing.T) {
	a := [8]int{1, 2, 3, 4, 10, 9, 8, 7}
	b := [11]int{11, 22, 33, 44, 1010, 99, 88, 77, 66, 55, 44}
	c := b
	aa := ToSlice(ReflectOnPtr(&a))
	ab := ToSlice(ReflectOnPtr(&b))
	_, ok := Copy(ab, aa)
	if !ok {
		t.Fatal("Copy failed.")
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			t.Errorf("(i) a[%d]=%d, b[%d]=%d", i, a[i], i, b[i])
		}
	}
	for i := len(a); i < len(b); i++ {
		if b[i] != c[i] {
			t.Errorf("(ii) b[%d]=%d, c[%d]=%d", i, b[i], i, c[i])
		} else {
			// t.Logf("elem %d is okay\n", i)
		}
	}
}

func TestBigUnnamedStruct(t *testing.T) {
	b := struct{ a, b, c, d int64 }{1, 2, 3, 4}
	v := ReflectOn(b)
	b1 := v.Interface().(struct {
		a, b, c, d int64
	})
	if b1.a != b.a || b1.b != b.b || b1.c != b.c || b1.d != b.d {
		t.Errorf("ReflectOn(%v).Interface().(*Big) = %v", b, b1)
	}
}

func TestBigStruct(t *testing.T) {
	b := big{1, 2, 3, 4, 5}
	v := ReflectOn(b)
	b1 := v.Interface().(big)
	if b1.a != b.a || b1.b != b.b || b1.c != b.c || b1.d != b.d || b1.e != b.e {
		t.Errorf("ReflectOn(%v).Interface().(big) = %v", b, b1)
	}
}

func TestDeepEqual(t *testing.T) {
	for idx, test := range deepEqualTests {
		if test.b == (self{}) {
			test.b = test.a
		}
		// t.Logf("Testing %d : %#v == %#v", idx, test.a, test.b)
		if r := DeepEqual(test.a, test.b); r != test.eq {
			t.Errorf("%d. DeepEqual(%v, %v) = %v, want %v", idx, test.a, test.b, r, test.eq)
		}
	}
}

func TestDeepEqualRecursiveStruct(t *testing.T) {
	a, b := new(Recursive), new(Recursive)
	*a = Recursive{12, a}
	*b = Recursive{12, b}
	if !DeepEqual(a, b) {
		t.Error("DeepEqual(recursive same) = false, want true")
	}
}
func TestDeepEqualComplexStruct(t *testing.T) {
	m := make(map[float64]float64)
	stra, strb := "hello", "hello"
	a, b := new(_Complex), new(_Complex)
	*a = _Complex{5, [3]*_Complex{a, b, a}, &stra, m}
	*b = _Complex{5, [3]*_Complex{b, a, a}, &strb, m}
	if !DeepEqual(a, b) {
		t.Error("DeepEqual(complex same) = false, want true")
	}
}

func TestDeepEqualComplexStructInequality(t *testing.T) {
	m := make(map[float64]float64)
	stra, strb := "hello", "helloo" // Difference is here
	a, b := new(_Complex), new(_Complex)
	*a = _Complex{5, [3]*_Complex{a, b, a}, &stra, m}
	*b = _Complex{5, [3]*_Complex{b, a, a}, &strb, m}
	if DeepEqual(a, b) {
		t.Error("DeepEqual(complex different) = true, want false")
	}
}

func TestDeepEqualUnexportedMap(t *testing.T) {
	// Check that DeepEqual can look at unexported fields.
	x1 := UnexpT{map[int]int{1: 2}}
	x2 := UnexpT{map[int]int{1: 2}}
	if !DeepEqual(&x1, &x2) {
		t.Error("DeepEqual(x1, x2) = false, want true")
	}

	y1 := UnexpT{map[int]int{2: 3}}
	if DeepEqual(&x1, &y1) {
		t.Error("DeepEqual(x1, y1) = true, want false")
	}
}

func TestValueOf(t *testing.T) {
	// Special case for nil
	if typ := ReflectOn(nil).Type; typ != nil {
		t.Errorf("expected nil type for nil value; got %v", typ)
	}
	for _, test := range deepEqualTests {
		v := ReflectOn(test.a)
		if !v.IsValid() {
			continue
		}
		typ := ReflectOn(test.a).Type
		if typ != v.Type {
			t.Errorf("ReflectOn(%v) = %v, but ReflectOn(%v).Type() = %v", test.a, typ, test.a, v.Type)
		}
	}
}

func Nil(a interface{}, t *testing.T) {
	n := ToStruct(ReflectOn(a)).Field(0)
	if !n.IsNil() {
		t.Errorf("%v should be nil", a)
	}
}

func NotNil(a interface{}, t *testing.T) {
	n := ToStruct(ReflectOn(a)).Field(0)
	if n.IsNil() {
		t.Errorf("value of type %v should not be nil", ReflectOn(a).Type.String())
	}
}

func TestIsNil(t *testing.T) {
	// These implement IsNil.
	// Wrap in extra struct to hide interface type.
	doNil := []interface{}{
		struct{ x *int }{},
		struct{ x interface{} }{},
		struct{ x map[string]int }{},
		struct{ x func() bool }{},
		struct{ x chan int }{},
		struct{ x []string }{},
	}
	for _, ts := range doNil {
		ty := ToStruct(ReflectOn(ts)).Field(0).Type
		v := Zero(ty)
		v.IsNil() // panics if not okay to call
	}

	// Check the implementations
	var pi struct {
		x *int
	}
	Nil(pi, t)
	pi.x = new(int)
	NotNil(pi, t)

	var si struct {
		x []int
	}
	Nil(si, t)
	si.x = make([]int, 10)
	NotNil(si, t)

	var ci struct {
		x chan int
	}
	Nil(ci, t)
	ci.x = make(chan int)
	NotNil(ci, t)

	var mi struct {
		x map[int]int
	}
	Nil(mi, t)
	mi.x = make(map[int]int)
	NotNil(mi, t)

	var ii struct {
		x interface{}
	}
	Nil(ii, t)
	ii.x = 2
	NotNil(ii, t)

	var fi struct {
		x func(t *testing.T)
	}
	Nil(fi, t)
	fi.x = TestIsNil
	NotNil(fi, t)
}

func TestInterfaceExtraction(t *testing.T) {
	var s struct {
		W io.Writer
	}

	s.W = os.Stdout
	v := ToStruct(ReflectOnPtr(&s)).Field(0).Interface()
	if v != s.W.(interface{}) {
		t.Error("Interface() on interface: ", v, s.W)
	} else {
		t.Logf("%#v", v)
	}
}

func TestNilPtrValueSub(t *testing.T) {
	var pi *int
	if pv := ReflectOnPtr(pi); pv.IsValid() {
		t.Error("ReflectOn((*int)(nil)).IsValid()")
	}
}

func TestMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	mv := ToMap(ReflectOn(m))
	if n := mv.Len(); n != len(m) {
		t.Errorf("Len = %d, want %d", n, len(m))
	}
	keys := mv.MapKeys()
	newmap := MakeMap(mv.Type)
	for k, v := range m {
		// Check that returned Keys match keys in range.
		// These aren't required to be in the same order.
		seen := false
		for _, kv := range keys {
			if kv.String().Get() == k {
				seen = true
				break
			}
		}
		if !seen {
			t.Errorf("Missing key %q", k)
		}

		// Check that value lookup is correct.
		vv := mv.MapIndex(ReflectOn(k))
		if vi := vv.Int().Get(); vi != int64(v) {
			t.Errorf("Key %q: have value %d, want %d", k, vi, v)
		}

		// Copy into new map.
		newmap.SetMapIndex(ReflectOn(k), ReflectOn(v))
	}
	vv := mv.MapIndex(ReflectOn("not-present"))
	if vv.IsValid() {
		t.Errorf("Invalid key: got non-nil value %s", ValueToString(vv))
	}

	newm := newmap.Interface().(map[string]int)
	if len(newm) != len(m) {
		t.Errorf("length after copy: newm=%d, m=%d", len(newm), len(m))
	}

	for k, v := range newm {
		mv, ok := m[k]
		if mv != v {
			t.Errorf("newm[%q] = %d, but m[%q] = %d, %v", k, v, k, mv, ok)
		}
	}

	newmap.SetMapIndex(ReflectOn("a"), Value{})
	v, ok := newm["a"]
	if ok {
		t.Errorf("newm[\"a\"] = %d after delete", v)
	}

	mv = ToMap(ReflectOnPtr(&m))
	mv.Set(Zero(mv.Type))
	if m != nil {
		t.Errorf("mv.Set(nil) failed")
	}
}

func TestNilMap(t *testing.T) {
	var m map[string]int
	mv := ToMap(ReflectOn(m))
	keys := mv.MapKeys()
	if len(keys) != 0 {
		t.Errorf(">0 keys for nil map: %v", keys)
	}

	// Check that value for missing key is zero.
	x := mv.MapIndex(ReflectOn("hello"))
	if x.Kind() != Invalid {
		t.Errorf("m.MapIndex(\"hello\") for nil map = %v, want Invalid Value", x)
	}

	// Check big value too.
	var mbig map[string][10 << 20]byte
	x = ToMap(ReflectOn(mbig)).MapIndex(ReflectOn("hello"))
	if x.Kind() != Invalid {
		t.Errorf("mbig.MapIndex(\"hello\") for nil map = %v, want Invalid Value", x)
	}

	// Test that deletes from a nil map succeed.
	mv.SetMapIndex(ReflectOn("hi"), Value{})
}

func TestCallConvert(t *testing.T) {
	v := ReflectOnPtr(new(io.ReadWriter))
	f := ReflectOn(func(r io.Reader) io.Reader { return r })
	out, ok := f.Call([]Value{v})
	if !ok {
		t.Error("Call failed")
	}
	if len(out) != 1 || out[0].Type != ReflectOn(new(io.Reader)).Type.Deref() || !out[0].IsNil() {
		t.Errorf("expected [nil], got %v", out)
	}
}

func TestFunc(t *testing.T) {
	ret, ok := ReflectOn(dummy).Call([]Value{
		ReflectOn(byte(10)),
		ReflectOn(20),
		ReflectOn(byte(30)),
		ReflectOn(two{40, 50}),
		ReflectOn(byte(60)),
		ReflectOn(float32(70)),
		ReflectOn(byte(80)),
	})
	if !ok {
		t.Error("Call failed")
	}
	if len(ret) != 7 {
		t.Errorf("Call returned %d values, want 7", len(ret))
	}

	i := byte(ret[0].Uint().Get())
	j := int(ret[1].Int().Get())
	k := byte(ret[2].Uint().Get())
	l := ret[3].Interface().(two)
	m := byte(ret[4].Uint().Get())
	n := float32(ret[5].Float().Get())
	o := byte(ret[6].Uint().Get())

	if i != 10 || j != 20 || k != 30 || l != (two{40, 50}) || m != 60 || n != 70 || o != 80 {
		t.Errorf("Call returned %d, %d, %d, %v, %d, %g, %d; want 10, 20, 30, [40, 50], 60, 70, 80", i, j, k, l, m, n, o)
	}

	for i, v := range ret {
		if v.CanAddr() {
			t.Errorf("result %d is addressable", i)
		}
	}
}

func TestCallWithStruct(t *testing.T) {
	r, ok := ReflectOn(returnEmpty).Call(nil)
	if !ok {
		t.Error("Call failed")
	}
	if len(r) != 1 || r[0].Type != ReflectOn(emptyStruct{}).Type {
		t.Errorf("returning empty struct returned %#v instead", r)
	}
	r, ok = ReflectOn(takesEmpty).Call([]Value{ReflectOn(emptyStruct{})})
	if !ok {
		t.Error("Call failed")
	}
	if len(r) != 0 {
		t.Errorf("takesEmpty returned values: %#v", r)
	}
	r, ok = ReflectOn(returnNonEmpty).Call([]Value{ReflectOn(42)})
	if !ok {
		t.Error("Call failed")
	}
	as := ToStruct(r[0])
	if len(r) != 1 || r[0].Type != ReflectOn(nonEmptyStruct{}).Type || as.Field(0).Int().Get() != 42 {
		t.Errorf("returnNonEmpty returned %#v", r)
	}
	r, ok = ReflectOn(takesNonEmpty).Call([]Value{ReflectOn(nonEmptyStruct{member: 42})})
	if !ok {
		t.Error("Call failed")
	}
	if len(r) != 1 || r[0].Type != ReflectOn(1).Type || r[0].Int().Get() != 42 {
		t.Errorf("takesNonEmpty returned %#v", r)
	}
}

func TestCallReturnsEmpty(t *testing.T) {
	// Issue 21717: past-the-end pointer write in Call with
	// nonzero-sized frame and zero-sized return value.
	runtime.GC()
	var finalized uint32
	f := func() (emptyStruct, *int) {
		i := new(int)
		runtime.SetFinalizer(i, func(*int) { atomic.StoreUint32(&finalized, 1) })
		return emptyStruct{}, i
	}
	cl, ok := ReflectOn(f).Call(nil)
	if !ok {
		t.Error("Call failed")
	}
	v := cl[0] // out[0] should not alias out[1]'s memory, so the finalizer should run.
	timeout := time.After(5 * time.Second)
	for atomic.LoadUint32(&finalized) == 0 {
		select {
		case <-timeout:
			t.Error("finalizer did not run")
		default:
		}
		runtime.Gosched()
		runtime.GC()
	}
	runtime.KeepAlive(v)
}

func TestMethod(t *testing.T) {
	p := Point{3, 4}

	// Curried method of value.
	tfunc := ReflectOn((func(int) int)(nil)).Type
	v := ToStruct(ReflectOn(p)).Method(1)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Value Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok := v.Call([]Value{ReflectOn(14)})
	if !ok {
		t.Errorf("Call failed")
	}
	i := cl[0].Int().Get()
	if i != 350 {
		t.Errorf("Value Method returned %d; want 350", i)
	}

	v = ToStruct(ReflectOn(p)).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Value MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = v.Call([]Value{ReflectOn(15)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 375 {
		t.Errorf("Value MethodByName returned %d; want 375", i)
	}

	v = ToStruct(ReflectOn(p)).MethodByName("NoArgs")
	v.Call(nil)

	// Curried method of pointer.
	v = ToStruct(ReflectOn(&p)).Method(1)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Value Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok = v.Call([]Value{ReflectOn(16)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 400 {
		t.Errorf("Pointer Value Method returned %d; want 400", i)
	}

	v = ToStruct(ReflectOn(&p)).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Value MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = v.Call([]Value{ReflectOn(17)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 425 {
		t.Errorf("Pointer Value MethodByName returned %d; want 425", i)
	}

	v = ToStruct(ReflectOn(&p)).MethodByName("NoArgs")
	v.Call(nil)

	// Curried method of interface value.
	// Have to wrap interface value in a struct to get at it.
	// Passing it to ValueOf directly would
	// access the underlying Point, not the interface.
	var x interface {
		Dist(int) int
	} = p
	pv := ToStruct(ReflectOnPtr(&x).Iface())
	// t.Logf("%d methods.", pv.NumMethod())
	v = pv.Method(0)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Interface Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok = v.Call([]Value{ReflectOn(18)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != -1 {
		t.Errorf("Interface Method returned %d; want -1", i)
	}
	v = pv.MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Interface MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = v.Call([]Value{ReflectOn(19)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 475 {
		t.Errorf("Interface MethodByName returned %d; want 475", i)
	}

}

func TestMethodValue(t *testing.T) {
	p := Point{3, 4}
	var i int64

	// Curried method of value.
	tfunc := ReflectOn((func(int) int)(nil)).Type

	v := ToStruct(ReflectOn(p)).Method(1)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Value Method Type is %s; want %s", tt, tfunc)
	}

	cl, ok := ReflectOn(v.Interface()).Call([]Value{ReflectOn(10)})
	if !ok {
		t.Errorf("Call failed")
	}

	i = cl[0].Int().Get()
	if i != 250 {
		t.Errorf("Value Method returned %d; want 250", i)
	}
	v = ToStruct(ReflectOn(p)).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Value MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(11)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 275 {
		t.Errorf("Value MethodByName returned %d; want 275", i)
	}
	v = ToStruct(ReflectOn(p)).MethodByName("NoArgs")
	ReflectOn(v.Interface()).Call(nil)
	v.Interface().(func())()

	// Curried method of pointer.
	v = ToStruct(ReflectOn(&p)).Method(1)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Value Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(12)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 300 {
		t.Errorf("Pointer Value Method returned %d; want 300", i)
	}
	v = ToStruct(ReflectOn(&p)).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Value MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(13)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 325 {
		t.Errorf("Pointer Value MethodByName returned %d; want 325", i)
	}
	v = ToStruct(ReflectOn(&p)).MethodByName("NoArgs")
	ReflectOn(v.Interface()).Call(nil)
	v.Interface().(func())()

	// Curried method of pointer to pointer.
	pp := &p
	v = ToStruct(ReflectOnPtr(&pp)).Method(1)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Pointer Value Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(14)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 350 {
		t.Errorf("Pointer Pointer Value Method returned %d; want 350", i)
	}
	v = ToStruct(ReflectOnPtr(&pp)).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Pointer Pointer Value MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(15)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 375 {
		t.Errorf("Pointer Pointer Value MethodByName returned %d; want 375", i)
	}

	// Curried method of interface value.
	// Have to wrap interface value in a struct to get at it.
	// Passing it to ValueOf directly would
	// access the underlying Point, not the interface.
	var s = struct {
		X interface {
			Dist(int) int
		}
	}{p}
	pv := ToStruct(ReflectOn(s)).Field(0)
	v = ToStruct(pv.Iface()).Method(0)
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Interface Method Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(16)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != -1 {
		t.Errorf("Interface Method returned %d; want -1", i)
	}
	v = ToStruct(pv.Iface()).MethodByName("Dist")
	if tt := v.MethodType(); tt != tfunc {
		t.Errorf("Interface MethodByName Type is %s; want %s", tt, tfunc)
	}
	cl, ok = ReflectOn(v.Interface()).Call([]Value{ReflectOn(17)})
	if !ok {
		t.Errorf("Call failed")
	}
	i = cl[0].Int().Get()
	if i != 425 {
		t.Errorf("Interface MethodByName returned %d; want 425", i)
	}
}

// Reflect version of $GOROOT/test/method5.go

// Concrete types implementing M method.
// Smaller than a word, word-sized, larger than a word.
// Value and pointer receivers.
func TestMethod5(t *testing.T) {
	CheckF := func(name string, f func(int, byte) (byte, int), inc int) {
		b, x := f(1000, 99)
		if b != 99 || x != 1000+inc {
			t.Errorf("%s(1000, 99) = %v, %v, want 99, %v", name, b, x, 1000+inc)
		}
	}

	CheckV := func(name string, i Value, inc int) {
		m := ToStruct(i).Method(0)
		bx, ok := m.Call([]Value{ReflectOn(1000), ReflectOn(byte(99))})
		if !ok {
			t.Error("Failed to call.")
		}
		b := bx[0].Interface()
		x := bx[1].Interface()
		if b != byte(99) || x != 1000+inc {
			t.Errorf("direct %s.M(1000, 99) = %v, %v, want 99, %v", name, b, x, 1000+inc)
		}

		CheckF(name+".M", ToStruct(i).Method(0).Interface().(func(int, byte) (byte, int)), inc)
	}

	var TinterType = TypeOf(new(Tinter)).Deref()

	CheckI := func(name string, i interface{}, inc int) {
		v := ReflectOn(i)
		CheckV(name, v, inc)
		CheckV("(i="+name+")", Convert(v, TinterType), inc)
	}

	sv := Tsmallv(1)
	CheckI("sv", sv, 1)
	CheckI("&sv", &sv, 1)

	sp := Tsmallp(2)
	CheckI("&sp", &sp, 2)

	wv := Twordv(3)
	CheckI("wv", wv, 3)
	CheckI("&wv", &wv, 3)

	wp := Twordp(4)
	CheckI("&wp", &wp, 4)

	bv := Tbigv([2]uintptr{5, 6})
	CheckI("bv", bv, 11)
	CheckI("&bv", &bv, 11)

	bp := Tbigp([2]uintptr{7, 8})
	CheckI("&bp", &bp, 15)

	t4 := Tm4{}
	t3 := Tm3{&t4}
	t2 := Tm2{&t3}
	t1 := Tm1{t2}
	CheckI("t4", t4, 40)
	CheckI("&t4", &t4, 40)
	CheckI("t3", t3, 40)
	CheckI("&t3", &t3, 40)
	CheckI("t2", t2, 40)
	CheckI("&t2", &t2, 40)
	CheckI("t1", t1, 40)
	CheckI("&t1", &t1, 40)
}

func TestInterfaceSet(t *testing.T) {
	p := &Point{3, 4}

	var s struct {
		I interface{}
		P interface {
			Dist(int) int
		}
	}
	sv := ToStruct(ReflectOnPtr(&s))
	sv.Field(0).Set(ReflectOn(p))
	if q := s.I.(*Point); q != p {
		t.Errorf("i: have %p want %p", q, p)
	}

	pv := sv.Field(1)
	pv.Set(ReflectOn(p))
	if q := s.P.(*Point); q != p {
		t.Errorf("i: have %p want %p", q, p)
	}
	cl, ok := ToStruct(pv.Iface()).Method(0).Call([]Value{ReflectOn(10)})
	if !ok {
		t.Error("Call failed")
	}
	i := cl[0].Int().Get()
	if i != -1 {
		t.Errorf("Interface Method returned %d; want -1", i)
	}
}

func TestImportPath(t *testing.T) {
	tests := []struct {
		t    *RType
		path string
	}{
		{ReflectOn(&base64.Encoding{}).Type.Deref(), "encoding/base64"},
		{ReflectOn(int(0)).Type, ""},
		{ReflectOn(int8(0)).Type, ""},
		{ReflectOn(int16(0)).Type, ""},
		{ReflectOn(int32(0)).Type, ""},
		{ReflectOn(int64(0)).Type, ""},
		{ReflectOn(uint(0)).Type, ""},
		{ReflectOn(uint8(0)).Type, ""},
		{ReflectOn(uint16(0)).Type, ""},
		{ReflectOn(uint32(0)).Type, ""},
		{ReflectOn(uint64(0)).Type, ""},
		{ReflectOn(uintptr(0)).Type, ""},
		{ReflectOn(float32(0)).Type, ""},
		{ReflectOn(float64(0)).Type, ""},
		{ReflectOn(complex64(0)).Type, ""},
		{ReflectOn(complex128(0)).Type, ""},
		{ReflectOn(byte(0)).Type, ""},
		{ReflectOn(rune(0)).Type, ""},
		{ReflectOn([]byte(nil)).Type, ""},
		{ReflectOn([]rune(nil)).Type, ""},
		{ReflectOn(string("")).Type, ""},
		{ReflectOn((*interface{})(nil)).Type.Deref(), ""},
		{ReflectOn((*byte)(nil)).Type, ""},
		{ReflectOn((*rune)(nil)).Type, ""},
		{ReflectOn((*int64)(nil)).Type, ""},
		{ReflectOn(map[string]int{}).Type, ""},
		{ReflectOn((*error)(nil)).Type.Deref(), ""},
		{ReflectOn((*Point)(nil)).Type, ""},
		{ReflectOn((*Point)(nil)).Type.Deref(), testPackageName},
	}
	for _, test := range tests {
		if path := test.t.PkgPath(); !strings.HasSuffix(path, test.path) {
			t.Errorf("%v.PkgPath() = %q, want %q", test.t, path, test.path)
		}
	}
}

func TestEmbeddedMethods(t *testing.T) {
	i := &InnerInt{3}
	stru := ToStruct(ReflectOn(i))
	m := stru.Method(0)
	cl, ok := m.Call(nil)
	if ok {
		if v := cl[0].Int().Get(); v != 3 {
			t.Errorf("i.M() = %d, want 3", v)
		}
	} else {
		t.Error("Failed calling")
	}

	m = stru.Method(1)
	cl, ok = m.Call(nil)
	if ok {
		if v := cl[0].Int().Get(); v != 3 {
			t.Errorf("i.M() = %d, want 3", v)
		}
	} else {
		t.Error("Failed calling")
	}

	o := &OuterInt{1, InnerInt{2}}
	m = ToStruct(ReflectOn(o)).Method(0)

	println("Calling first method.")
	cl, ok = m.Call(nil)
	if ok {
		if v := cl[0].Int().Get(); v != 2 {
			t.Errorf("i.M() = %d, want 2", v)
		}
	} else {
		t.Error("Failed getting method")
	}

	println("Calling M method.")
	f := (*OuterInt).M
	if v := f(o); v != 2 {
		t.Errorf("f(o) = %d, want 2", v)
	}

}

func TestNumMethodOnDDD(t *testing.T) {
	rv := ToStruct(ReflectOn((FuncDDD)(nil)))
	if n := rv.NumMethod(); n != 1 {
		t.Errorf("NumMethod()=%d, want 1", n)
	}
}

func TestPtrTo(t *testing.T) {
	// This block of code means that the ptrToThis field of the
	// reflect data for *unsafe.Pointer is non zero, see
	// https://golang.org/issue/19003
	var x unsafe.Pointer
	var y = &x
	var z = &y

	var i int

	typ := ReflectOn(z).Type
	orig := ReflectOn(z).Type
	// t.Logf("Before test : %#v", typ)

	for i = 0; i < 100; i++ {
		typ = typ.PtrTo()
	}

	for i = 0; i < 100; i++ {
		typ = typ.Deref()
	}

	// t.Logf("After test : %#v", typ)
	if typ != orig {
		t.Errorf("after 100 PtrTo, have %s, want %s", typ, orig)
	}
}

func TestPtrToGC(t *testing.T) {
	type T *uintptr
	tt := ReflectOn(T(nil)).Type
	pt := tt.PtrTo()
	const n = 100
	var x []interface{}
	for i := 0; i < n; i++ {
		v := New(pt)
		p := new(*uintptr)
		*p = new(uintptr)
		**p = uintptr(i)
		v.Deref().Set(Convert(ReflectOn(p), pt))
		x = append(x, v.Interface())
	}
	runtime.GC()

	for i, xi := range x {
		k := ReflectOnPtr(xi).Deref().Deref().Interface().(uintptr)
		if k != uintptr(i) {
			t.Errorf("lost x[%d] = %d, want %d", i, k, i)
		}
	}
}

func TestAddr(t *testing.T) {
	var p struct {
		X, Y int
	}

	v := ReflectOnPtr(&p)
	v = v.Addr()
	v = v.Deref()
	v = ToStruct(v).Field(0)
	v.Int().Set(2)
	if p.X != 2 {
		t.Errorf("Addr.Deref.Set failed to set value")
	}

	// Again but take address of the ReflectOn value.
	// Exercises generation of PtrTypes not present in the binary.
	q := &p
	v = ReflectOnPtr(&q)
	v = v.Addr()
	v = v.Deref()
	v = v.Deref()
	v = v.Addr()
	v = v.Deref()
	v = ToStruct(v).Field(0)
	v.Int().Set(3)
	if p.X != 3 {
		t.Errorf("Addr.Deref.Set failed to set value")
	}

	// Starting without pointer we should get changed value
	// in interface.
	qq := p
	v = ReflectOnPtr(&qq)
	v0 := v
	v = v.Addr()
	v = v.Deref()
	v = ToStruct(v).Field(0)
	v.Int().Set(4)
	if p.X != 3 { // should be unchanged from last time
		t.Errorf("somehow value Set changed original p")
	}
	p = v0.Interface().(struct {
		X, Y int
	})
	if p.X != 4 {
		t.Errorf("Addr.Deref.Set valued to set value in top value")
	}

	// Verify that taking the address of a type gives us a pointer
	// which we can convert back using the usual interface
	// notation.
	var s struct {
		B *bool
	}
	r := ReflectOnPtr(&s)
	sr := ToStruct(r)
	fz := sr.Field(0)
	addr := fz.Addr()
	ps := addr.Interface()
	// shouldn't panic
	*(ps.(**bool)) = new(bool)
	if s.B == nil {
		t.Errorf("Addr.Interface direct assignment failed")
	}
}

func TestAllocations(t *testing.T) {
	noAlloc(t, 100, func(j int) {
		var i interface{}
		var v Value

		// We can uncomment this when compiler escape analysis
		// is good enough to see that the integer assigned to i
		// does not escape and therefore need not be allocated.
		//
		// i = 42 + j
		// v = ReflectOn(i)
		// if int(v.Int()) != 42+j {
		// 	panic("wrong int")
		// }

		i = func(j int) int { return j }
		v = ReflectOn(i)
		if v.Interface().(func(int) int)(j) != j {
			panic("wrong result")
		}
	})
}

func TestSmallNegativeInt(t *testing.T) {
	i := int16(-1)
	v := ReflectOn(i)
	if v.Int().Get() != -1 {
		t.Errorf("int16(-1).Int() returned %v", v.Int())
	}
}

func TestIndex(t *testing.T) {
	xs := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	v := ToSlice(ReflectOn(xs)).Index(3).Interface().(byte)
	if v != xs[3] {
		t.Errorf("xs.Index(3) = %v; expected %v", v, xs[3])
	}
	xa := [8]byte{10, 20, 30, 40, 50, 60, 70, 80}
	v = ToSlice(ReflectOn(xa)).Index(2).Interface().(byte)
	if v != xa[2] {
		t.Errorf("xa.Index(2) = %v; expected %v", v, xa[2])
	}
	s := "0123456789"
	v = ToSlice(ReflectOn(s)).Index(3).Interface().(byte)
	if v != s[3] {
		t.Errorf("s.Index(3) = %v; expected %v", v, s[3])
	}
}

func TestSlice(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	v := ToSlice(ReflectOn(xs)).Slice(3, 5).Interface().([]int)
	if len(v) != 2 {
		t.Errorf("len(xs.Slice(3, 5)) = %d", len(v))
	}
	if cap(v) != 5 {
		t.Errorf("cap(xs.Slice(3, 5)) = %d", cap(v))
	}
	if !DeepEqual(v[0:5], xs[3:]) {
		t.Errorf("xs.Slice(3, 5)[0:5] = %v", v[0:5])
	}

	xa := [8]int{10, 20, 30, 40, 50, 60, 70, 80}
	v = ToSlice(ReflectOnPtr(&xa)).Slice(2, 5).Interface().([]int)
	if len(v) != 3 {
		t.Errorf("len(xa.Slice(2, 5)) = %d", len(v))
	}
	if cap(v) != 6 {
		t.Errorf("cap(xa.Slice(2, 5)) = %d", cap(v))
	}
	if !DeepEqual(v[0:6], xa[2:]) {
		t.Errorf("xs.Slice(2, 5)[0:6] = %v", v[0:6])
	}

	s := "0123456789"
	vs := ToSlice(ReflectOn(s)).Slice(3, 5).Interface().(string)
	if vs != s[3:5] {
		t.Errorf("s.Slice(3, 5) = %q; expected %q", vs, s[3:5])
	}

	rv := ToSlice(ReflectOnPtr(&xs))
	sv := rv.Slice(3, 4)
	ptr2 := sv.Pointer()
	sv = sv.Slice(5, 5)
	ptr3 := sv.Pointer()
	if ptr3 != ptr2 {
		t.Errorf("xs.Slice(3,4).Slice3(5,5).Pointer() = %#x, want %#x", ptr3, ptr2)
	}
}

func TestSlice3(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	v := ToSlice(ReflectOn(xs)).Slice3(3, 5, 7).Interface().([]int)
	if len(v) != 2 {
		t.Errorf("len(xs.Slice3(3, 5, 7)) = %d", len(v))
	}
	if cap(v) != 4 {
		t.Errorf("cap(xs.Slice3(3, 5, 7)) = %d", cap(v))
	}
	if !DeepEqual(v[0:4], xs[3:7:7]) {
		t.Errorf("xs.Slice3(3, 5, 7)[0:4] = %v", v[0:4])
	}
	rv := ToSlice(ReflectOnPtr(&xs))

	xa := [8]int{10, 20, 30, 40, 50, 60, 70, 80}
	v = ToSlice(ReflectOnPtr(&xa)).Slice3(2, 5, 6).Interface().([]int)
	if len(v) != 3 {
		t.Errorf("len(xa.Slice(2, 5, 6)) = %d", len(v))
	}
	if cap(v) != 4 {
		t.Errorf("cap(xa.Slice(2, 5, 6)) = %d", cap(v))
	}
	if !DeepEqual(v[0:4], xa[2:6:6]) {
		t.Errorf("xs.Slice(2, 5, 6)[0:4] = %v", v[0:4])
	}
	rv = ToSlice(ReflectOnPtr(&xa))

	s := "hello world"
	rv = ToSlice(ReflectOnPtr(&s))

	rv = ToSlice(ReflectOnPtr(&xs))
	rv = rv.Slice3(3, 5, 7)
	ptr2 := rv.Pointer()
	rv = rv.Slice3(4, 4, 4)
	ptr3 := rv.Pointer()
	if ptr3 != ptr2 {
		t.Errorf("xs.Slice3(3,5,7).Slice3(4,4,4).Pointer() = %#x, want %#x", ptr3, ptr2)
	}
}

func TestSetLenCap(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	vs := ToSlice(ReflectOnPtr(&xs))

	vs.SetLen(5)
	if len(xs) != 5 || cap(xs) != 8 {
		t.Errorf("after SetLen(5), len, cap = %d, %d, want 5, 8", len(xs), cap(xs))
	}
	vs.SetCap(6)
	if len(xs) != 5 || cap(xs) != 6 {
		t.Errorf("after SetCap(6), len, cap = %d, %d, want 5, 6", len(xs), cap(xs))
	}
	vs.SetCap(5)
	if len(xs) != 5 || cap(xs) != 5 {
		t.Errorf("after SetCap(5), len, cap = %d, %d, want 5, 5", len(xs), cap(xs))
	}
}

func TestFuncArg(t *testing.T) {
	f1 := func(i int, f func(int) int) int { return f(i) }
	f2 := func(i int) int { return i + 1 }
	r, ok := ReflectOn(f1).Call([]Value{ReflectOn(100), ReflectOn(f2)})
	if ok {
		if r[0].Int().Get() != 101 {
			t.Errorf("function returned %d, want 101", r[0].Int())
		}
	} else {
		t.Error("Call failed")
	}
}

func TestStructArg(t *testing.T) {
	type padded struct {
		B string
		C int32
	}
	var (
		gotA  padded
		gotB  uint32
		wantA = padded{"3", 4}
		wantB = uint32(5)
	)
	f := func(a padded, b uint32) {
		gotA, gotB = a, b
	}
	ReflectOn(f).Call([]Value{ReflectOn(wantA), ReflectOn(wantB)})
	if gotA != wantA || gotB != wantB {
		t.Errorf("function called with (%v, %v), want (%v, %v)", gotA, gotB, wantA, wantB)
	}
}

func TestTagGet(t *testing.T) {
	for _, tt := range tagGetTests {
		if v := GetTagNamed(tt.Tag, tt.Key); v != tt.Value {
			t.Errorf("StructTag(%#q).Get(%#q) = %#q, want %#q", tt.Tag, tt.Key, v, tt.Value)
		}
	}
}

func TestBytes(t *testing.T) {
	type B []byte
	x := B{1, 2, 3, 4}
	y := ToSlice(ReflectOn(x)).Bytes()
	if !bytes.Equal(x, y) {
		t.Errorf("ReflectOn(%v).Bytes() = %v", x, y)
	}
	if &x[0] != &y[0] {
		t.Errorf("ReflectOn(%p).Bytes() = %p", &x[0], &y[0])
	}
}

func TestSetBytes(t *testing.T) {
	type B []byte
	var x B
	y := []byte{1, 2, 3, 4}
	ToSlice(ReflectOnPtr(&x)).SetBytes(y)
	if !bytes.Equal(x, y) {
		t.Errorf("ReflectOn(%v).Bytes() = %v", x, y)
	}
	if &x[0] != &y[0] {
		t.Errorf("ReflectOn(%p).Bytes() = %p", &x[0], &y[0])
	}
}

func TestUnexported(t *testing.T) {
	var pub Public
	pub.S = "S"
	pub.T = pub.A[:]
	v := ToStruct(ReflectOnPtr(&pub))
	isValid(v.Field(0))
	isValid(v.Field(1))
	isValid(v.Field(2))
	isValid(v.FieldByName("X"))
	isValid(v.FieldByName("Y"))

	isNonNil(v.Field(0).Interface())
	isNonNil(v.Field(1).Interface())
	isNonNil(ToSlice(ToStruct(v.Field(2)).Field(2)).Index(0))
	isNonNil(v.FieldByName("X").Interface())
	isNonNil(v.FieldByName("Y").Interface())

	var priv Private
	v = ToStruct(ReflectOnPtr(&priv))
	isValid(v.Field(0))
	isValid(v.Field(1))
	isValid(v.FieldByName("x"))
	isValid(v.FieldByName("y"))
}

func TestSetPanic(t *testing.T) {
	ok := func(f func()) { f() }
	bad := func(f func()) {
		defer func() {
			if recover() == nil {
				// yeap, we don't panic anymore
			}
		}()
		f()
	}
	clear := func(v Value) { v.Set(Zero(v.Type)) }

	type t0 struct {
		W int
	}

	type t1 struct {
		Y int
		t0
	}

	type T2 struct {
		Z       int
		namedT0 t0
	}

	type T struct {
		X int
		t1
		T2
		NamedT1 t1
		NamedT2 T2
		namedT1 t1
		namedT2 T2
	}

	// not addressable
	v := ToStruct(ReflectOn(T{}))
	bad(func() { clear(v.Field(0)) })                                       // .X
	bad(func() { clear(v.Field(1)) })                                       // .t1
	bad(func() { clear(ToStruct(v.Field(1)).Field(0)) })                    // .t1.Y
	bad(func() { clear(ToStruct(v.Field(1)).Field(1)) })                    // .t1.t0
	bad(func() { clear(ToStruct(ToStruct(v.Field(1)).Field(1)).Field(0)) }) // .t1.t0.W
	bad(func() { clear(v.Field(2)) })                                       // .T2
	bad(func() { clear(ToStruct(v.Field(2)).Field(0)) })                    // .T2.Z
	bad(func() { clear(ToStruct(v.Field(2)).Field(1)) })                    // .T2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(2)).Field(1)).Field(0)) }) // .T2.namedT0.W
	bad(func() { clear(v.Field(3)) })                                       // .NamedT1
	bad(func() { clear(ToStruct(v.Field(3)).Field(0)) })                    // .NamedT1.Y
	bad(func() { clear(ToStruct(v.Field(3)).Field(1)) })                    // .NamedT1.t0
	bad(func() { clear(ToStruct(ToStruct(v.Field(3)).Field(1)).Field(0)) }) // .NamedT1.t0.W
	bad(func() { clear(v.Field(4)) })                                       // .NamedT2
	bad(func() { clear(ToStruct(v.Field(4)).Field(0)) })                    // .NamedT2.Z
	bad(func() { clear(ToStruct(v.Field(4)).Field(1)) })                    // .NamedT2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(4)).Field(1)).Field(0)) }) // .NamedT2.namedT0.W
	bad(func() { clear(v.Field(5)) })                                       // .namedT1
	bad(func() { clear(ToStruct(v.Field(5)).Field(0)) })                    // .namedT1.Y
	bad(func() { clear(ToStruct(v.Field(5)).Field(1)) })                    // .namedT1.t0
	bad(func() { clear(ToStruct(ToStruct(v.Field(5)).Field(1)).Field(0)) }) // .namedT1.t0.W
	bad(func() { clear(v.Field(6)) })                                       // .namedT2
	bad(func() { clear(ToStruct(v.Field(6)).Field(0)) })                    // .namedT2.Z
	bad(func() { clear(ToStruct(v.Field(6)).Field(1)) })                    // .namedT2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(6)).Field(1)).Field(0)) }) // .namedT2.namedT0.W

	// addressable
	v = ToStruct(ReflectOnPtr(&T{}))
	ok(func() { clear(v.Field(0)) })                                        // .X
	bad(func() { clear(v.Field(1)) })                                       // .t1
	ok(func() { clear(ToStruct(v.Field(1)).Field(0)) })                     // .t1.Y
	bad(func() { clear(ToStruct(v.Field(1)).Field(1)) })                    // .t1.t0
	ok(func() { clear(ToStruct(ToStruct(v.Field(1)).Field(1)).Field(0)) })  // .t1.t0.W
	ok(func() { clear(v.Field(2)) })                                        // .T2
	ok(func() { clear(ToStruct(v.Field(2)).Field(0)) })                     // .T2.Z
	bad(func() { clear(ToStruct(v.Field(2)).Field(1)) })                    // .T2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(2)).Field(1)).Field(0)) }) // .T2.namedT0.W
	ok(func() { clear(v.Field(3)) })                                        // .NamedT1
	ok(func() { clear(ToStruct(v.Field(3)).Field(0)) })                     // .NamedT1.Y
	bad(func() { clear(ToStruct(v.Field(3)).Field(1)) })                    // .NamedT1.t0
	ok(func() { clear(ToStruct(ToStruct(v.Field(3)).Field(1)).Field(0)) })  // .NamedT1.t0.W
	ok(func() { clear(v.Field(4)) })                                        // .NamedT2
	ok(func() { clear(ToStruct(v.Field(4)).Field(0)) })                     // .NamedT2.Z
	bad(func() { clear(ToStruct(v.Field(4)).Field(1)) })                    // .NamedT2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(4)).Field(1)).Field(0)) }) // .NamedT2.namedT0.W
	bad(func() { clear(v.Field(5)) })                                       // .namedT1
	bad(func() { clear(ToStruct(v.Field(5)).Field(0)) })                    // .namedT1.Y
	bad(func() { clear(ToStruct(v.Field(5)).Field(1)) })                    // .namedT1.t0
	bad(func() { clear(ToStruct(ToStruct(v.Field(5)).Field(1)).Field(0)) }) // .namedT1.t0.W
	bad(func() { clear(v.Field(6)) })                                       // .namedT2
	bad(func() { clear(ToStruct(v.Field(6)).Field(0)) })                    // .namedT2.Z
	bad(func() { clear(ToStruct(v.Field(6)).Field(1)) })                    // .namedT2.namedT0
	bad(func() { clear(ToStruct(ToStruct(v.Field(6)).Field(1)).Field(0)) }) // .namedT2.namedT0.W
}

func TestCallPanic(t *testing.T) {
	type t0 interface {
		W()
		w()
	}
	type T1 interface {
		Y()
		y()
	}
	type T2 struct {
		T1
		t0
	}
	type T struct {
		t0 // 0
		T1 // 1

		NamedT0 t0 // 2
		NamedT1 T1 // 3
		NamedT2 T2 // 4

		namedT0 t0 // 5
		namedT1 T1 // 6
		namedT2 T2 // 7
	}
	ok := func(f func()) { f() }
	bad := func(f func()) {
		defer func() {
			if recover() == nil {
				// yeap, we don't panic anymore
			}
		}()
		f()
	}
	call := func(v Value) { v.Call(nil) }

	i := timp(0)
	v := ToStruct(ReflectOn(T{i, i, i, i, T2{i, i}, i, i, T2{i, i}}))
	ok(func() { call(ToStruct(v.Field(0)).Method(0)) })  // .t0.W
	bad(func() { call(ToStruct(v.Field(0)).Method(0)) }) // .t0.W
	bad(func() { call(ToStruct(v.Field(0)).Method(1)) }) // .t0.w
	bad(func() { call(ToStruct(v.Field(0)).Method(2)) }) // .t0.w
	ok(func() { call(ToStruct(v.Field(1)).Method(0)) })  // .T1.Y
	ok(func() { call(ToStruct(v.Field(1)).Method(0)) })  // .T1.Y
	bad(func() { call(ToStruct(v.Field(1)).Method(1)) }) // .T1.y
	bad(func() { call(ToStruct(v.Field(1)).Method(2)) }) // .T1.y

	ok(func() { call(ToStruct(v.Field(2)).Method(0)) })  // .NamedT0.W
	ok(func() { call(ToStruct(v.Field(2)).Method(0)) })  // .NamedT0.W
	bad(func() { call(ToStruct(v.Field(2)).Method(1)) }) // .NamedT0.w
	bad(func() { call(ToStruct(v.Field(2)).Method(2)) }) // .NamedT0.w

	ok(func() { call(ToStruct(v.Field(3)).Method(0)) })  // .NamedT1.Y
	ok(func() { call(ToStruct(v.Field(3)).Method(0)) })  // .NamedT1.Y
	bad(func() { call(ToStruct(v.Field(3)).Method(1)) }) // .NamedT1.y
	bad(func() { call(ToStruct(v.Field(3)).Method(3)) }) // .NamedT1.y

	ok(func() { call(ToStruct(ToStruct(v.Field(4)).Field(0)).Method(0)) })  // .NamedT2.T1.Y
	ok(func() { call(ToStruct(ToStruct(v.Field(4)).Field(0)).Method(0)) })  // .NamedT2.T1.W
	ok(func() { call(ToStruct(ToStruct(v.Field(4)).Field(1)).Method(0)) })  // .NamedT2.t0.W
	bad(func() { call(ToStruct(ToStruct(v.Field(4)).Field(1)).Method(0)) }) // .NamedT2.t0.W

	bad(func() { call(ToStruct(v.Field(5)).Method(0)) }) // .namedT0.W
	bad(func() { call(ToStruct(v.Field(5)).Method(0)) }) // .namedT0.W
	bad(func() { call(ToStruct(v.Field(5)).Method(1)) }) // .namedT0.w
	bad(func() { call(ToStruct(v.Field(5)).Method(2)) }) // .namedT0.w

	bad(func() { call(ToStruct(v.Field(6)).Method(0)) }) // .namedT1.Y
	bad(func() { call(ToStruct(v.Field(6)).Method(0)) }) // .namedT1.Y
	bad(func() { call(ToStruct(v.Field(6)).Method(0)) }) // .namedT1.y
	bad(func() { call(ToStruct(v.Field(6)).Method(0)) }) // .namedT1.y

	bad(func() { call(ToStruct(ToStruct(v.Field(7)).Field(0)).Method(0)) }) // .namedT2.T1.Y
	bad(func() { call(ToStruct(ToStruct(v.Field(7)).Field(0)).Method(0)) }) // .namedT2.T1.W
	bad(func() { call(ToStruct(ToStruct(v.Field(7)).Field(1)).Method(0)) }) // .namedT2.t0.W
	bad(func() { call(ToStruct(ToStruct(v.Field(7)).Field(1)).Method(0)) }) // .namedT2.t0.W
}

func TestAlias(t *testing.T) {
	x := string("hello")
	v := ReflectOnPtr(&x)
	oldvalue := v.Interface()
	v.String().Set("world")
	newvalue := v.Interface()

	if oldvalue != "hello" || newvalue != "world" {
		t.Errorf("aliasing: old=%q new=%q, want hello, world", oldvalue, newvalue)
	}
}

func TestConvert(t *testing.T) {
	canConvert := map[[2]*RType]bool{}
	all := map[*RType]bool{}

	for _, tt := range convertTests {
		t1 := tt.in.Type
		if !t1.ConvertibleTo(t1) {
			t.Errorf("(%s).ConvertibleTo(%s) = false, want true", t1, t1)
			continue
		}

		t2 := tt.out.Type
		if !t1.ConvertibleTo(t2) {
			t.Errorf("(%s).ConvertibleTo(%s) = false, want true", t1, t2)
			continue
		}

		all[t1] = true
		all[t2] = true
		canConvert[[2]*RType{t1, t2}] = true

		// vout1 represents the in value converted to the in type.
		v1 := tt.in
		vout1 := Convert(v1, t1)
		out1 := vout1.Interface()
		if vout1.Type != tt.in.Type || !DeepEqual(out1, tt.in.Interface()) {
			t.Errorf("ReflectOn(%T(%[1]v)).Convert(%s) = %T(%[3]v), want %T(%[4]v)", tt.in.Interface(), t1, out1, tt.in.Interface())
		}

		// vout2 represents the in value converted to the out type.
		vout2 := Convert(v1, t2)
		out2 := vout2.Interface()
		if vout2.Type != tt.out.Type || !DeepEqual(out2, tt.out.Interface()) {
			t.Errorf("ReflectOn(%T(%[1]v)).Convert(%s) = %T(%[3]v), want %T(%[4]v)", tt.in.Interface(), t2, out2, tt.out.Interface())
		}

		// vout3 represents a new value of the out type, set to vout2.  This makes
		// sure the converted value vout2 is really usable as a regular value.
		vout3 := New(t2).Deref()
		vout3.Set(vout2)
		out3 := vout3.Interface()
		if vout3.Type != tt.out.Type || !DeepEqual(out3, tt.out.Interface()) {
			t.Errorf("Set(ReflectOn(%T(%[1]v)).Convert(%s)) = %T(%[3]v), want %T(%[4]v)", tt.in.Interface(), t2, out3, tt.out.Interface())
		}

		if v1.IsRO() {
			t.Errorf("table entry %v is RO, should not be", v1)
		}
		if vout1.IsRO() {
			t.Errorf("self-conversion output %v is RO, should not be", vout1)
		}
		if vout2.IsRO() {
			t.Errorf("conversion output %v is RO, should not be", vout2)
		}
		if vout3.IsRO() {
			t.Errorf("set(conversion output) %v is RO, should not be", vout3)
		}
		if conv := Convert(MakeRO(v1), t1); !conv.IsRO() {
			t.Errorf("RO self-conversion output %v is not RO, should be", v1)
		}
		if conv := Convert(MakeRO(v1), t2); !conv.IsRO() {
			t.Errorf("RO conversion output %v is not RO, should be", v1)
		}
	}

	// Assume that of all the types we saw during the tests,
	// if there wasn't an explicit entry for a conversion between
	// a pair of types, then it's not to be allowed. This checks for
	// things like 'int64' converting to '*int'.
	for t1 := range all {
		for t2 := range all {
			expectOK := t1 == t2 || canConvert[[2]*RType{t1, t2}] || t2.Kind() == Interface && t2.NoOfIfaceMethods() == 0
			if ok := t1.ConvertibleTo(t2); ok != expectOK {
				t.Errorf("(%s).ConvertibleTo(%s) = %v, want %v", t1, t2, ok, expectOK)
			}
		}
	}
}
func TestComparable(t *testing.T) {
	for _, tt := range comparableTests {
		if ok := tt.typ.Comparable(); ok != tt.ok {
			t.Errorf("ReflectOn(%v).Comparable() = %v, want %v", tt.typ, ok, tt.ok)
		}
	}
}

func TestOverflow(t *testing.T) {
	if ovf := ReflectOn(float64(0)).Float().Overflows(1e300); ovf {
		t.Errorf("%v wrongly overflows float64", 1e300)
	}

	maxFloat32 := float64((1<<24 - 1) << (127 - 23))
	if ovf := ReflectOn(float32(0)).Float().Overflows(maxFloat32); ovf {
		t.Errorf("%v wrongly overflows float32", maxFloat32)
	}
	ovfFloat32 := float64((1<<24-1)<<(127-23) + 1<<(127-52))
	if ovf := ReflectOn(float32(0)).Float().Overflows(ovfFloat32); !ovf {
		t.Errorf("%v should overflow float32", ovfFloat32)
	}
	if ovf := ReflectOn(float32(0)).Float().Overflows(-ovfFloat32); !ovf {
		t.Errorf("%v should overflow float32", -ovfFloat32)
	}

	maxInt32 := int64(0x7fffffff)
	if ovf := ReflectOn(int32(0)).Int().Overflows(maxInt32); ovf {
		t.Errorf("%v wrongly overflows int32", maxInt32)
	}
	if ovf := ReflectOn(int32(0)).Int().Overflows(-1 << 31); ovf {
		t.Errorf("%v wrongly overflows int32", -int64(1)<<31)
	}
	ovfInt32 := int64(1 << 31)
	if ovf := ReflectOn(int32(0)).Int().Overflows(ovfInt32); !ovf {
		t.Errorf("%v should overflow int32", ovfInt32)
	}

	maxUint32 := uint64(0xffffffff)
	if ovf := ReflectOn(uint32(0)).Uint().Overflows(maxUint32); ovf {
		t.Errorf("%v wrongly overflows uint32", maxUint32)
	}
	ovfUint32 := uint64(1 << 32)
	if ovf := ReflectOn(uint32(0)).Uint().Overflows(ovfUint32); !ovf {
		t.Errorf("%v should overflow uint32", ovfUint32)
	}
}

func TestArrayOf(t *testing.T) {
	// check construction and use of type not in binary
	tests := []struct {
		n          int
		value      func(i int) interface{}
		comparable bool
		want       string
	}{
		{
			n:          0,
			value:      func(i int) interface{} { type Tint int; return Tint(i) },
			comparable: true,
			want:       "[]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tint int; return Tint(i) },
			comparable: true,
			want:       "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tfloat float64; return Tfloat(i) },
			comparable: true,
			want:       "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tstring string; return Tstring(strconv.Itoa(i)) },
			comparable: true,
			want:       "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tstruct struct{ V int }; return Tstruct{i} },
			comparable: true,
			want:       "[{0} {1} {2} {3} {4} {5} {6} {7} {8} {9}]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tint int; return []Tint{Tint(i)} },
			comparable: false,
			want:       "[[0] [1] [2] [3] [4] [5] [6] [7] [8] [9]]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tint int; return [1]Tint{Tint(i)} },
			comparable: true,
			want:       "[[0] [1] [2] [3] [4] [5] [6] [7] [8] [9]]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tstruct struct{ V [1]int }; return Tstruct{[1]int{i}} },
			comparable: true,
			want:       "[{[0]} {[1]} {[2]} {[3]} {[4]} {[5]} {[6]} {[7]} {[8]} {[9]}]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type Tstruct struct{ V []int }; return Tstruct{[]int{i}} },
			comparable: false,
			want:       "[{[0]} {[1]} {[2]} {[3]} {[4]} {[5]} {[6]} {[7]} {[8]} {[9]}]",
		},
		{
			n:          10,
			value:      func(i int) interface{} { type TstructUV struct{ U, V int }; return TstructUV{i, i} },
			comparable: true,
			want:       "[{0 0} {1 1} {2 2} {3 3} {4 4} {5 5} {6 6} {7 7} {8 8} {9 9}]",
		},
		{
			n: 10,
			value: func(i int) interface{} {
				type TstructUV struct {
					U int
					V float64
				}
				return TstructUV{i, float64(i)}
			},
			comparable: true,
			want:       "[{0 0} {1 1} {2 2} {3 3} {4 4} {5 5} {6 6} {7 7} {8 8} {9 9}]",
		},
	}

	for _, table := range tests {
		at := ArrayOf(ReflectOn(table.value(0)).Type, table.n)
		v := ToSlice(New(at).Deref())
		vok := ToSlice(New(at).Deref())
		vnot := ToSlice(New(at).Deref())
		for i := 0; i < v.Len(); i++ {
			v.Index(i).Set(ReflectOn(table.value(i)))
			vok.Index(i).Set(ReflectOn(table.value(i)))
			j := i
			if i+1 == v.Len() {
				j = i + 1
			}
			vnot.Index(i).Set(ReflectOn(table.value(j))) // make it differ only by last element
		}
		s := fmt.Sprint(v.Interface())
		if s != table.want {
			t.Errorf("constructed array = %s, want %s", s, table.want)
		}

		if table.comparable != at.Comparable() {
			t.Errorf("constructed array (%#v) is comparable=%v, want=%v", v.Interface(), at.Comparable(), table.comparable)
		}
		if table.comparable {
			if table.n > 0 {
				if DeepEqual(vnot.Interface(), v.Interface()) {
					t.Errorf(
						"arrays (%#v) compare ok (but should not)",
						v.Interface(),
					)
				}
			}
			if !DeepEqual(vok.Interface(), v.Interface()) {
				t.Errorf(
					"arrays (%#v) compare NOT-ok (but should)",
					v.Interface(),
				)
			}
		}
	}

	// check that type already in binary is found
	type T int
	checkSameType(t, Zero(ArrayOf(ReflectOn(T(1)).Type, 5)).Interface(), [5]T{})
}

func TestArrayOfGC(t *testing.T) {
	type T *uintptr
	tt := ReflectOn(T(nil)).Type
	const n = 100
	var x []interface{}
	for i := 0; i < n; i++ {
		v := ToSlice(New(ArrayOf(tt, n)).Deref())
		for j := 0; j < v.Len(); j++ {
			p := new(uintptr)
			*p = uintptr(i*n + j)
			v.Index(j).Set(ReflectOn(p))
			Convert(v.Index(j), tt)
		}
		x = append(x, v.Interface())
	}
	runtime.GC()

	for i, xi := range x {
		v := ToSlice(ReflectOn(xi))
		for j := 0; j < v.Len(); j++ {
			k := v.Index(j).Deref().Interface()
			if k != uintptr(i*n+j) {
				t.Errorf("lost x[%d][%d] = %d, want %d", i, j, k, i*n+j)
			}
		}
	}
}

func TestArrayOfAlg(t *testing.T) {
	at := ArrayOf(ReflectOn(byte(0)).Type, 6)
	v1 := ToSlice(New(at).Deref())
	v2 := New(at).Deref()
	if v1.Interface() != v1.Interface() {
		t.Errorf("constructed array %v not equal to itself", v1.Interface())
	}
	v1.Index(5).Set(ReflectOn(byte(1)))
	if i1, i2 := v1.Interface(), v2.Interface(); i1 == i2 {
		t.Errorf("constructed arrays %v and %v should not be equal", i1, i2)
	}

	at = ArrayOf(ReflectOn([]int(nil)).Type, 6)
	v1 = ToSlice(New(at).Deref())
}

func TestArrayOfGenericAlg(t *testing.T) {
	at1 := ArrayOf(ReflectOn(string("")).Type, 5)
	at := ArrayOf(at1, 6)
	v1 := ToSlice(New(at).Deref())
	v2 := ToSlice(New(at).Deref())
	if v1.Interface() != v1.Interface() {
		t.Errorf("constructed array %v not equal to itself", v1.Interface())
	}

	ToSlice(v1.Index(0)).Index(0).Set(ReflectOn("abc"))
	ToSlice(v2.Index(0)).Index(0).Set(ReflectOn("efg"))
	if i1, i2 := v1.Interface(), v2.Interface(); i1 == i2 {
		t.Errorf("constructed arrays %v and %v should not be equal", i1, i2)
	}

	ToSlice(v1.Index(0)).Index(0).Set(ReflectOn("abc"))
	ToSlice(v2.Index(0)).Index(0).Set(ReflectOn((ToSlice(v1.Index(0)).Index(0).String().Get() + " ")[:3]))
	if i1, i2 := v1.Interface(), v2.Interface(); i1 != i2 {
		t.Errorf("constructed arrays %v and %v should be equal", i1, i2)
	}

	// Test hash
	m := MakeMap(MapOf(at, ReflectOn(int(0)).Type))
	m.SetMapIndex(v1.Value, ReflectOn(1))
	if i1, i2 := v1.Interface(), v2.Interface(); !m.MapIndex(v2.Value).IsValid() {
		t.Errorf("constructed arrays %v and %v have different hashes", i1, i2)
	}
}

func TestSliceOf(t *testing.T) {
	// check construction and use of type not in binary
	type T int
	st := SliceOf(ReflectOn(T(1)).Type)
	if got, want := st.String(), "[]reflect_test.T"; got != want {
		t.Errorf("SliceOf(T(1)).String()=%q, want %q", got, want)
	}
	v := MakeSlice(st, 10, 10)
	runtime.GC()
	for i := 0; i < v.Len(); i++ {
		v.Index(i).Set(ReflectOn(T(i)))
		runtime.GC()
	}
	s := fmt.Sprint(v.Interface())
	want := "[0 1 2 3 4 5 6 7 8 9]"
	if s != want {
		t.Errorf("constructed slice = %s, want %s", s, want)
	}

	// check that type already in binary is found
	type T1 int
	checkSameType(t, Zero(SliceOf(ReflectOn(T1(1)).Type)).Interface(), []T1{})
}

func TestSliceOverflow(t *testing.T) {
	// check that MakeSlice panics when size of slice overflows uint
	const S = 1e6
	s := uint(S)
	l := (1<<(unsafe.Sizeof((*byte)(nil))*8)-1)/s + 1
	if l*s >= s {
		t.Error("slice size does not overflow")
	}
	var x [S]byte
	st := SliceOf(ReflectOn(x).Type)
	defer func() {
		err := recover()
		if err == nil {
			t.Error("slice overflow does not panic")
		}
	}()
	MakeSlice(st, int(l), int(l))
}

func TestSliceOfGC(t *testing.T) {
	type T *uintptr
	tt := ReflectOn(T(nil)).Type
	st := SliceOf(tt)
	const n = 100
	var x []interface{}
	for i := 0; i < n; i++ {
		v := MakeSlice(st, n, n)
		for j := 0; j < v.Len(); j++ {
			p := new(uintptr)
			*p = uintptr(i*n + j)
			v.Index(j).Set(Convert(ReflectOn(p), tt))
		}
		x = append(x, v.Interface())
	}
	runtime.GC()

	for i, xi := range x {
		v := ToSlice(ReflectOn(xi))
		for j := 0; j < v.Len(); j++ {
			k := v.Index(j).Deref().Interface()
			if k != uintptr(i*n+j) {
				t.Errorf("lost x[%d][%d] = %d, want %d", i, j, k, i*n+j)
			}
		}
	}
}

func TestMapOf(t *testing.T) {
	// check construction and use of type not in binary
	type K string
	type V float64

	v := MakeMap(MapOf(ReflectOn(K("")).Type, ReflectOn(V(0)).Type))
	runtime.GC()
	v.SetMapIndex(ReflectOn(K("a")), ReflectOn(V(1)))
	runtime.GC()

	s := fmt.Sprint(v.Interface())
	want := "map[a:1]"
	if s != want {
		t.Errorf("constructed map = %s, want %s", s, want)
	}

	// check that type already in binary is found
	checkSameType(t, Zero(MapOf(ReflectOn(V(0)).Type, ReflectOn(K("")).Type)).Interface(), map[V]K(nil))
}

func TestMapOfGCKeys(t *testing.T) {
	type T *uintptr
	tt := ReflectOn(T(nil)).Type
	mt := MapOf(tt, ReflectOn(false).Type)

	// NOTE: The garbage collector handles allocated maps specially,
	// so we have to save pointers to maps in x; the pointer code will
	// use the gc info in the newly constructed map type.
	const n = 100
	var x []interface{}
	for i := 0; i < n; i++ {
		v := MakeMap(mt)
		for j := 0; j < n; j++ {
			p := new(uintptr)
			*p = uintptr(i*n + j)
			v.SetMapIndex(Convert(ReflectOn(p), tt), ReflectOn(true))
		}
		pv := New(mt)
		pv.Deref().Set(v.Value)
		x = append(x, pv.Interface())
	}
	runtime.GC()

	for i, xi := range x {
		v := ToMap(ReflectOnPtr(xi))
		var out []int
		for _, kv := range v.MapKeys() {
			out = append(out, int(kv.Deref().Interface().(uintptr)))
		}
		sort.Ints(out)
		for j, k := range out {
			if k != i*n+j {
				t.Errorf("lost x[%d][%d] = %d, want %d", i, j, k, i*n+j)
			}
		}
	}
}

func TestMapOfGCValues(t *testing.T) {
	type T *uintptr
	tt := ReflectOn(T(nil)).Type
	mt := MapOf(ReflectOn(1).Type, tt)

	// NOTE: The garbage collector handles allocated maps specially,
	// so we have to save pointers to maps in x; the pointer code will
	// use the gc info in the newly constructed map type.
	const n = 100
	var x []interface{}
	for i := 0; i < n; i++ {
		v := MakeMap(mt)
		for j := 0; j < n; j++ {
			p := new(uintptr)
			*p = uintptr(i*n + j)
			v.SetMapIndex(ReflectOn(j), Convert(ReflectOn(p), tt))
		}
		pv := New(mt)
		pv.Deref().Set(v.Value)
		x = append(x, pv.Interface())
	}
	runtime.GC()

	for i, xi := range x {
		v := ToMap(ReflectOnPtr(xi))
		for j := 0; j < n; j++ {
			k := v.MapIndex(ReflectOn(j)).Deref().Interface().(uintptr)
			if k != uintptr(i*n+j) {
				t.Errorf("lost x[%d][%d] = %d, want %d", i, j, k, i*n+j)
			}
		}
	}
}

func TestTypelinksSorted(t *testing.T) {
	var last string
	for i, n := range TypeLinks() {
		if n < last {
			t.Errorf("typelinks not sorted: %q [%d] > %q [%d]", last, i-1, n, i)
		}
		last = n
	}
}

func TestAllocsInterfaceBig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	v := ReflectOn(S{})
	if allocs := testing.AllocsPerRun(100, func() { v.Interface() }); allocs > 0 {
		t.Error("allocs:", allocs)
	}
}

func TestAllocsInterfaceSmall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	v := ReflectOn(int64(0))
	if allocs := testing.AllocsPerRun(100, func() { v.Interface() }); allocs > 0 {
		t.Error("allocs:", allocs)
	}
}

func TestReflectMethodTraceback(t *testing.T) {
	p := Point{3, 4}
	m := ToStruct(ReflectOn(p)).MethodByName("GCMethod")
	cl, ok := ReflectOn(m.Interface()).Call([]Value{ReflectOn(5)})
	if ok {
		i := cl[0].Int().Get()
		if i != 8 {
			t.Errorf("Call returned %d; want 8", i)
		}
	} else {
		t.Error("Call failed")
	}
}

func TestBigZero(t *testing.T) {
	const size = 1 << 10
	var v [size]byte
	z := Zero(ReflectOn(v).Type).Interface().([size]byte)
	for i := 0; i < size; i++ {
		if z[i] != 0 {
			t.Errorf("Zero object not all zero, index %d", i)
		}
	}
}

func TestFieldByIndexNil(t *testing.T) {
	type P struct {
		F int
	}
	type T struct {
		*P
	}
	v := ToStruct(ReflectOn(T{}))

	v.FieldByName("P") // should be fine

	defer func() {
		if err := recover(); err == nil {
			// t.Errorf("no error")
		} else if !strings.Contains(fmt.Sprint(err), "nil pointer to embedded struct") {
			t.Errorf(`err=%q, wanted error containing "nil pointer to embedded struct"`, err)
		}
	}()

	v.FieldByName("F") // should panic, but we don't care anymore
}

func TestValueString(t *testing.T) {
	rv := ReflectOn(Impl{})
	if rv.String().Debug != "<"+testPackageName+".Impl Value>" {
		t.Errorf("ReflectOn(Impl{}).String() = %q, want %q", rv.String().Debug, "<"+testPackageName+".Impl Value>")
	}

	method := ToStruct(rv).Method(0)
	if method.String().Debug != "<func() Value>" {
		t.Errorf("ReflectOn(Impl{}).Method(0).String() = %q, want %q", method.String().Debug, "<func() Value>")
	}
}

func TestInvalid(t *testing.T) {
	// Used to have inconsistency between IsValid() and Kind() != Invalid.
	type T struct{ v interface{} }

	v := ToStruct(ReflectOn(T{})).Field(0)
	if v.IsValid() != true || v.Kind() != Interface {
		t.Errorf("field: IsValid=%v, Kind=%v, want true, Interface", v.IsValid(), v.Kind())
	}
	v = v.Iface()
	if v.IsValid() != false || v.Kind() != Invalid {
		t.Errorf("field elem: IsValid=%v, Kind=%v, want false, Invalid", v.IsValid(), v.Kind())
	}
}

// Issue 8917.
func TestLargeGCProg(t *testing.T) {
	fv := ReflectOn(func([256]*byte) {})
	fv.Call([]Value{ReflectOn([256]*byte{})})
}

func TestKeepMethodLive(t *testing.T) {
	// Test that we keep methodValue live as long as it is
	// referenced on the stack.
	KeepMethodLive{}.Method1(10)
}

func TestFuncLayout(t *testing.T) {
	for _, lt := range funcLayoutTests {
		typ, argsize, retOffset, stack, gc, ptrs := FuncLayout(lt.t, lt.rcvr)
		if typ.Size() != lt.size {
			t.Errorf("funcLayout(%v, %v).size=%d, want %d", lt.t, lt.rcvr, typ.Size(), lt.size)
		}
		if argsize != lt.argsize {
			t.Errorf("funcLayout(%v, %v).argsize=%d, want %d", lt.t, lt.rcvr, argsize, lt.argsize)
		}
		if retOffset != lt.retOffset {
			t.Errorf("funcLayout(%v, %v).retOffset=%d, want %d", lt.t, lt.rcvr, retOffset, lt.retOffset)
		}
		if !bytes.Equal(stack, lt.stack) {
			t.Errorf("funcLayout(%v, %v).stack=%v, want %v", lt.t, lt.rcvr, stack, lt.stack)
		}
		if !bytes.Equal(gc, lt.gc) {
			t.Errorf("funcLayout(%v, %v).gc=%v, want %v", lt.t, lt.rcvr, gc, lt.gc)
		}
		if ptrs && len(stack) == 0 || !ptrs && len(stack) > 0 {
			t.Errorf("funcLayout(%v, %v) pointers flag=%v, want %v", lt.t, lt.rcvr, ptrs, !ptrs)
		}
	}
}

func TestGCBits(t *testing.T) {
	verifyGCBits(t, ReflectOn((*byte)(nil)).Type, []byte{1})

	// Building blocks for types seen by the compiler (like [2]Xscalar).
	// The compiler will create the type structures for the derived types,
	// including their GC metadata.
	type Xscalar struct{ x uintptr }
	type Xptr struct{ x *byte }
	type Xptrscalar struct {
		*byte
		uintptr
	}
	type Xscalarptr struct {
		uintptr
		*byte
	}
	type Xbigptrscalar struct {
		_ [100]*byte
		_ [100]uintptr
	}

	var Tscalar, Tint64, Tptr, Tscalarptr, Tptrscalar, Tbigptrscalar *RType
	{
		// Building blocks for types constructed by reflect.
		// This code is in a separate block so that code below
		// cannot accidentally refer to these.
		// The compiler must NOT see types derived from these
		// (for example, [2]Scalar must NOT appear in the program),
		// or else reflect will use it instead of having to construct one.
		// The goal is to test the construction.
		type Scalar struct{ x uintptr }
		type Ptr struct{ x *byte }
		type Ptrscalar struct {
			*byte
			uintptr
		}
		type Scalarptr struct {
			uintptr
			*byte
		}
		type Bigptrscalar struct {
			_ [100]*byte
			_ [100]uintptr
		}
		type Int64 int64
		Tscalar = ReflectOn(Scalar{}).Type
		Tint64 = ReflectOn(Int64(0)).Type
		Tptr = ReflectOn(Ptr{}).Type
		Tscalarptr = ReflectOn(Scalarptr{}).Type
		Tptrscalar = ReflectOn(Ptrscalar{}).Type
		Tbigptrscalar = ReflectOn(Bigptrscalar{}).Type
	}

	empty := []byte{}

	verifyGCBits(t, ReflectOn(Xscalar{}).Type, empty)
	verifyGCBits(t, Tscalar, empty)
	verifyGCBits(t, ReflectOn(Xptr{}).Type, lit(1))
	verifyGCBits(t, Tptr, lit(1))
	verifyGCBits(t, ReflectOn(Xscalarptr{}).Type, lit(0, 1))
	verifyGCBits(t, Tscalarptr, lit(0, 1))
	verifyGCBits(t, ReflectOn(Xptrscalar{}).Type, lit(1))
	verifyGCBits(t, Tptrscalar, lit(1))

	verifyGCBits(t, ReflectOn([0]Xptr{}).Type, empty)
	verifyGCBits(t, ArrayOf(Tptr, 0), empty)
	verifyGCBits(t, ReflectOn([1]Xptrscalar{}).Type, lit(1))
	verifyGCBits(t, ArrayOf(Tptrscalar, 1), lit(1))
	verifyGCBits(t, ReflectOn([2]Xscalar{}).Type, empty)
	verifyGCBits(t, ArrayOf(Tscalar, 2), empty)
	verifyGCBits(t, ReflectOn([10000]Xscalar{}).Type, empty)
	verifyGCBits(t, ArrayOf(Tscalar, 10000), empty)
	verifyGCBits(t, ReflectOn([2]Xptr{}).Type, lit(1, 1))
	verifyGCBits(t, ArrayOf(Tptr, 2), lit(1, 1))
	verifyGCBits(t, ReflectOn([10000]Xptr{}).Type, rep(10000, lit(1)))
	verifyGCBits(t, ArrayOf(Tptr, 10000), rep(10000, lit(1)))
	verifyGCBits(t, ReflectOn([2]Xscalarptr{}).Type, lit(0, 1, 0, 1))
	verifyGCBits(t, ArrayOf(Tscalarptr, 2), lit(0, 1, 0, 1))
	verifyGCBits(t, ReflectOn([10000]Xscalarptr{}).Type, rep(10000, lit(0, 1)))
	verifyGCBits(t, ArrayOf(Tscalarptr, 10000), rep(10000, lit(0, 1)))
	verifyGCBits(t, ReflectOn([2]Xptrscalar{}).Type, lit(1, 0, 1))
	verifyGCBits(t, ArrayOf(Tptrscalar, 2), lit(1, 0, 1))
	verifyGCBits(t, ReflectOn([10000]Xptrscalar{}).Type, rep(10000, lit(1, 0)))
	verifyGCBits(t, ArrayOf(Tptrscalar, 10000), rep(10000, lit(1, 0)))
	verifyGCBits(t, ReflectOn([1][10000]Xptrscalar{}).Type, rep(10000, lit(1, 0)))
	verifyGCBits(t, ArrayOf(ArrayOf(Tptrscalar, 10000), 1), rep(10000, lit(1, 0)))
	verifyGCBits(t, ReflectOn([2][10000]Xptrscalar{}).Type, rep(2*10000, lit(1, 0)))
	verifyGCBits(t, ArrayOf(ArrayOf(Tptrscalar, 10000), 2), rep(2*10000, lit(1, 0)))
	verifyGCBits(t, ReflectOn([4]Xbigptrscalar{}).Type, join(rep(3, join(rep(100, lit(1)), rep(100, lit(0)))), rep(100, lit(1))))
	verifyGCBits(t, ArrayOf(Tbigptrscalar, 4), join(rep(3, join(rep(100, lit(1)), rep(100, lit(0)))), rep(100, lit(1))))

	verifyGCBitsSlice(t, ReflectOn([]Xptr{}).Type, 0, empty)
	verifyGCBitsSlice(t, SliceOf(Tptr), 0, empty)
	verifyGCBitsSlice(t, ReflectOn([]Xptrscalar{}).Type, 1, lit(1))
	verifyGCBitsSlice(t, SliceOf(Tptrscalar), 1, lit(1))
	verifyGCBitsSlice(t, ReflectOn([]Xscalar{}).Type, 2, lit(0))
	verifyGCBitsSlice(t, SliceOf(Tscalar), 2, lit(0))
	verifyGCBitsSlice(t, ReflectOn([]Xscalar{}).Type, 10000, lit(0))
	verifyGCBitsSlice(t, SliceOf(Tscalar), 10000, lit(0))
	verifyGCBitsSlice(t, ReflectOn([]Xptr{}).Type, 2, lit(1))
	verifyGCBitsSlice(t, SliceOf(Tptr), 2, lit(1))
	verifyGCBitsSlice(t, ReflectOn([]Xptr{}).Type, 10000, lit(1))
	verifyGCBitsSlice(t, SliceOf(Tptr), 10000, lit(1))
	verifyGCBitsSlice(t, ReflectOn([]Xscalarptr{}).Type, 2, lit(0, 1))
	verifyGCBitsSlice(t, SliceOf(Tscalarptr), 2, lit(0, 1))
	verifyGCBitsSlice(t, ReflectOn([]Xscalarptr{}).Type, 10000, lit(0, 1))
	verifyGCBitsSlice(t, SliceOf(Tscalarptr), 10000, lit(0, 1))
	verifyGCBitsSlice(t, ReflectOn([]Xptrscalar{}).Type, 2, lit(1, 0))
	verifyGCBitsSlice(t, SliceOf(Tptrscalar), 2, lit(1, 0))
	verifyGCBitsSlice(t, ReflectOn([]Xptrscalar{}).Type, 10000, lit(1, 0))
	verifyGCBitsSlice(t, SliceOf(Tptrscalar), 10000, lit(1, 0))
	verifyGCBitsSlice(t, ReflectOn([][10000]Xptrscalar{}).Type, 1, rep(10000, lit(1, 0)))
	verifyGCBitsSlice(t, SliceOf(ArrayOf(Tptrscalar, 10000)), 1, rep(10000, lit(1, 0)))
	verifyGCBitsSlice(t, ReflectOn([][10000]Xptrscalar{}).Type, 2, rep(10000, lit(1, 0)))
	verifyGCBitsSlice(t, SliceOf(ArrayOf(Tptrscalar, 10000)), 2, rep(10000, lit(1, 0)))
	verifyGCBitsSlice(t, ReflectOn([]Xbigptrscalar{}).Type, 4, join(rep(100, lit(1)), rep(100, lit(0))))
	verifyGCBitsSlice(t, SliceOf(Tbigptrscalar), 4, join(rep(100, lit(1)), rep(100, lit(0))))

	verifyGCBits(t, ReflectOn((chan [100]Xscalar)(nil)).Type, lit(1))

	verifyGCBits(t, ReflectOn((func([10000]Xscalarptr))(nil)).Type, lit(1))

	verifyGCBits(t, ReflectOn((map[[10000]Xscalarptr]Xscalar)(nil)).Type, lit(1))
	verifyGCBits(t, MapOf(ArrayOf(Tscalarptr, 10000), Tscalar), lit(1))

	verifyGCBits(t, ReflectOn((*[10000]Xscalar)(nil)).Type, lit(1))
	verifyGCBits(t, ArrayOf(Tscalar, 10000).PtrTo(), lit(1))

	verifyGCBits(t, ReflectOn(([][10000]Xscalar)(nil)).Type, lit(1))
	verifyGCBits(t, SliceOf(ArrayOf(Tscalar, 10000)), lit(1))

	hdr := make([]byte, 8/PtrSize)

	verifyMapBucket := func(t *testing.T, k, e *RType, m interface{}, want []byte) {
		verifyGCBits(t, MapBucketOf(k, e), want)
		verifyGCBits(t, CachedBucketOf(ReflectOn(m).Type), want)
	}
	verifyMapBucket(t,
		Tscalar, Tptr,
		map[Xscalar]Xptr(nil),
		join(hdr, rep(8, lit(0)), rep(8, lit(1)), lit(1)))
	verifyMapBucket(t,
		Tscalarptr, Tptr,
		map[Xscalarptr]Xptr(nil),
		join(hdr, rep(8, lit(0, 1)), rep(8, lit(1)), lit(1)))
	verifyMapBucket(t, Tint64, Tptr,
		map[int64]Xptr(nil),
		join(hdr, rep(8, rep(8/PtrSize, lit(0))), rep(8, lit(1)), naclpad(), lit(1)))
	verifyMapBucket(t,
		Tscalar, Tscalar,
		map[Xscalar]Xscalar(nil),
		empty)
	verifyMapBucket(t,
		ArrayOf(Tscalarptr, 2), ArrayOf(Tptrscalar, 3),
		map[[2]Xscalarptr][3]Xptrscalar(nil),
		join(hdr, rep(8*2, lit(0, 1)), rep(8*3, lit(1, 0)), lit(1)))
	verifyMapBucket(t,
		ArrayOf(Tscalarptr, 64/PtrSize), ArrayOf(Tptrscalar, 64/PtrSize),
		map[[64 / PtrSize]Xscalarptr][64 / PtrSize]Xptrscalar(nil),
		join(hdr, rep(8*64/PtrSize, lit(0, 1)), rep(8*64/PtrSize, lit(1, 0)), lit(1)))
	verifyMapBucket(t,
		ArrayOf(Tscalarptr, 64/PtrSize+1), ArrayOf(Tptrscalar, 64/PtrSize),
		map[[64/PtrSize + 1]Xscalarptr][64 / PtrSize]Xptrscalar(nil),
		join(hdr, rep(8, lit(1)), rep(8*64/PtrSize, lit(1, 0)), lit(1)))
	verifyMapBucket(t,
		ArrayOf(Tscalarptr, 64/PtrSize), ArrayOf(Tptrscalar, 64/PtrSize+1),
		map[[64 / PtrSize]Xscalarptr][64/PtrSize + 1]Xptrscalar(nil),
		join(hdr, rep(8*64/PtrSize, lit(0, 1)), rep(8, lit(1)), lit(1)))
	verifyMapBucket(t,
		ArrayOf(Tscalarptr, 64/PtrSize+1), ArrayOf(Tptrscalar, 64/PtrSize+1),
		map[[64/PtrSize + 1]Xscalarptr][64/PtrSize + 1]Xptrscalar(nil),
		join(hdr, rep(8, lit(1)), rep(8, lit(1)), lit(1)))
}

func TestTypeOfValueOf(t *testing.T) {
	// Check that all the type constructors return concrete *RType implementations.
	// It's difficult to test directly because the reflect package is only at arm's length.
	// The easiest thing to do is just call a function that crashes if it doesn't get an *RType.
	check := func(name string, typ *RType) {
		if underlying := ReflectOn(typ).Type.String(); underlying != "*reflect.RType" {
			t.Errorf("%v returned %v, not *reflect.RType", name, underlying)
		}
	}

	type T struct{ int }
	check("TypeOf", ReflectOn(T{}).Type)
	check("ArrayOf", ArrayOf(ReflectOn(T{}).Type, 10))
	check("MapOf", MapOf(ReflectOn(T{}).Type, ReflectOn(T{}).Type))
	check("PtrTo", ReflectOn(T{}).Type.PtrTo())
	check("SliceOf", SliceOf(ReflectOn(T{}).Type))
}

func TestPtrToMethods(t *testing.T) {
	var y struct{ XM }
	yp := New(ReflectOn(y).Type).Interface()
	_, ok := yp.(fmt.Stringer)
	if !ok {
		t.Error("does not implement Stringer, but should")
	}
}

func TestMapAlloc(t *testing.T) {
	m := ToMap(ReflectOn(make(map[int]int, 10)))
	k := ReflectOn(5)
	v := ReflectOn(7)
	allocs := testing.AllocsPerRun(100, func() {
		m.SetMapIndex(k, v)
	})
	if allocs > 0.5 {
		t.Errorf("allocs per map assignment: want 0 got %f", allocs)
	}

	const size = 1000
	tmp := 0
	val := ReflectOnPtr(&tmp)
	allocs = testing.AllocsPerRun(100, func() {
		mv := MakeMapWithSize(ReflectOn(map[int]int{}).Type, size)
		// Only adding half of the capacity to not trigger re-allocations due too many overloaded buckets.
		for i := 0; i < size/2; i++ {
			val.Int().Set(int64(i))
			mv.SetMapIndex(val, val)
		}
	})
	if allocs > 10 {
		t.Errorf("allocs per map assignment: want at most 10 got %f", allocs)
	}
	// Empirical testing shows that with capacity hint single run will trigger 3 allocations and without 91. I set
	// the threshold to 10, to not make it overly brittle if something changes in the initial allocation of the
	// map, but to still catch a regression where we keep re-allocating in the hashmap as new entries are added.
}

func TestNames(t *testing.T) {
	for _, test := range nameTests {
		typ := ReflectOn(test.v).Type.Deref()
		if got := typ.Name(); got != test.want {
			t.Errorf("%v Name()=%q, want %q", typ, got, test.want)
		}
	}
}

func TestExported(t *testing.T) {
	type ΦExported struct{}
	type φUnexported struct{}
	type BigP *big
	type P int
	type p *P
	type P2 p
	type p3 p

	type exportTest struct {
		v    interface{}
		want bool
	}
	exportTests := []exportTest{
		{D1{}, true},
		{(*D1)(nil), true},
		{big{}, false},
		{(*big)(nil), false},
		{(BigP)(nil), true},
		{(*BigP)(nil), true},
		{ΦExported{}, true},
		{φUnexported{}, false},
		{P(0), true},
		{(p)(nil), false},
		{(P2)(nil), true},
		{(p3)(nil), false},
	}

	for i, test := range exportTests {
		typ := ReflectOn(test.v).Type
		if got := IsExported(typ); got != test.want {
			t.Errorf("%d: %s exported=%v, want %v", i, typ.Name(), got, test.want)
		}
	}
}

func TestNameBytesAreAligned(t *testing.T) {
	typ := ReflectOn(embed{}).Type
	b := FirstMethodNameBytes(typ)
	v := uintptr(unsafe.Pointer(b))
	if v%unsafe.Alignof((*byte)(nil)) != 0 {
		t.Errorf("reflect.name.bytes pointer is not aligned: %x", v)
	}
}

func TestTypeStrings(t *testing.T) {
	type stringTest struct {
		typ  *RType
		want string
	}
	stringTests := []stringTest{
		{ReflectOn(func(int) {}).Type, "func(int)"},
		{ReflectOn(XM{}).Type, "" + testPackageName + ".XM"},
		{ReflectOn(new(XM)).Type, "*" + testPackageName + ".XM"},
		{ReflectOn(new(XM).String).Type, "func() string"},
		{MapOf(ReflectOn(int(0)).Type, ReflectOn(XM{}).Type), "map[int]" + testPackageName + ".XM"},
		{ArrayOf(ReflectOn(XM{}).Type, 3), "[3]" + testPackageName + ".XM"},
		{ArrayOf(ReflectOn(struct{}{}).Type, 3), "[3]struct {}"},
	}

	for i, test := range stringTests {
		if got, want := test.typ.String(), test.want; got != want {
			t.Errorf("type [%d] String()=%q, want %q", i, got, want)
		}
	}
}

func TestOffsetLock(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		i := i
		wg.Add(1)
		go func() {
			for j := 0; j < 50; j++ {
				ResolveReflectName(fmt.Sprintf("OffsetLockName:%d:%d", i, j))
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSwapper(t *testing.T) {
	type I int
	var a, b, c I
	type pair struct {
		x, y int
	}
	type pairPtr struct {
		x, y int
		p    *I
	}
	type S string

	tests := []struct {
		in   interface{}
		i, j int
		want interface{}
	}{
		{
			in:   []int{1, 20, 300},
			i:    0,
			j:    2,
			want: []int{300, 20, 1},
		},
		{
			in:   []uintptr{1, 20, 300},
			i:    0,
			j:    2,
			want: []uintptr{300, 20, 1},
		},
		{
			in:   []int16{1, 20, 300},
			i:    0,
			j:    2,
			want: []int16{300, 20, 1},
		},
		{
			in:   []int8{1, 20, 100},
			i:    0,
			j:    2,
			want: []int8{100, 20, 1},
		},
		{
			in:   []*I{&a, &b, &c},
			i:    0,
			j:    2,
			want: []*I{&c, &b, &a},
		},
		{
			in:   []string{"eric", "sergey", "larry"},
			i:    0,
			j:    2,
			want: []string{"larry", "sergey", "eric"},
		},
		{
			in:   []S{"eric", "sergey", "larry"},
			i:    0,
			j:    2,
			want: []S{"larry", "sergey", "eric"},
		},
		{
			in:   []pair{{1, 2}, {3, 4}, {5, 6}},
			i:    0,
			j:    2,
			want: []pair{{5, 6}, {3, 4}, {1, 2}},
		},
		{
			in:   []pairPtr{{1, 2, &a}, {3, 4, &b}, {5, 6, &c}},
			i:    0,
			j:    2,
			want: []pairPtr{{5, 6, &c}, {3, 4, &b}, {1, 2, &a}},
		},
	}

	for i, tt := range tests {
		inStr := fmt.Sprint(tt.in)
		Swapper(tt.in)(tt.i, tt.j)
		if !DeepEqual(tt.in, tt.want) {
			t.Errorf("%d. swapping %v and %v of %v = %v; want %v", i, tt.i, tt.j, inStr, tt.in, tt.want)
		}
	}
}

// TestUnaddressableField tests that the reflect package will not allow
// a type from another package to be used as a named type with an
// unexported field.
//
// This ensures that unexported fields cannot be modified by other packages.
func TestUnaddressableField(t *testing.T) {
	var b Buffer // type defined in reflect, a different package
	var localBuffer struct {
		buf []byte
	}
	lv := ReflectOnPtr(&localBuffer)
	rv := ReflectOn(b)

	shouldPanic := func(f func()) {
		defer func() {
			if recover() == nil {
				// yeap, we don't panic anymore
			}
		}()
		f()
	}
	shouldPanic(func() {
		lv.Set(rv)
	})
}

func TestAliasNames(t *testing.T) {
	t1 := Talias1{byte: 1, uint8: 2, int: 3, int32: 4, rune: 5}
	out := fmt.Sprintf("%#v", t1)
	want := "" + testPackageName + ".Talias1{byte:0x1, uint8:0x2, int:3, int32:4, rune:5}"
	if out != want {
		t.Errorf("Talias1 print:\nhave: %s\nwant: %s", out, want)
	}

	t2 := Talias2{Tint: 1, Tint2: 2}
	out = fmt.Sprintf("%#v", t2)
	want = "" + testPackageName + ".Talias2{Tint:1, Tint2:2}"
	if out != want {
		t.Errorf("Talias2 print:\nhave: %s\nwant: %s", out, want)
	}
}

func TestIssue22031(t *testing.T) {
	type s []struct{ C int }

	type t1 struct{ s }
	type t2 struct{ f s }
	r1 := ReflectOn(t1{s{{}}})
	r2 := ReflectOn(t2{s{{}}})
	f1 := ToStruct(r1).Field(0)
	f2 := ToStruct(r2).Field(0)
	e1 := ToSlice(f1).Index(0)
	e2 := ToSlice(f2).Index(0)
	tests := []Value{
		ToStruct(e1).Field(0),
		ToStruct(e2).Field(0),
	}

	for i, test := range tests {
		if test.CanSet() {
			t.Errorf("%d: CanSet: got true, want false", i)
		}
	}
}

func TestIssue22073(t *testing.T) {
	m := ToStruct(ReflectOn(NonExportedFirst(0))).Method(0)
	if got := m.NumOut(); got != 0 {
		t.Errorf("NumOut: got %v, want 0", got)
	}
	// Shouldn't panic.
	m.Call([]Value{ReflectOn("Creveți")})
}

func TestImplicitMapConversion(t *testing.T) {
	// Test implicit conversions in MapIndex and SetMapIndex.
	{
		// direct
		m := make(map[int]int)
		mv := ToMap(ReflectOn(m))
		mv.SetMapIndex(ReflectOn(1), ReflectOn(2))
		x, ok := m[1]
		if x != 2 {
			t.Errorf("#1 after SetMapIndex(1,2): %d, %t (map=%v)", x, ok, m)
		}
		if n := mv.MapIndex(ReflectOn(1)).Interface().(int); n != 2 {
			t.Errorf("#1 MapIndex(1) = %d", n)
		}
	}
	{
		// convert interface key
		m := make(map[interface{}]int)
		mv := ToMap(ReflectOn(m))
		mv.SetMapIndex(ReflectOn(1), ReflectOn(2))
		x, ok := m[1]
		if x != 2 {
			t.Errorf("#2 after SetMapIndex(1,2): %d, %t (map=%v)", x, ok, m)
		}
		if n := mv.MapIndex(ReflectOn(1)).Interface().(int); n != 2 {
			t.Errorf("#2 MapIndex(1) = %d", n)
		}
	}
	{
		// convert interface value
		m := make(map[int]interface{})
		mv := ToMap(ReflectOn(m))
		mv.SetMapIndex(ReflectOn(1), ReflectOn(2))
		x, ok := m[1]
		if x != 2 {
			t.Errorf("#3 after SetMapIndex(1,2): %d, %t (map=%v)", x, ok, m)
		}
		if n := mv.MapIndex(ReflectOn(1)).Interface().(int); n != 2 {
			t.Errorf("#3 MapIndex(1) = %d", n)
		}
	}
	{
		// convert both interface key and interface value
		m := make(map[interface{}]interface{})
		mv := ToMap(ReflectOn(m))
		mv.SetMapIndex(ReflectOn(1), ReflectOn(2))
		x, ok := m[1]
		if x != 2 {
			t.Errorf("#4 after SetMapIndex(1,2): %d, %t (map=%v)", x, ok, m)
		}
		if n := mv.MapIndex(ReflectOn(1)).Interface().(int); n != 2 {
			t.Errorf("#4 MapIndex(1) = %d", n)
		}
	}
	{
		// convert both, with non-empty interfaces
		m := make(map[io.Reader]io.Writer)
		mv := ToMap(ReflectOn(m))
		b1 := new(bytes.Buffer)
		b2 := new(bytes.Buffer)
		mv.SetMapIndex(ReflectOn(b1), ReflectOn(b2))
		x, ok := m[b1]
		if x != b2 {
			t.Errorf("#5 after SetMapIndex(b1, b2): %p (!= %p), %t (map=%v)", x, b2, ok, m)
		}
		if p := mv.MapIndex(ReflectOn(b1)).Iface().UnsafePointer().Get(); p != unsafe.Pointer(b2) {
			t.Errorf("#5 MapIndex(b1) = %#x want %p", p, b2)
		}
	}
	{
		// convert identical underlying types
		// TODO(rsc): Should be able to define MyBuffer here.
		// 6l prints very strange messages about .this.Bytes etc
		// when we do that though, so MyBuffer is defined
		// at top level.
		m := make(map[*MyBuffer]*bytes.Buffer)
		mv := ToMap(ReflectOn(m))
		b1 := new(MyBuffer)
		b2 := new(bytes.Buffer)
		mv.SetMapIndex(ReflectOn(b1), ReflectOn(b2))
		x, ok := m[b1]
		if x != b2 {
			t.Errorf("#7 after SetMapIndex(b1, b2): %p (!= %p), %t (map=%v)", x, b2, ok, m)
		}

		if p := mv.MapIndex(ReflectOn(b1)).UnsafePointer().Get(); p != unsafe.Pointer(b2) {
			t.Errorf("#7 MapIndex(b1) = %#x want %p", p, b2)
		}
	}

}

func TestImplicitSetConversion(t *testing.T) {
	// Assume TestImplicitMapConversion covered the basics.
	// Just make sure conversions are being applied at all.
	var r io.Reader
	b := new(bytes.Buffer)
	rv := ReflectOnPtr(&r)
	rv.Set(ReflectOn(b))
	if r != b {
		t.Errorf("after Set: r=%T(%v)", r, r)
	}
}

func TestImplicitCallConversion(t *testing.T) {
	// Arguments must be assignable to parameter types.
	fv := ReflectOn(io.WriteString)
	b := new(bytes.Buffer)
	fv.Call([]Value{ReflectOn(b), ReflectOn("hello world")})
	if b.String() != "hello world" {
		t.Errorf("After call: string=%q want %q", b.String(), "hello world")
	}
}

func TestImplicitAppendConversion(t *testing.T) {
	// Arguments must be assignable to the slice's element type.
	s := []io.Reader{}
	sv := ToSlice(ReflectOnPtr(&s))
	b := new(bytes.Buffer)
	sv.Set(sv.Append(ReflectOn(b)).Value)
	if len(s) != 1 || s[0] != b {
		t.Errorf("after append: s=%v want [%p]", s, b)
	}
}

func TestImplements(t *testing.T) {
	for _, tt := range implementsTests {
		xv := ReflectOn(tt.x).Type.Deref()
		xt := ReflectOn(tt.t).Type.Deref()
		if b := xv.Implements(xt); b != tt.b {
			t.Errorf("(%s).Implements(%s) = %v, want %v", xv.String(), xt.String(), b, tt.b)
		}
	}
}

func TestAssignableTo(t *testing.T) {
	for _, tt := range append(assignableTests, implementsTests...) {
		xv := ReflectOn(tt.x).Type.Deref()
		xt := ReflectOn(tt.t).Type.Deref()
		if b := xv.AssignableTo(xt); b != tt.b {
			t.Errorf("(%s).AssignableTo(%s) = %v, want %v", xv.String(), xt.String(), b, tt.b)
		}
	}
}

func init() {
	loop1 = &loop2
	loop2 = &loop1

	loopy1 = &loopy2
	loopy2 = &loopy1
	var argAlign uintptr = PtrSize
	if runtime.GOARCH == "amd64p32" {
		argAlign = 2 * PtrSize
	}
	roundup := func(x uintptr, a uintptr) uintptr {
		return (x + a - 1) / a * a
	}

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func(a, b string) string { return "" }).Type,
			6 * PtrSize,
			4 * PtrSize,
			4 * PtrSize,
			[]byte{1, 0, 1},
			[]byte{1, 0, 1, 0, 1},
		})

	var r []byte
	if PtrSize == 4 {
		r = []byte{0, 0, 0, 1}
	} else {
		r = []byte{0, 0, 1}
	}
	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func(a, b, c uint32, p *byte, d uint16) {}).Type,
			roundup(roundup(3*4, PtrSize)+PtrSize+2, argAlign),
			roundup(3*4, PtrSize) + PtrSize + 2,
			roundup(roundup(3*4, PtrSize)+PtrSize+2, argAlign),
			r,
			r,
		})

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func(a map[int]int, b uintptr, c interface{}) {}).Type,
			4 * PtrSize,
			4 * PtrSize,
			4 * PtrSize,
			[]byte{1, 0, 1, 1},
			[]byte{1, 0, 1, 1},
		})

	type S struct {
		a, b uintptr
		c, d *byte
	}
	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func(a S) {}).Type,
			4 * PtrSize,
			4 * PtrSize,
			4 * PtrSize,
			[]byte{0, 0, 1, 1},
			[]byte{0, 0, 1, 1},
		})

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			ReflectOn((*byte)(nil)).Type,
			ReflectOn(func(a uintptr, b *int) {}).Type,
			roundup(3*PtrSize, argAlign),
			3 * PtrSize,
			roundup(3*PtrSize, argAlign),
			[]byte{1, 0, 1},
			[]byte{1, 0, 1},
		})

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func(a uintptr) {}).Type,
			roundup(PtrSize, argAlign),
			PtrSize,
			roundup(PtrSize, argAlign),
			[]byte{},
			[]byte{},
		})

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			nil,
			ReflectOn(func() uintptr { return 0 }).Type,
			PtrSize,
			0,
			0,
			[]byte{},
			[]byte{},
		})

	funcLayoutTests = append(funcLayoutTests,
		funcLayoutTest{
			ReflectOn(uintptr(0)).Type,
			ReflectOn(func(a uintptr) {}).Type,
			2 * PtrSize,
			2 * PtrSize,
			2 * PtrSize,
			[]byte{1},
			[]byte{1},
			// Note: this one is tricky, as the receiver is not a pointer. But we
			// pass the receiver by reference to the autogenerated pointer-receiver
			// version of the function.
		})
}

func TestNewVsConstr(t *testing.T) {
	for _, test := range convertTests {
		typ := test.out.Type
		constr := Constr(typ)
		out := New(typ).Deref()
		if !DeepEqual(constr.Type, out.Type) {
			t.Errorf("Error : types are not the same")
		}
	}

	oneStringType := ReflectOn("SomeString").Type
	constr := Constr(oneStringType)
	out := New(oneStringType).Deref().String()
	out.Set("Deref")
	constr.String().Set("Direct")
	// t.Logf("Test set string %q %q", constr.String().Get(), out.Get())
}

func TestMakeFunc(t *testing.T) {
	f := dummy
	fv := MakeFunc(TypeOf(f), func(in []Value) []Value { return in })
	ReflectOn(&f).Deref().Set(fv)

	// Call g with small arguments so that there is
	// something predictable (and different from the
	// correct results) in those positions on the stack.
	g := dummy
	g(1, 2, 3, two{4, 5}, 6, 7, 8)

	// Call constructed function f.
	i, j, k, l, m, n, o := f(10, 20, 30, two{40, 50}, 60, 70, 80)
	if i != 10 || j != 20 || k != 30 || l != (two{40, 50}) || m != 60 || n != 70 || o != 80 {
		t.Errorf("Call returned %d, %d, %d, %v, %d, %g, %d; want 10, 20, 30, [40, 50], 60, 70, 80", i, j, k, l, m, n, o)
	}
}

func TestMakeFuncInterface(t *testing.T) {
	fn := func(i int) int {
		return i
	}
	incr := func(in []Value) []Value {
		//t.Logf("Called with %#v", in)
		return []Value{ReflectOn(int(in[0].Int().Get() + 1))}
	}
	fv := MakeFunc(TypeOf(fn), incr)
	ReflectOn(&fn).Deref().Set(fv)
	if r := fn(2); r != 3 {
		t.Errorf("Call returned %d, want 3", r)
	}
	result, ok := fv.Call([]Value{ReflectOn(14)})
	if !ok {
		t.Fatalf("Call failed.")
	}
	if r := result[0].Int().Get(); r != 15 {
		t.Errorf("Call returned %d, want 15", r)
	}
	if r := fv.Interface().(func(int) int)(26); r != 27 {
		t.Errorf("Call returned %d, want 27", r)
	}
}
