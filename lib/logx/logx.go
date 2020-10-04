package logx

import (
	"encoding/json"
	"fmt"
	"goa/lib/timex"
	"io"
	"log"
	"runtime/debug"
	"sync/atomic"
)

const (
	// 日志级别值
	InfoLevel = iota
	ErrorLevel
	FatalLevel
)

const (
	// 日志级别名称
	infoLevel  = "info"  // 信息级
	errorLevel = "error" // 错误级
	fatalLevel = "fatal" // 重大级
	slowLevel  = "slow"  // 慢级别
	statLevel  = "stat"  // 统计级

	// 日志文件
	accessFilename = "access.log"
	errorFilename  = "error.log"
	fatalFilename  = "fatal.log"
	slowFilename   = "slow.log"
	statFilename   = "stat.log"

	// 日期格式
	timeFormat = "2000-01-01T10:00:00.000Z08"
)

var (
	// 日志类型
	infoLog  io.WriteCloser // 信息日志
	errorLog io.WriteCloser // 错误日志
	fatalLog io.WriteCloser // 重大日志
	slowLog  io.WriteCloser // 慢日志
	statLog  io.WriteCloser // 统计日志
	stackLog io.Writer      // 堆栈日志

	initialized uint32 // 初始状态
	logLevel    uint32 // 日志级别
)

// 日志结构
type entry struct {
	Timestamp string `json:"@timestamp"`
	Level     string `json:"level"`
	Duration  string `json:"duration,omitempty"`
	Content   string `json:"content"`
}

func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

func Info(v ...interface{}) {
	syncInfo(fmt.Sprint(v...))
}

func syncInfo(msg string) {
	if shouldLog(InfoLevel) {
	}
}

// shouldLog 对比日志级别确定是否需要记录
func shouldLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

// 输出错误堆栈
func ErrorStack(v ...interface{}) {
	syncStack(fmt.Sprint(v...))
}

func syncStack(msg string) {
	output(stackLog, errorLevel, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
}

func output(writer io.Writer, level, msg string) {
	outputJson(writer, entry{
		Timestamp: getTimestamp(),
		Level:     level,
		Content:   msg,
	})
}

func outputJson(writer io.Writer, e entry) {
	if content, err := json.Marshal(e); err != nil {
		log.Println(err.Error())
	} else if atomic.LoadUint32(&initialized) == 0 || writer == nil {
		log.Println(string(content))
	} else {
		n, err := writer.Write(append(content, '\n'))
		log.Printf("写日志错误(%d): %v\n", n, err)
	}
}

func getTimestamp() string {
	return timex.Time().Format(timeFormat)
}

func setupLogLevel(c LogConf) {
	switch c.Level {
	case infoLevel:
		SetLevel(InfoLevel)
	case errorLevel:
		SetLevel(ErrorLevel)
	case fatalLevel:
		SetLevel(FatalLevel)
	}
}
