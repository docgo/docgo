package main

import (
	"example.com/demomod/pkg1"
	"fmt"
)

type AType int
type BType interface {
	CInterface
	SomeFunction(func() int) int
}
type CInterface interface{}

func main() {
	var q BType
	var load pkg1.Dumper
	load = pkg1.IntDumper{}
	load.Load(1, 5, 10)

	_ = q
	fmt.Println(load.Dump())
}
