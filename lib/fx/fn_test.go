package fx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFrom(t *testing.T) {
	const N = 5
	var count int32
	var wait sync.WaitGroup
	wait.Add(1)
	From(func(source chan<- interface{}) {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for i := 0; i < 2*N; i++ {
			select {
			case source <- i:
				fmt.Println("add", 1)
				atomic.AddInt32(&count, 1)
			case <-ticker.C:
				wait.Done()
				return
			}
		}
	}).Buffer(N).ForAll(func(pipe <-chan interface{}) {
		wait.Wait()
		// 要多等一个，才能发送到通道
		assert.Equal(t, int32(N+1), atomic.LoadInt32(&count))
		fmt.Println(N+1, atomic.LoadInt32(&count))
	})
}

func TestJust(t *testing.T) {
	var result int
	result2, err := Just(1, 2, 3, 4).Buffer(-1).Reduce(func(pipe <-chan interface{}) (interface{}, error) {
		for item := range pipe {
			result += item.(int)
		}
		return result, nil
	})
	fmt.Println(result)
	fmt.Println(result2)
	fmt.Println(err)
}
