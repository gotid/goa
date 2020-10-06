package stringx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	Seed(time.Now().UnixNano())
	assert.True(t, len(Rand()) > 0)
	assert.True(t, len(RandId()) > 0)
	fmt.Println(Rand())
	fmt.Println(RandId())

	const size = 10
	assert.True(t, len(Randn(size)) == size)
	fmt.Println(Randn(size))
}

func BenchmarkRand(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Randn(10)
		}
	})
}

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Randn(10)
	}
}
