package breaker

import (
	"fmt"
	"goa/lib/collection"
	"goa/lib/proc"
	"goa/lib/stat"
)

type (
	internalPromise interface {
		Accept()
		Reject()
	}

	// 内部节流阀接口，如 google_throttle 实现
	internalThrottle interface {
		allow() (internalPromise, error)
		doReq(req Request, fallback Fallback, acceptable Acceptable) error
	}

	promiseWithReason struct {
		promise internalPromise
		errWin  *collection.ErrorWindow
	}

	loggedThrottle struct {
		name string
		internalThrottle
		errWin *collection.ErrorWindow
	}
)

func newLoggedThrottle(name string, t internalThrottle) loggedThrottle {
	return loggedThrottle{
		name:             name,
		internalThrottle: t,
		errWin:           collection.NewErrorWindow(),
	}
}

func (t loggedThrottle) allow() (Promise, error) {
	promise, err := t.internalThrottle.allow()
	return promiseWithReason{
		promise: promise,
		errWin:  t.errWin,
	}, t.logError(err)
}

func (t loggedThrottle) doReq(req Request, fallback Fallback, acceptable Acceptable) error {
	return t.logError(t.internalThrottle.doReq(req, fallback, func(err error) bool {
		accept := acceptable(err)
		if !accept {
			t.errWin.Add(err.Error())
		}
		return accept
	}))
}

func (t loggedThrottle) logError(err error) error {
	if err == ErrServiceUnavaliable {
		stat.Report(fmt.Sprintf(
			"进程(%s/%d), 调用者名称: %s, 断路器已打开，请求被丢弃\n最新错误：\n%s",
			proc.ProcessName(), proc.Pid(), t.name, t.errWin))
	}
	return err
}

func (p promiseWithReason) Accept() {
	p.promise.Accept()
}

func (p promiseWithReason) Reject(reason string) {
	p.errWin.Add(reason)
	p.promise.Reject()
}
