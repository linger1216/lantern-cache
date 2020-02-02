package lantern_cache

import (
	"fmt"
	"strings"
)

type Hasher interface {
	Hash([]byte) uint64
}

func NewHasher(policy string) Hasher {
	if len(policy) == 0 {
		policy = "fnv"
	}
	policy = strings.ToLower(policy)
	switch policy {
	case "fnv":
		return newFowlerNollVoHasher()
	default:
		panic(fmt.Errorf("hash can't support policy %s", policy))
	}
}
