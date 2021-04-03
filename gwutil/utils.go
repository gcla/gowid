// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package gwutil provides general-purpose utilities that are not used by
// the core of gowid but that have proved useful for several pre-canned
// widgets.
package gwutil

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"sort"

	log "github.com/sirupsen/logrus"
)

//======================================================================

// Min returns the smaller of >1 integer arguments.
func Min(i int, js ...int) int {
	res := i
	for _, j := range js {
		if j < i {
			res = j
		}
	}
	return res
}

// Min returns the larger of >1 integer arguments.
func Max(i int, js ...int) int {
	res := i
	for _, j := range js {
		if j > i {
			res = j
		}
	}
	return res
}

// LimitTo is a one-liner that uses Min and Max to bound a value. Assumes
// a <= b.
func LimitTo(a, v, b int) int {
	if v < a {
		return a
	}
	if v > b {
		return b
	}
	return v
}

// StringOfLength returns a string consisting of n runes.
func StringOfLength(r rune, n int) string {
	res := make([]rune, n)
	for i := 0; i < n; i++ {
		res[i] = r
	}
	return string(res)
}

// Map is the traditional functional map function for strings.
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// IPow returns a raised to the bth power.
func IPow(a, b int) int {
	var result int = 1

	for 0 != b {
		if 0 != (b & 1) {
			result *= a
		}
		b >>= 1
		a *= a
	}

	return result
}

// Sum is a variadic function that returns the sum of its integer arguments.
func Sum(input ...int) int {
	sum := 0
	for i := range input {
		sum += input[i]
	}
	return sum
}

//======================================================================

type fract struct {
	fp  float64
	idx int
}

type fractlist []fract

func (slice fractlist) Len() int {
	return len(slice)
}

// Note > to skip the reverse
func (slice fractlist) Less(i, j int) bool {
	return slice[i].fp > slice[j].fp
}

func (slice fractlist) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// HamiltonAllocation implements the Hamilton Method (Largest remainder method) to calculate
// integral ratios. (Like it is used in some elections.)
//
// This is shamelessly cribbed from https://excess.org/svn/urwid/contrib/trunk/rbreu_scrollbar.py
//
// counts -- list of integers ('votes per party')
// alloc -- total amount to be allocated ('total amount of seats')
//
func HamiltonAllocation(counts []int, alloc int) []int {

	totalCounts := Sum(counts...)

	if totalCounts == 0 {
		return counts
	}

	res := make([]int, len(counts))
	quotas := make([]float64, len(counts))
	fracts := fractlist(make([]fract, len(counts)))

	for i, c := range counts {
		quotas[i] = (float64(c) * float64(alloc)) / float64(totalCounts)
	}

	for i, fp := range quotas {
		_, f := math.Modf(fp)
		fracts[i] = fract{fp: f, idx: i}
	}

	sort.Sort(fracts)

	for i, fp := range quotas {
		n, _ := math.Modf(fp)
		res[i] = int(n)
	}

	remainder := alloc - Sum(res...)

	for i := 0; i < remainder; i++ {
		res[fracts[i].idx] += 1
	}

	return res
}

//======================================================================

// LStripByte returns a slice of its first argument which contains all
// bytes up to but not including its second argument.
func LStripByte(data []byte, s byte) []byte {
	var i int
	for i = 0; i < len(data); i++ {
		if data[i] != s {
			break
		}
	}
	return data[i:]
}

//======================================================================

type IOption interface {
	IsNone() bool
	Value() interface{}
}

// For fmt.Stringer
func OptionString(opt IOption) string {
	if opt.IsNone() {
		return "None"
	} else {
		return fmt.Sprintf("%v", opt.Value())
	}
}

//======================================================================

// IntOption is intended to represent an Option[int]
type IntOption struct {
	some bool
	val  int
}

var _ fmt.Stringer = IntOption{}
var _ IOption = IntOption{}

func SomeInt(x int) IntOption {
	return IntOption{true, x}
}

func NoneInt() IntOption {
	return IntOption{}
}

func (i IntOption) IsNone() bool {
	return !i.some
}

func (i IntOption) Value() interface{} {
	return i.Val()
}

func (i IntOption) Val() int {
	if i.IsNone() {
		panic(errors.New("Called Val on empty IntOption"))
	}
	return i.val
}

// For fmt.Stringer
func (i IntOption) String() string {
	return OptionString(i)
}

//======================================================================

// Int64Option is intended to represent an Option[int]
type Int64Option struct {
	some bool
	val  int64
}

var _ fmt.Stringer = Int64Option{}
var _ IOption = Int64Option{}

func SomeInt64(x int64) Int64Option {
	return Int64Option{true, x}
}

func NoneInt64() Int64Option {
	return Int64Option{}
}

func (i Int64Option) IsNone() bool {
	return !i.some
}

func (i Int64Option) Value() interface{} {
	return i.Val()
}

func (i Int64Option) Val() int64 {
	if i.IsNone() {
		panic(errors.New("Called Val on empty Int64Option"))
	}
	return i.val
}

// For fmt.Stringer
func (i Int64Option) String() string {
	return OptionString(i)
}

//======================================================================

// RuneOption is intended to represent an Option[rune]
type RuneOption struct {
	some bool
	val  rune
}

var _ fmt.Stringer = RuneOption{}
var _ IOption = RuneOption{}

func SomeRune(x rune) RuneOption {
	return RuneOption{true, x}
}

func NoneRune() RuneOption {
	return RuneOption{}
}

func (i RuneOption) IsNone() bool {
	return !i.some
}

func (i RuneOption) Value() interface{} {
	return i.Val()
}

func (i RuneOption) Val() rune {
	if i.IsNone() {
		panic(errors.New("Called Val on empty ByteOption"))
	}
	return i.val
}

func (i RuneOption) String() string {
	return OptionString(i)
}

//======================================================================

const float64EqualityThreshold = 1e-5

// AlmostEqual returns true if its two arguments are within 1e-5 of each other.
func AlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// Round returns a float64 representing the closest whole number
// to the supplied float64 argument.
func Round(f float64) float64 {
	if f < 0 {
		return math.Ceil(f - 0.5)
	} else {
		return math.Floor(f + 0.5)
	}
}

// RoundFloatToInt returns an int representing the closest int to the
// supplied float, rounding up or down.
func RoundFloatToInt(val float32) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

//======================================================================

// If is a convenience function for mimicking a ternary operator e.g. If(x<y, x, y).(int)
func If(statement bool, a, b interface{}) interface{} {
	if statement {
		return a
	}
	return b
}

//======================================================================

// StartProfilingCPU is a function I used when debugging and optimizing gowid. It starts
// the Go-profiler with output going to the specified file.
func StartProfilingCPU(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	if err = pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
}

// StopProfilingCPU will stop the CPU profiler.
func StopProfilingCPU() {
	pprof.StopCPUProfile()
}

//======================================================================

// ProfileHeap is a function I used when debugging and optimizing gowid. It
// writes a Go-heap-profile to the filename specified.
func ProfileHeap(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err = pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
