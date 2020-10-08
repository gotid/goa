package hash

import (
	"crypto/md5"
	"fmt"
	"github.com/spaolacci/murmur3"
)

// Hash 字节大于10，性能远高于MD5，且碰撞率更低
func Hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}

func MD5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

func MD5Hex(data []byte) string {
	return fmt.Sprintf("%x", MD5(data))
}
