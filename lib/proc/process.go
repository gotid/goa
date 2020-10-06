package proc

import (
	"os"
	"path/filepath"
)

var (
	procName string
	pid      int
)

// init 初始化当前程序进程名称和进程编号
func init() {
	procName = filepath.Base(os.Args[0])
	pid = os.Getpid()
}

func ProcessName() string {
	return procName
}

func Pid() int {
	return pid
}
