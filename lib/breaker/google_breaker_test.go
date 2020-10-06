package breaker

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"goa/lib/container"
	"goa/lib/mathx"
	"math"
	"math/rand"
	"testing"
	"time"
)

const (
	testBuckets  = 10
	testInterval = 10 * time.Millisecond
)

func TestGoogleBreakerClose(t *testing.T) {
	b := getGoogleBreaker()

	markSuccess(b, 80)
	assert.Nil(t, b.accept())

	markSuccess(b, 120)
	assert.Nil(t, b.accept())

	win := b.stat.Win()
	fmt.Println(win.Buckets())
}

func TestGoogleBreaker_Open(t *testing.T) {
	b := getGoogleBreaker()
	markSuccess(b, 10)
	assert.Nil(t, b.accept())
	markFailed(b, 100) // 模拟很多个失败请求，从而使用断路器保护
	time.Sleep(2 * testInterval)
	win := b.stat.Win()
	fmt.Println(win.Buckets())

	// 断路器
	verify(t, func() bool {
		return b.accept() != nil
	})
}

func TestGoogleBreakerFallback(t *testing.T) {
	b := getGoogleBreaker()

	markSuccess(b, 1)
	assert.Nil(t, b.accept())

	markFailed(b, 100000)
	time.Sleep(2 * testInterval)

	win := b.stat.Win()
	fmt.Println(win.Buckets())

	verify(t, func() bool {
		return b.doReq(func() error {
			return errors.New("模拟请求出错啦")
		}, func(err error) error {
			// 模拟补救，无错返回
			return nil
		}, defaultAcceptable) == nil
	})
}

func TestGoogleBreakerReject(t *testing.T) {
	b := getGoogleBreaker()
	markSuccess(b, 100)
	assert.Nil(t, b.accept())
	markFailed(b, 10000)
	time.Sleep(testInterval)

	// 判断返回类型
	assert.Equal(t, ErrServiceUnavaliable, b.doReq(func() error {
		return ErrServiceUnavaliable
	}, nil, defaultAcceptable))

	win := b.stat.Win()
	fmt.Println(win.Buckets())
}

func TestGoogleBreakerAcceptable(t *testing.T) {
	b := getGoogleBreaker()
	errAccetable := errors.New("某种特定的错误")
	assert.Equal(t, errAccetable, b.doReq(func() error {
		return errAccetable
	}, nil, func(err error) bool {
		// 模拟错误可以接受
		return err == errAccetable
	}))
	win := b.stat.Win()
	fmt.Println(win.Buckets())
	var total int64
	b.stat.Reduce(func(b *container.Bucket) {
		total += int64(b.Accepts)
	})
	fmt.Println("接受次数", total)
}

func TestGoogleBreakerNotAcceptable(t *testing.T) {
	b := getGoogleBreaker()
	errAccetable := errors.New("某种特定的错误")
	assert.Equal(t, errAccetable, b.doReq(func() error {
		return errAccetable
	}, nil, func(err error) bool {
		// 模拟错误不可以接受
		return err != errAccetable
	}))
	win := b.stat.Win()
	fmt.Println(win.Buckets())
	var total int64
	b.stat.Reduce(func(b *container.Bucket) {
		total += int64(b.Accepts)
	})
	fmt.Println("接受次数", total)
}

func TestGoogleBreakerPanic(t *testing.T) {
	b := getGoogleBreaker()
	assert.Panics(t, func() {
		err := b.doReq(func() error {
			panic("失败好痛")
		}, nil, defaultAcceptable)
		fmt.Println(err)
	})
	win := b.stat.Win()
	fmt.Println(win.Buckets())
}

func TestGoogleBreakerHalfOpen(t *testing.T) {
	b := getGoogleBreaker()
	assert.Nil(t, b.accept())
	t.Run("接受单个失败/接受", func(t *testing.T) {
		markFailed(b, 10000)
		time.Sleep(2 * time.Millisecond)
		verify(t, func() bool {
			return b.accept() != nil
		})
	})
	t.Run("接受单个失败/允许", func(t *testing.T) {
		markFailed(b, 10000)
		time.Sleep(2 * time.Millisecond)
		verify(t, func() bool {
			_, err := b.allow()
			return err != nil
		})
	})
	time.Sleep(testInterval * testBuckets)
	t.Run("接受单个成功", func(t *testing.T) {
		assert.Nil(t, b.accept())
		markSuccess(b, 10000)
		verify(t, func() bool {
			return b.accept() == nil
		})
	})
}

func TestGoogleBreakerSelfProtection(t *testing.T) {
	t.Run("总请求小于100次", func(t *testing.T) {
		b := getGoogleBreaker()
		markFailed(b, 4)
		time.Sleep(testInterval)
		assert.Nil(t, b.accept())
	})
	t.Run("总请求大于100次，总请求小于2*success?", func(t *testing.T) {
		b := getGoogleBreaker()
		size := rand.Intn(10000)
		success := int(math.Ceil(float64(size))) + 1
		fmt.Printf("size: %d, success: %d, failed: %d", size, success, size-success)
		markSuccess(b, success)
		markFailed(b, size-success)
		assert.Nil(t, b.accept())
	})
}

func TestGoogleBreakerHistory(t *testing.T) {
	var b *googleBreaker
	var requests, accepts int64

	sleep := testInterval
	t.Run("接受数 == 请求数", func(t *testing.T) {
		b = getGoogleBreaker()
		markSuccessWithDuration(b, 10, sleep/2)
		requests, accepts = b.history()
		assert.Equal(t, int64(10), requests)
		assert.Equal(t, int64(10), accepts)
	})

	t.Run("失败数 == 请求数", func(t *testing.T) {
		b = getGoogleBreaker()
		markFailedWithDuration(b, 10, sleep/2)
		requests, accepts = b.history()
		assert.Equal(t, int64(10), requests)
		assert.Equal(t, int64(0), accepts)
	})

	t.Run("接受数 = 1/2 * 请求数, 失败数 = 1/2 * 请求数", func(t *testing.T) {
		b = getGoogleBreaker()
		markFailedWithDuration(b, 5, sleep/2)
		markSuccessWithDuration(b, 5, sleep/2)
		requests, accepts = b.history()
		assert.Equal(t, int64(10), requests)
		assert.Equal(t, int64(5), accepts)
	})

	t.Run("自动重设滚动窗口计数器", func(t *testing.T) {
		b = getGoogleBreaker()
		time.Sleep(testInterval * testBuckets)
		requests, accepts = b.history()
		assert.Equal(t, int64(0), requests)
		assert.Equal(t, int64(0), accepts)
	})
}

func BenchmarkGoogleBreakerAllow(b *testing.B) {
	breaker := getGoogleBreaker()
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		breaker.accept()
		if i%2 == 0 {
			breaker.markSuccess()
		} else {
			breaker.markFailure()
		}
	}
}

func getGoogleBreaker() *googleBreaker {
	return &googleBreaker{
		k:     5,
		state: StateClosed,
		stat:  container.NewRollingWindow(testBuckets, testInterval),
		prob:  mathx.NewProb(),
	}
}

func markSuccess(b *googleBreaker, count int) {
	for i := 0; i < count; i++ {
		promise, err := b.allow()
		if err != nil {
			break
		}
		promise.Accept()
	}
}

func markFailed(b *googleBreaker, count int) {
	for i := 0; i < count; i++ {
		p, err := b.allow()
		if err == nil {
			p.Reject()
		}
	}
}

func verify(t *testing.T, fn func() bool) {
	var count int
	for i := 0; i < 100; i++ {
		if fn() {
			count++
		}
	}
	fmt.Printf("次数: %d\n", count)
	assert.True(t, count >= 80, fmt.Sprintf("应大于80，实际为 %d", count))
}

func markSuccessWithDuration(b *googleBreaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.markSuccess()
		time.Sleep(sleep)
	}
}

func markFailedWithDuration(b *googleBreaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.markFailure()
		time.Sleep(sleep)
	}
}
