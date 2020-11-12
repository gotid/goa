package fx

import (
	"fmt"
	"github.com/z-sdk/goa/lib/lang"
	"github.com/z-sdk/goa/lib/threading"
	"sort"
	"sync"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

type (
	rxOption struct {
		unlimitedWorkers bool
		workers          int
	}

	GenerateFunc func(source chan<- interface{})
	KeyFunc      func(item interface{}) interface{}
	ForAllFunc   func(pipe <-chan interface{})
	ForEachFunc  func(item interface{})
	WalkFunc     func(item interface{}, pipe chan<- interface{})
	ParallelFunc func(item interface{})
	FilterFunc   func(item interface{}) bool
	MapFunc      func(item interface{}) interface{}
	ReduceFunc   func(pipe <-chan interface{}) (interface{}, error)
	LessFunc     func(a interface{}, b interface{}) bool
	Option       func(opts *rxOption)
)

type Stream struct {
	source <-chan interface{}
}

// From 从生成函数 GenerateFunc 生成数据流 Stream
func From(generate GenerateFunc) Stream {
	source := make(chan interface{})

	threading.GoSafe(func() {
		defer close(source)
		generate(source)
	})

	return Range(source)
}

// Just 转换任意项为数据流 Stream
func Just(items ...interface{}) Stream {
	source := make(chan interface{}, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)

	return Range(source)
}

// Range 将指定通道转换为 Stream 流
func Range(source <-chan interface{}) Stream {
	return Stream{source: source}
}

// Buffer 缓冲数据到一个大小为n的通道
//
func (s Stream) Buffer(n int) Stream {
	if n < 0 {
		n = 0
	}

	source := make(chan interface{}, n)
	go func() {
		for item := range s.source {
			fmt.Println("buffer: source <- ", item)
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Count 计算结果中的元素个数
func (s Stream) Count() (count int) {
	for range s.source {
		count++
	}
	return
}

// Distinct 基于给定的 KeyFunc 移除重复项
func (s Stream) Distinct(fn KeyFunc) Stream {
	source := make(chan interface{})

	threading.GoSafe(func() {
		defer close(source)

		keys := make(map[interface{}]lang.PlaceholderType)
		for item := range s.source {
			key := fn(item)
			if _, ok := keys[key]; !ok {
				source <- item
				keys[key] = lang.Placeholder
			}
		}
	})

	return Range(source)
}

// Group 基于给定的 KeyFun 对各项分组
func (s Stream) Group(fn KeyFunc) Stream {
	groups := make(map[interface{}][]interface{})
	for item := range s.source {
		key := fn(item)
		groups[key] = append(groups[key], item)
	}

	source := make(chan interface{})
	go func() {
		for _, group := range groups {
			source <- group
		}
		close(source)
	}()

	return Range(source)
}

// ForAll 处理来自source的当前所有数据流，不处理后续流。
func (s Stream) ForAll(fn ForAllFunc) {
	fn(s.source)
}

// ForEach 处理当前source中的每个项，不处理后续流。
func (s Stream) ForEach(fn ForEachFunc) {
	for item := range s.source {
		fn(item)
	}
}

// Walk 让调用者处理每个项，可以处理后续流。
func (s Stream) Walk(fn WalkFunc, opts ...Option) Stream {
	option := buildOptions(opts...)
	if option.unlimitedWorkers {
		return s.walkUnlimited(fn, option)
	} else {
		return s.walkLimited(fn, option)
	}
}

func (s Stream) walkUnlimited(fn WalkFunc, option *rxOption) Stream {
	pipe := make(chan interface{}, defaultWorkers)

	go func() {
		var wg sync.WaitGroup

		for {
			item, ok := <-s.source
			if !ok {
				break
			}

			wg.Add(1)
			threading.GoSafe(func() {
				defer wg.Done()
				fn(item, pipe)
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

func (s Stream) walkLimited(fn WalkFunc, option *rxOption) Stream {
	pipe := make(chan interface{}, option.workers)

	go func() {
		var wg sync.WaitGroup
		pool := make(chan lang.PlaceholderType, option.workers)

		for {
			pool <- lang.Placeholder
			item, ok := <-s.source
			if !ok {
				<-pool
				break
			}

			wg.Add(1)
			threading.GoSafe(func() {
				defer func() {
					wg.Done()
					<-pool
				}()

				fn(item, pipe)
			})
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

// Parallel 并行应用给定的 ParallelFunc 到各项。
func (s Stream) Parallel(fn ParallelFunc, opts ...Option) {
	s.Walk(func(item interface{}, pipe chan<- interface{}) {
		fn(item)
	}, opts...).Done()
}

// Filter 通过给定 FilterFunc 过滤项目并返回
func (s Stream) Filter(fn FilterFunc, opts ...Option) Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		if fn(item) {
			pipe <- item
		}
	}, opts...)
}

// Map 转换数据流中的每一项并返回
func (s Stream) Map(fn MapFunc, opts ...Option) Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		pipe <- fn(item)
	}, opts...)
}

// Reduce 是一个调用者处理底层通道的工具方法。
func (s Stream) Reduce(fn ReduceFunc) (interface{}, error) {
	return fn(s.source)
}

// Sort 对来自数据源的项进行排序
func (s Stream) Sort(less LessFunc) Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})

	return Just(items...)
}

// Revers 反转数据流中的元素。
func (s Stream) Reverse() Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	// 反转，官方的方法
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}

	return Just(items...)
}

// Split 切流至前N项
func (s Stream) Split(n int) Stream {
	if n < 1 {
		panic("n 应大于 0")
	}

	source := make(chan interface{})
	go func() {
		var chunk []interface{}
		for item := range s.source {
			chunk = append(chunk, item)
			if len(chunk) == n {
				source <- chunk
				chunk = nil
			}
		}
		if chunk != nil {
			source <- chunk
		}
		close(source)
	}()

	return Range(source)
}

// Done 等待所有上游操作完成。
func (s Stream) Done() {
	for range s.source {
	}
}

func buildOptions(opts ...Option) *rxOption {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func newOptions() *rxOption {
	return &rxOption{
		workers: defaultWorkers,
	}
}
