// This is the second package in the
// module and it contains even more code.
package pkg2

// An int wrapper, containing an unnamed int field
type Num struct {
	int
}

// Casts an int into the Num struct that wraps it
func CreateNum(x int) Num {
	return Num{x}
}

// Multiplies a Num by two and converts it to a Go-style int
func MulTwo(x Num) int {
	return x.int * 2
}