// 营救——恢复弥补包
package rescue

import "github.com/z-sdk/goa/lib/logx"

// Recover 恢复弥补函数：执行一组清理函数并尝试输出错误堆栈
func Recover(cleanUps ...func()) {
	for _, cleanUp := range cleanUps {
		cleanUp()
	}

	// 处理 panic 日志
	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
