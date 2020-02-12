package lantern

const (
	CleanCount                = 1 << 24
	chunkSize                 = 64 * 1024
	chunksPerAlloc            = 1024
	MaxKeySize                = 1 << 16
	MaxValueSize              = 1 << 16 // 64k
	EntryTimeStampFieldSizeOf = 8
	EntryKeyFieldSizeOf       = 2
	EntryValueFieldSizeOf     = 2
	EntryHeadFieldSizeOf      = EntryTimeStampFieldSizeOf + EntryKeyFieldSizeOf + EntryValueFieldSizeOf
	OffsetSizeOf              = 40
	LoopSizeOf                = 64 - OffsetSizeOf
)
