package timex

import (
	"errors"
	"github.com/z-sdk/goa/lib/lang"
	"time"
)

type FakeTicker interface {
	Ticker
	Tick()
	Done()
	Wait(d time.Duration) error
}

type fakeTicker struct {
	c    chan time.Time
	done chan lang.PlaceholderType
}

func (ft fakeTicker) Chan() <-chan time.Time {
	return ft.c
}

func (ft fakeTicker) Stop() {
	close(ft.c)
}

func (ft fakeTicker) Tick() {
	ft.c <- Time()
}

func (ft fakeTicker) Done() {
	ft.done <- lang.Placeholder
}

func (ft fakeTicker) Wait(d time.Duration) error {
	select {
	case <-time.After(d):
		return errors.New("超时")
	case <-ft.done:
		return nil
	}
}

func NewFakeTicker() FakeTicker {
	return &fakeTicker{
		c:    make(chan time.Time, 1),
		done: make(chan lang.PlaceholderType, 1),
	}
}
