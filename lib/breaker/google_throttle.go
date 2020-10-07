package breaker

import (
	"fmt"
	"goa/lib/container"
	"goa/lib/logx"
	"goa/lib/mathx"
	"math"
	"sync/atomic"
	"time"
)

const (
	window  = 10 * time.Second // 一个滚动窗时长，默认10秒
	buckets = 1                // 一个滚动窗允许通过的桶数，默认40个

	K          = 1.1 // 请求接受比例，越大接受度则越高，越小自适应节流则越积极
	protection = 1   // 保护的请求数
)

type (
	// googleThrottle 是谷歌处理 overload 过载问题的节流阀实现。
	// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
	googleThrottle struct {
		k     float64                  // 请求接受比例
		state int32                    // 断路器跳闸状态
		stat  *container.RollingWindow // 统计窗口计数器（采用滚窗算法）
		prob  *mathx.Prob
	}

	googlePromise struct {
		b *googleThrottle
	}
)

func newGoogleBreaker() *googleThrottle {
	bucketDuration := time.Duration(int64(window) / int64(buckets)) // 单桶时长，默认250毫秒
	statWindow := container.NewRollingWindow(buckets, bucketDuration)
	return &googleThrottle{
		k:     K,
		state: StateClosed,
		stat:  statWindow,
		prob:  mathx.NewProb(),
	}
}

// 先看断路器是否接受，如果接受则发返回 googlePromise 等待处理
func (t *googleThrottle) allow() (internalPromise, error) {
	if err := t.accept(); err != nil {
		return nil, err
	}

	// 接受成功，则返回googlePromise，由其标记结果
	return googlePromise{b: t}, nil
}

func (t *googleThrottle) doReq(req Request, fallback Fallback, acceptable Acceptable) error {
	// 首先，试探断路器是否接受请求
	if err := t.accept(); err != nil {
		// 尝试采用应急方案
		if fallback != nil {
			return fallback(err)
		} else {
			return err
		}
	}

	// 最后，如果有错则标记为失败，并抛出异常
	defer func() {
		if e := recover(); e != nil {
			t.markFailure()
			panic(e)
		}
	}()

	// 然后，执行请求，根据返回错误的可接受度标记结果
	reqError := req()
	if acceptable(reqError) {
		t.markSuccess()
	} else {
		t.markFailure()
	}

	// 不论错误是否可接受，都要返回
	return reqError
}

// accept 根据客户端请求拒绝率返回错误
func (t *googleThrottle) accept() error {
	requests, accepts := t.history()

	// 计算客户端请求拒绝率
	// https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	// 常量 K 为1.5意味着，请求150次只成功100次，则
	weightedAccepts := t.k * float64(accepts)
	//droppedRequests := float64(requests-protection) - weightedAccepts
	droppedRequests := float64(requests) - weightedAccepts
	dropRatio := math.Max(0, droppedRequests/float64(requests+1))

	fmt.Printf("dropRation = max(0, (%d-%.0f*%d)/(%d+1)) -> %f\n", requests, t.k, accepts, requests, dropRatio)

	// 无需拒绝
	if dropRatio <= 0 {
		if atomic.LoadInt32(&t.state) == StateOpen {
			logx.Error("关闭断路器")
			atomic.CompareAndSwapInt32(&t.state, StateOpen, StateClosed)
		}
		return nil
	}

	// 未开断路器，则需打开
	if atomic.LoadInt32(&t.state) == StateClosed {
		logx.Error("打开断路器")
		atomic.CompareAndSwapInt32(&t.state, StateClosed, StateOpen)

	}

	// 并非每次阻断，而是随机拦截，以此给后端重生的机会
	if t.prob.TrueOnProb(dropRatio) {
		return ErrServiceUnavaliable
	}

	return nil
}

// 历史总数（请求多少次，同意多少次）
func (t *googleThrottle) history() (requests int64, accepts int64) {
	t.stat.Reduce(func(b *container.Bucket) {
		requests += int64(b.Requests)
		accepts += int64(b.Accepts)
	})
	return
}

func (t *googleThrottle) markSuccess() {
	t.stat.Add(1)
}

func (t *googleThrottle) markFailure() {
	t.stat.Add(0)
}

func (p googlePromise) Accept() {
	p.b.markSuccess()
}

func (p googlePromise) Reject() {
	p.b.markFailure()
}
