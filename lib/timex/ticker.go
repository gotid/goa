package timex

import (
	"time"
)

type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type ticker struct {
	*time.Ticker
}

func (t ticker) Chan() <-chan time.Time {
	return t.C
}

func NewTicker(d time.Duration) Ticker {
	return &ticker{
		time.NewTicker(d),
	}
}
