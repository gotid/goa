package stringx

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	idLen          = 8
	defaultRandLen = 8
	letterIdxBits  = 6                    // 6位表示一个字母索引
	letterIdxMax   = 63 / letterIdxBits   // 63位的字母指数拟合
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var src = newLockedSource(time.Now().UnixNano())

type lockedSource struct {
	source rand.Source
	lock   sync.Mutex
}

func newLockedSource(seed int64) *lockedSource {
	return &lockedSource{
		source: rand.NewSource(seed),
	}
}

func (s *lockedSource) Int63() int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.source.Int63()
}

func (s *lockedSource) Seed(seed int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.source.Seed(seed)
}

func Rand() string {
	return Randn(defaultRandLen)
}

func RandId() string {
	b := make([]byte, idLen)
	_, err := crand.Read(b)
	if err != nil {
		return Randn(idLen)
	}

	return fmt.Sprintf("%x%x%x%x", b[0:2], b[2:4], b[4:6], b[6:8])
}

func Randn(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func Seed(seed int64) {
	src.Seed(seed)
}
