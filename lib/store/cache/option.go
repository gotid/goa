package cache

import "time"

const (
	defaultExpires         = time.Hour * 24 * 7
	defaultNotFoundExpires = time.Minute // 防缓存穿透，设置未找到记录一分钟过期
)

type (
	Options struct {
		Expires         time.Duration
		NotFoundExpires time.Duration
	}

	Option func(o *Options)
)

func newOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}

	if o.Expires <= 0 {
		o.Expires = defaultExpires
	}
	if o.NotFoundExpires <= 0 {
		o.NotFoundExpires = defaultNotFoundExpires
	}

	return o
}

func WithExpires(expires time.Duration) Option {
	return func(o *Options) {
		o.Expires = expires
	}
}

func WithNotFoundExpires(expires time.Duration) Option {
	return func(o *Options) {
		o.NotFoundExpires = expires
	}
}
