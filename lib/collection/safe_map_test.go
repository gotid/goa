package collection

import (
	"github.com/stretchr/testify/assert"
	"goa/lib/stringx"
	"testing"
)

func TestSafeMap(t *testing.T) {
	tests := []struct {
		size      int
		exception int
	}{
		{100000, 2000},
		{100000, 50},
	}
	for _, test := range tests {
		t.Run(stringx.Rand(), func(t *testing.T) {
			testSafeMapWithParameters(t, test.size, test.exception)
		})
	}
}

func testSafeMapWithParameters(t *testing.T, size, exception int) {
	m := NewSafeMap()

	for i := 0; i < size; i++ {
		m.Set(i, i)
	}
	for i := 0; i < size; i++ {
		if i%exception != 0 {
			m.Del(i)
		}
	}

	assert.Equal(t, size/exception, m.Size())

	for i := size; i < size<<1; i++ {
		m.Set(i, i)
	}
	for i := size; i < size<<1; i++ {
		if i%exception != 0 {
			m.Del(i)
		}
	}

	for i := 0; i < size<<1; i++ {
		val, ok := m.Get(i)
		if i%exception == 0 {
			assert.True(t, ok)
			assert.Equal(t, i, val.(int))
		} else {
			assert.False(t, ok)
		}
	}
}
