package fx

import "github.com/z-sdk/goa/lib/errorx"

const defaultRetryTimes = 3

type (
	retryOption struct {
		times int
	}

	RetryOption func(*retryOption)
)

func newRetryOption() *retryOption {
	return &retryOption{times: defaultRetryTimes}
}

// WithRetries 自定义重试次数
func WithRetries(times int) RetryOption {
	return func(option *retryOption) {
		option.times = times
	}
}

// DoWithRetries 带有重试次数地执行函数
func DoWithRetries(fn func() error, opts ...RetryOption) error {
	options := newRetryOption()
	for _, opt := range opts {
		opt(options)
	}

	var es errorx.Errors
	for i := 0; i < options.times; i++ {
		if err := fn(); err != nil {
			es.Add(err)
		} else {
			return nil
		}
	}

	return es.Error()
}
