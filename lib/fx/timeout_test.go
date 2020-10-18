package fx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWithPanic(t *testing.T) {
	assert.Panics(t, func() {
		_ = DoWithTimeout(func() error {
			panic("ouch")
		}, 50*time.Millisecond)
	})
}

func TestWithTimeout(t *testing.T) {
	assert.Equal(t, ErrTimeout, DoWithTimeout(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, time.Millisecond))
}

func TestWithoutTimeout(t *testing.T) {
	assert.Nil(t, DoWithTimeout(func() error {
		return nil
	}, 50*time.Millisecond))
}

func TestWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := DoWithTimeout(func() error {
		time.Sleep(time.Minute)
		return nil
	}, time.Second, WithContext(ctx))
	assert.Equal(t, ErrCanceled, err)
}
