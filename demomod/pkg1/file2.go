package pkg1

type Dumper interface {
	Load(...interface{})
	Dump() string
}

type IntDumper struct {
	mem []int
}

