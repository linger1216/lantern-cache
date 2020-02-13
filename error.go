package lantern

import "fmt"

var (
	ErrorCopy                   = fmt.Errorf("invalid copy")
	ErrorBloomFilterInvalidPara = fmt.Errorf("ErrorBloomFilterInvalidPara")
	ErrorBloomFilterInvalidSize = fmt.Errorf("ErrorBloomFilterInvalidSize")

	ErrorBitsetInvalid = fmt.Errorf("ErrorBitsetInvalid")
	ErrorNotPowerOfTwo = fmt.Errorf("ErrorNotPowerOfTwo")

	ErrorCostTooLarge = fmt.Errorf("ErrorCostTooLarge")
)
