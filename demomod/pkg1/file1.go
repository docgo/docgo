package pkg1

import "fmt"

type MyType struct {
	num float64
}

// Does something
func PublicFunction(t MyType) (x int, err error) {
	return int(t.num), nil
}

type Dumper interface {
	Load(...interface{})
	Dump() string
}

type IntDumper struct {
	mem []int
}

func (i IntDumper) Load(data ...interface{}) {
	for _, entry := range data {
		i.mem = append(i.mem, entry.(int))
	}
}

func (i IntDumper) Dump() string {
	out := ""
	for _, number := range i.mem {
		out += fmt.Sprintf("%d ", number)
	}
	return out
}