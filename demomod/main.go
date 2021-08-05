// This is a comment/doc for the main package
// This package features some interesting syntax
// and type demos for testing out the project.
package main

import (
	"example.com/demomod/pkg1"
	"fmt"
	"strings"
)

// Force multiplier
const MAIN_IMAGINARY_NO = 3.0i + 2

var MainNilVariable interface{}

// Rune is a int32 used for characters
var SomeCh rune = 1 << 30

// Byte is a uint8 used for blobs
var SomeByte byte = 255

type (
	MainT1 = string
	MainT2 = map[string]interface{}
	MainT3 = []MainT1
)

// Coordinates of the main camera
var MainX, MainY, MainZ float64
var MainA, MainB int32 = 1, 2

type MainArray [3]int
var MainArrayExample = MainArray{1, 2, 3}

func MainGetSlice(x []string) (y []string) {
	defer func() { y = append(y, "finalString") }()
	y = make([]string, 2)
	y[0] = strings.ToUpper(x[0])
	return
}

func main() {
	sl := MainGetSlice([]string{"someExample", "andAnother"})
	fmt.Println(sl)
	fmt.Println("XYZ", MainX, MainY, MainZ)

	var load pkg1.Dumper
	load = pkg1.IntDumper{}
	load.Load(1, 5, 10)

	fmt.Println(load.Dump())
}
