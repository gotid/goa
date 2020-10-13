package errorx

import "sync/atomic"

type AtomicError struct {
	err atomic.Value
}

func (ae *AtomicError) Set(err error) {
	ae.err.Store(err)
}

func (ae *AtomicError) Load() error {
	if x := ae.err.Load(); x != nil {
		return x.(error)
	}
	return nil
}
