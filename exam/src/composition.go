package main

import (
	"fmt"
)

// Link: https://www.goinggo.net/2015/09/composition-with-go.html
// Interface provide flexibility and abstraction
// Go: no need to declare explicitly but simply implement methods in type
type vehicle interface {
	brand() string
	drive()
}

// Car struct type
type car struct {
	name     string
	maxSpeed int
}

// Car implements vehicle interface methods
func (c car) drive() {
	fmt.Println("Car drives with speed: ", c.maxSpeed)
}
func (c car) brand() string {
	return c.name
}

// Generic interface method
func printVehicleBrand(v vehicle) {
	fmt.Println(v.brand())
}

func interfaceExample() {
	bmw := car{name: "BMW", maxSpeed: 200}
	printVehicleBrand(bmw)
	bmw.drive()
}

func main() {
	interfaceExample()
}

// GO COMPOSITION (STRUCT TYPES OF TYPES)

// User struct type
type User struct {
	Name  string
	Email string
}

// GO: COMPOSITION (NOT INHERITANCE) -> no relationship between User and Admin type
// Embedded Types (Struct types with anonymous or embedded fields)
// Embed type into struct -> name of type acts as field name for embedded field
// Embed existing struct type User within struct type Admin
type Admin struct {
	User
	Level string
}
