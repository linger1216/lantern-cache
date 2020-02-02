package lantern_cache

import "fmt"

var (
	ErrorInValidStackType = fmt.Errorf("invalid slot stack type should be uint32")

	ErrorInvalidEntry = fmt.Errorf("invalid entry")

	// common
	ErrorCopy = fmt.Errorf("invalid copy")

	// chunk
	ErrorChunkAlloc = fmt.Errorf("alloc chunk error")

	// bucket
	ErrorEntryTooBig          = fmt.Errorf("value too big")
	ErrorChunkIndexOutOfRange = fmt.Errorf("chunk index out of range")

	// cache
	ErrorNotFound    = fmt.Errorf("not found")
	ErrorValueExpire = fmt.Errorf("value expire")

	// slot
	ErrorSlotDelete         = fmt.Errorf("")
	ErrorSlotCapacityExceed = fmt.Errorf("capacity full")
	ErrorSlotStackEmpty     = fmt.Errorf("")

	// ring
	ErrorRingDelete = fmt.Errorf("")
)
