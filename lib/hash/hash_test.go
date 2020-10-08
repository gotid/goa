package hash

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	text      = "hello, world!\n"
	md5Digest = "910c8bc73110b0cd1bc5d2bcae782511"
)

func TestMD5(t *testing.T) {
	actual := fmt.Sprintf("%x", Md5([]byte(text)))
	assert.Equal(t, md5Digest, actual)
}

func TestHash(t *testing.T) {
	fmt.Println(Hash([]byte("1")))
	fmt.Println(Hash([]byte("2")))
	fmt.Println(Hash([]byte("12")))
	fmt.Println(Hash([]byte("999999999999999999")))
}

func BenchmarkMD5Hex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MD5([]byte(text))
	}
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Hash([]byte(text))
	}
}
