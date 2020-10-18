package fx

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRetry(t *testing.T) {
	// 默认重试3次，因为一直返回错，所以最终 error 不为空
	assert.NotNil(t, DoWithRetries(func() error {
		return errors.New("any")
	}))

	// 试到第三次不再有错，所以最终没有 error
	var times int
	assert.Nil(t, DoWithRetries(func() error {
		times++
		if times == defaultRetryTimes {
			return nil
		}
		return errors.New("any")
	}))

	// 试到第四次才能没错，但是最多只是三次，所以还是有错
	times = 0
	assert.NotNil(t, DoWithRetries(func() error {
		times++
		if times == defaultRetryTimes+1 {
			return nil
		}
		return errors.New("any")
	}))

	// 自定义可重试6次，最后一次无错，所以最终无错
	var total = 2 * defaultRetryTimes
	times = 0
	assert.Nil(t, DoWithRetries(func() error {
		times++
		if times == total {
			return nil
		}
		return errors.New("any")
	}, WithRetries(total)))
}
