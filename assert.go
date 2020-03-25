package lantern

import "fmt"

func assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func ensure(condition bool) {
	assert(condition, "something wrong")
}
