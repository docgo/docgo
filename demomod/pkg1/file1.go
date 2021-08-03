// This is some text I wrote for the package numbered one
// in the original file docstring. It follows the standard
// Go practice for documenting the purpose of a package.
package pkg1

import (
	"fmt"
	"example.com/demomod/pkg1/pkg2"
)

type MyType struct {
	num float64
}

// Converts MyType into an int. Example:
//   ExampleFn(MyType(3)) == 3, nil
// This text goes after the code.
func ExampleFn(t MyType) (x int, err error) {
	return int(t.num), nil
}

func (i IntDumper) Load(data ...interface{}) {
	for _, entry := range data {
		i.mem = append(i.mem, entry.(int))
	}
}

func (i IntDumper) Dump() string {
	out := ""
	for _, number := range i.mem {
		number = pkg2.MulTwo(pkg2.CreateNum(number))
		out += fmt.Sprintf("%d ", number)
	}
	return out
}