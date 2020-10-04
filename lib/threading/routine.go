package threading

import (
	"bytes"
	"goa/lib/rescue"
	"runtime"
	"strconv"
)

// GoSafe 并发安全执行函数：panic 则输出错误堆栈
func GoSafe(fn func()) {
	go RunSafe(fn)
}

// RunSafe 安全运行：panic 则输出错误堆栈
func RunSafe(fn func()) {
	defer rescue.Recover()

	fn()
}

// 仅用于调试，永不用于生产环境
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	id, _ := strconv.ParseUint(string(b), 10, 64)
	return id
}
