package main

import (
	"example.com/demomod/pkg1"
	"fmt"
)

func main() {
	var load pkg1.Dumper
	load = pkg1.IntDumper{}
	load.Load(1, 5, 10)

	fmt.Println(load.Dump())
}
