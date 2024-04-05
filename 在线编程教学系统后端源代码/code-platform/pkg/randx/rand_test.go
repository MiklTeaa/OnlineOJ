package randx_test

import (
	"fmt"
	"sync"
	"testing"

	. "code-platform/pkg/randx"

	"github.com/stretchr/testify/require"
)

func TestNewRandCode(t *testing.T) {
	const length = 6

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := NewRandCode(length)
			require.NoError(t, err)

			fmt.Println(code)
		}()
	}
	wg.Wait()
}
