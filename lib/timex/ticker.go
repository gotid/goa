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

func (r ticker) Chan() <-chan time.Time {
	panic("implement me")
}

func NewTicker(d time.Duration) Ticker {
	return &ticker{
		time.NewTicker(d),
	}
}
