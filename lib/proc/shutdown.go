package proc

import (
	"goa/lib/logx"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	shutdownListeners = new(ListenerManager)
	waitTime          = 5 * time.Second
)

// AddShutdownListener 添加一个程序关闭监听器
func AddShutdownListener(listener func()) (waitForCalled func()) {
	return shutdownListeners.add(listener)
}

// gracefulStop 平滑停止程序（为关闭类监听器的执行留有时间）
func gracefulStop(signals chan os.Signal) {
	signal.Stop(signals)

	logx.Info("得到终止信号 SIGTERM，程序关闭中...")

	// 通知关闭类监听器，如定时执行器在关闭前进行任务 Flush
	shutdownListeners.notify()

	// 静候5秒，等待监听器处理完毕。
	time.Sleep(waitTime)

	// 关闭程序
	logx.Info("已等待 %v 秒，即将强制杀死该进程", waitTime)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
