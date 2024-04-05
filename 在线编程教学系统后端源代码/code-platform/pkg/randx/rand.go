package randx

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"
)

var maxInt = big.NewInt(int64(len(randTable)))

var (
	builderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
)

const (
	randTable = "WXSzjPliuVGUwkqfspbg9KDvMyoRNE0LI2T4CthJnYmcr1FZ6O5QAd7eB83xaH"
)

func NewRandCode(length int) (string, error) {
	b := builderPool.Get().(*strings.Builder)
	b.Grow(length)
	defer func() {
		b.Reset()
		builderPool.Put(b)
	}()

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			return "", err
		}
		b.WriteByte(randTable[n.Uint64()])
	}

	return b.String(), nil
}
