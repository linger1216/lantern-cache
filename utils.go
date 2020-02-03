package lantern_cache

import (
	"math/rand"
	"strconv"
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

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//	E: Exabyte
//	P: Petabyte
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func humanSize(bytes int64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = "E"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "P"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0B"
	}

	result := strconv.FormatFloat(value, 'f', 2, 64)
	result = strings.TrimSuffix(result, ".00")
	return result + unit
}
