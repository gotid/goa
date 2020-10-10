package breaker

import (
	"errors"
	"github.com/z-sdk/goa/lib/stringx"
)

const (
	StateClosed State = iota // 断路器关闭
	StateOpen                // 断路器打开
)

var ErrServiceUnavaliable = errors.New("断路器已打开，服务不可用")

type (
	State      = int32                     // 断路器状态
	Request    func() error                // 待执行的请求
	Acceptable func(reqError error) bool   // 判断错误是否可接受的函数
	Fallback   func(acceptErr error) error // 应急备用函数
	Option     func(b *breaker)            // 断路器可选项应用函数

	Breaker interface {
		// Name 是 netflixBreaker 断路器的名称
		Name() string

		// Allow 检查请求是否被允许。
		// 若允许，则返回 Promise，调用者需在成功时调用 promise.Resolve()，失败时调用 promise.Reject()。
		// 若不允许，则返回 ErrServiceUnavaliable。
		Allow() (Promise, error)

		// Do 运行指定请求，如果断路器允许的话。
		// Do 直接返回错误，如果断路器拒绝的话。
		// 如果处理中出现错误，断路器则抛出错误。
		Do(req Request) error

		// DoWithAcceptable 运行指定请求，如果断路器允许的话。
		// DoWithAcceptable 直接返回错误，如果断路器拒绝的话。
		// 如果处理中出现错误，断路器则抛出错误。
		// acceptable 自定义检查是否可调用，即使 err 不是 nil。
		DoWithAcceptable(req Request, acceptable Acceptable) error

		// DoWithFailback 运行指定请求，如果断路器允许的话。
		// DoWithFailback 直接返回错误，如果断路器拒绝的话。
		// 如果处理中出现错误，断路器则抛出错误。
		// fallback 根据 err 提供备用应急计划。
		DoWithFailback(req Request, fallback Fallback) error

		// DoWithFailbackAcceptable 如果断路器允许的话，运行指定请求。
		// 如果断路器拒绝的话，调用 DoWithFailback 运行应急备用的计划。
		// 如果处理中出现错误，断路器则抛出错误。
		// acceptable 自定义检查是否可调用，即使 err 不是 nil。
		DoWithFailbackAcceptable(req Request, fallback Fallback, acceptable Acceptable) error
	}

	Promise interface {
		Accept()
		Reject(reason string)
	}

	throttle interface {
		allow() (Promise, error)
		doReq(req Request, fallback Fallback, acceptable Acceptable) error
	}

	breaker struct {
		name string
		throttle
	}
)

func NewBreaker(opts ...Option) Breaker {
	var b breaker
	for _, opt := range opts {
		opt(&b)
	}
	if len(b.name) == 0 {
		b.name = stringx.Rand()
	}
	b.throttle = newLoggedThrottle(b.name, newGoogleBreaker())
	return &b
}

func WithName(name string) Option {
	return func(b *breaker) {
		b.name = name
	}
}

func (b breaker) Name() string {
	return b.name
}

func (b breaker) Allow() (Promise, error) {
	return b.throttle.allow()
}

func (b breaker) Do(req Request) error {
	return b.throttle.doReq(req, nil, defaultAcceptable)
}

func (b breaker) DoWithAcceptable(req Request, acceptable Acceptable) error {
	return b.throttle.doReq(req, nil, acceptable)
}

func (b breaker) DoWithFailback(req Request, fallback Fallback) error {
	return b.throttle.doReq(req, fallback, defaultAcceptable)
}

func (b breaker) DoWithFailbackAcceptable(req Request, fallback Fallback, acceptable Acceptable) error {
	return b.throttle.doReq(req, fallback, acceptable)
}

func defaultAcceptable(err error) bool {
	return err == nil
}
