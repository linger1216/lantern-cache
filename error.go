package lantern

import "fmt"

var (
	ErrorCopy                   = fmt.Errorf("invalid copy")
	ErrorBloomFilterInvalidPara = fmt.Errorf("ErrorBloomFilterInvalidPara")
	ErrorBloomFilterInvalidSize = fmt.Errorf("ErrorBloomFilterInvalidSize")

	ErrorBitsetInvalid = fmt.Errorf("ErrorBitsetInvalid")
	ErrorNotPowerOfTwo = fmt.Errorf("ErrorNotPowerOfTwo")

	ErrorCostTooLarge = fmt.Errorf("ErrorCostTooLarge")

	// node
	ErrorNodeFull    = fmt.Errorf("ErrorNodeFull")
	ErrorNodeReadEof = fmt.Errorf("ErrorNodeReadEof index")

	ErrorInvalidPara  = fmt.Errorf("ErrorInvalidPara")
	ErrorNoExpiration = fmt.Errorf("ErrorNoExpiration")
	ErrorNoEntry      = fmt.Errorf("ErrorNoEntry")
)
