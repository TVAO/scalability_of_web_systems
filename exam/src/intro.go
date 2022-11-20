package main

import (
	"fmt"
	"runtime"
)

func main() {
	mapExample()
}

// Encoder/decoder: serialize and deserialize data as a stream
// Marshal/unmarshal: (de)serialize data as strings

// Interface
type geometry interface {
	area() float64
}

// Rectangle struct type
type rect struct {
	width, height float64
}

// Implement interface on rect by implementing methods
func (r rect) area() float64 {
	return r.width * r.height
}

// Variable interface type -> Generic measure function
func measure(g geometry) {
	fmt.Println(g)
	fmt.Println(g.area())
}

// Rect struct type implement geometry interface
// So we can use it as argument to generic measure function expecting geometry interface type
func interfaceExample() {
	r := rect{width: 3, height: 4}
	measure(r)
}

// Struct : typed collection of fields used to group data to form records
type person struct {
	name string
	age  int
}

func structExample() {
	// Create struct
	fmt.Println(person{"Bob", 20})

	// Name fields upon creation
	fmt.Println(person{name: "Alice", age: 30})

	// &prefix yields pointer to struct
	fmt.Println(&person{name: "Ann", age: 40})

	// Access fields with dot
	s := person{name: "Sean", age: 50}
	fmt.Println(s.name)

	// Struct pointers are auto dereferences
	sp := &s
	fmt.Println(sp.age)

	// Structs are MUTABLE
	sp.age = 51
	fmt.Println(sp.age)
}

// Concurrency ((u)buffered channel and goroutine)

// Pointer
// &i syntax gives memory address of i (pointer to i)
// zeroval does not change i in main but, zerooptr does
// since it has a reference to memory address of variable
func zeroval(ival int) {
	ival = 0
}
func zeroptr(iptr *int) {
	*iptr = 0
}
func pointerExample() {
	i := 1
	fmt.Println("initial:", i) // 1

	zeroval(i)
	fmt.Println("zeroval:", i) // 1

	zeroptr(&i)
	fmt.Println("zeroptr:", i) // 0 -> due to p reference

	fmt.Println("pointer:", &i) // Memory address
}

// Slice : dynamically sized array
// Formed by two indices (low and high bound): a[low : high]
func sliceExample() {

	// Make slice
	s := make([]string, 3)
	fmt.Println("empty:", s)

	// Insert value
	s[0] = "a"
	fmt.Println("Slice:", s)
	fmt.Println("get", s[2])

	// Get length
	fmt.Println("len:", len(s))

	// Copy slice
	c := make([]string, len(s))
	copy(c, s) // Copy into c from s
	fmt.Println("copy:", c)

	// "Slice" operator [low:high]
	subSlice := s[2:5]
	fmt.Println("subSlice:", subSlice)

	// Declare and initialize single line
	t := []string{"g", "h"}
	fmt.Println("dcl:", t)

	// Multi-dimensional slice (with varying inner slice lengths)
	twoD := make([][]int, 3)
	for i := 0; i < 3; i++ {
		innerLen := i + 1
		twoD[i] = make([]int, innerLen)
		for j := 0; j < innerLen; j++ {
			twoD[i][j] = i + j
		}
	}
	fmt.Println("2d: ", twoD)
}

// Map
func mapExample() {
	// Create map
	m := make(map[string]int) // String key to int value

	// Insert value
	m["k1"] = 7
	m["k2"] = 13
	fmt.Println("map:", m)

	// Retrieve value
	v1 := m["k1"]
	fmt.Println("v1: ", v1)

	// Length of map
	fmt.Println("len:", len(m))

	// Delete key value
	delete(m, "k2")
	fmt.Println("map:", m)

	// Check if value exists
	_, present := m["k2"]
	fmt.Println("present:", present)

	// Easy create map
	n := map[string]int{"foo": 1, "bar": 2}
	fmt.Println("map:", n)
}

func forLoop() {
	// For loop with condition
	for i := 0; i <= 3; i++ {
		fmt.Println(i)
	}

	// For loop (while until break)
	for {
		fmt.Println("loop")
		break // Continue supported too
	}
}

// Function with any argument number
func variadicFunc() {
	sum(1, 2)
	sum(1, 2, 3)
	nums := []int{1, 2, 3, 4}
	sum(nums...)
}
func sum(nums ...int) {
	fmt.Print(nums, " ")
	total := 0
	for _, num := range nums {
		total += num
	}
	fmt.Println(total)
}

// Anonymous function (closure) -> useful to define inline function

/*
Go functions may be closures.
A closure is a function value that references variables from outside its body.
The function may access and assign to the referenced variables;
in this sense the function is "bound" to the variables.
For example, the adder function returns a closure. Each closure is bound to its own sum variable.
*/
func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}
func closure() {
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

// Recursive function
func fact(n int) int {
	if n == 0 {
		return 1
	}
	return n * fact(n-1)
}

// fibonacci returns a function that returns
// successive fibonacci numbers from each
// successive call
func fibonacci() func() int {
	first, second := 0, 1
	return func() int {
		ret := first
		first, second = second, first+second
		return ret
	}
}

// Switch
func switchExample() {
	fmt.Print("Go runs on ")
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("OS X.")
	case "linux":
		fmt.Println("Linux.")
	default:
		// freebsd, openbsd,
		// plan9, windows...
		fmt.Printf("%s.", os)
	}
}

// Range
func rangeExample() {
	nums := []int{2, 3, 4}
	for i, num := range nums {
		if num == 3 {
			fmt.Println("index:", i)
		}
	}
}
