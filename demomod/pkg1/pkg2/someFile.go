// This is the second package in the
// module and it contains even more code.
package pkg2

type Num struct {
	int
}

// Creates a Num from an integer
func CreateNum(x int) Num {
	return Num{x}
}

// Multiply a number by two
func MulTwo(x Num) int {
	return x.int * 2
}

type Dumper float64