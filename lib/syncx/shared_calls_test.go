package syncx

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCachedCalls_Get(t *testing.T) {
	calls := NewSharedCalls()
	result, hit, err := calls.Do("ping", func() (result interface{}, err error) {
		return "pong", nil
	})
	got := fmt.Sprintf("%v (%T) %t", result, result, hit)
	expect := "pong (string) false"
	if got != expect {
		t.Errorf("Do = %v; expect %v", got, expect)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestCachedCalls_GetErr(t *testing.T) {
	calls := NewSharedCalls()
	someErr := errors.New("一些错误")
	result, _, err := calls.Do("ping", func() (result interface{}, err error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("返回 error = %v; 期望 someErr", someErr)
	}
	if result != nil {
		t.Errorf("期望nil，返回 %v", err)
	}
}

func TestExclusiveCallDoDupSuppress(t *testing.T) {
	calls := NewSharedCalls()
	c := make(chan string)
	var callCounter int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&callCounter, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			result, _, err := calls.Do("ping", fn)
			if err != nil {
				t.Errorf("Do error: %result", err)
			}
			if result.(string) != "pong" {
				t.Errorf("得到 %q; 期望 %q", result, "pong")
			}
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // let goroutines above block
	c <- "pong"
	wg.Wait()
	if got := atomic.LoadInt32(&callCounter); got != 1 {
		t.Errorf("实际调用次数 = %d; 期望 1", got)
	}
}
