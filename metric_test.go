package lantern

import (
	"fmt"
	"testing"
)

func Test_newMetrics(t *testing.T) {
	s := newStats()
	s.hit.add(1)
	fmt.Println(s.hit.get())
}
