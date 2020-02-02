package lantern_cache

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unsafe"
)

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)

	for start := 0; start < length; start++ {
		t := rand.Intn(2)
		if t == 0 {
			// A-Z
			rs = append(rs, string(rand.Intn(26)+65))
		} else {
			// a-z
			rs = append(rs, string(rand.Intn(26)+97))
		}
	}
	return strings.Join(rs, "")
}

func RandomNumber(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max) + min
}

func isPowerOfTwo(number uint32) bool {
	return (number & (number - 1)) == 0
}

func humanSize(b int64) string {
	if b/(1024*1024*1024) > 0 {
		return fmt.Sprintf("%dG", b/(1024*1024*1024))
	}
	if b/(1024*1024) > 0 {
		return fmt.Sprintf("%dM", b/(1024*1024))
	}
	if b/(1024) > 0 {
		return fmt.Sprintf("%dK", b/(1024))
	}
	return fmt.Sprintf("%dB", b)
}

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
