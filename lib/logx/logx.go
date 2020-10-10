package logx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/z-sdk/goa/lib/iox"
	"github.com/z-sdk/goa/lib/sysx"
	"github.com/z-sdk/goa/lib/timex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 日志级别值
	InfoLevel = iota
	SlowLevel
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

	// 日志模式
	consoleMode = "console"
	volumeMode  = "volume"

	timeFormat          = "2006-01-02T15:04:05.000Z07" // 日期格式
	callerInnerDepth    = 5                            // 堆栈调用深度
	flags               = 0x0
	backupFileDelimiter = "-" // 日志备份文件分隔符
)

var (
	// 日志类型
	infoLogger  io.WriteCloser // 信息日志
	errorLogger io.WriteCloser // 错误日志
	fatalLogger io.WriteCloser // 重大日志
	slowLogger  io.WriteCloser // 慢日志
	statLogger  io.WriteCloser // 统计日志
	stackLogger io.Writer      // 堆栈日志

	initialized  uint32    // 初始状态
	logLevel     uint32    // 日志级别
	writeConsole bool      // 写控制台
	once         sync.Once // 一次操作对象
	options      logOptions

	ErrLogServiceNameNotSet = errors.New("日志服务名称必须设置")
	ErrLogPathNotSet        = errors.New("日志路径必须设置")
	ErrLogNotInitialized    = errors.New("日志尚未初始化")
)

type (
	// 日志结构
	logEntry struct {
		Timestamp string `json:"@timestamp"`
		Level     string `json:"level"`
		Duration  string `json:"duration,omitempty"`
		Content   string `json:"content"`
	}

	// 日志配置选项
	logOptions struct {
		gzipEnabled           bool
		logStackCooldownMills int
		keepDays              int
	}

	LogOption func(options *logOptions)

	// 用于 durationLogger/traceLogger
	Logger interface {
		Info(...interface{})
		Infof(string, ...interface{})
		Error(...interface{})
		Errorf(string, ...interface{})
		Slow(...interface{})
		Slowf(string, ...interface{})
		WithDuration(time.Duration) Logger
	}
)

// MustSetup 必须成功不能有错，否则直接退出系统
func MustSetup(c LogConf) {
	Must(Setup(c))
}

// Must 必须无错，否则退出程序
func Must(err error) {
	if err != nil {
		msg := formatWithCaller(err.Error(), 3)
		log.Print(msg)
		output(fatalLogger, fatalLevel, msg)
		os.Exit(1)
	}
}

func Setup(c LogConf) error {
	switch c.Mode {
	case consoleMode:
		setupWithConsole(c)
		return nil
	case volumeMode:
		return setupWithVolume(c)
	default:
		return setupWithFile(c)
	}
}

func Close() error {
	if writeConsole {
		return nil
	}

	if atomic.LoadUint32(&initialized) == 0 {
		return ErrLogNotInitialized
	}

	atomic.StoreUint32(&initialized, 0)

	loggers := []io.WriteCloser{infoLogger, errorLogger, fatalLogger, slowLogger, statLogger}
	for _, logger := range loggers {
		if logger != nil {
			if err := infoLogger.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func Disable() {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)

		infoLogger = iox.NopCloser(ioutil.Discard)
		errorLogger = iox.NopCloser(ioutil.Discard)
		fatalLogger = iox.NopCloser(ioutil.Discard)
		slowLogger = iox.NopCloser(ioutil.Discard)
		statLogger = iox.NopCloser(ioutil.Discard)
		stackLogger = ioutil.Discard
	})
}

func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

func WithKeepDays(days int) LogOption {
	return func(opts *logOptions) {
		opts.keepDays = days
	}
}

func WithGzip() LogOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

func WithCooldownMillis(millis int) LogOption {
	return func(opts *logOptions) {
		opts.logStackCooldownMills = millis
	}
}

func Info(v ...interface{}) {
	syncInfo(fmt.Sprint(v...))
}

func Infof(format string, args ...interface{}) {
	syncInfo(fmt.Sprintf(format, args...))
}

func Error(v ...interface{}) {
	ErrorCaller(1, v...)
}

func Errorf(format string, args ...interface{}) {
	ErrorCallerf(1, format, args...)
}

func ErrorCaller(callDepth int, v ...interface{}) {
	syncError(fmt.Sprint(v...), callDepth+callerInnerDepth)
}

func ErrorCallerf(callDepth int, format string, args ...interface{}) {
	syncError(fmt.Sprintf(format, args...), callDepth+callerInnerDepth)
}

func ErrorStack(v ...interface{}) {
	syncStack(fmt.Sprint(v...))
}

func Fatal(v ...interface{}) {
	syncFatal(fmt.Sprint(v...))
}

func Fatalf(format string, args ...interface{}) {
	syncFatal(fmt.Sprintf(format, args...))
}

func Slow(v ...interface{}) {
	syncSlow(fmt.Sprint(v...))
}

func Slowf(format string, v ...interface{}) {
	syncSlow(fmt.Sprintf(format, v...))
}

func Stat(v ...interface{}) {
	syncStat(fmt.Sprint(v...))
}

func Statf(format string, v ...interface{}) {
	syncStat(fmt.Sprintf(format, v...))
}

func setupLogLevel(c LogConf) {
	switch c.Level {
	case infoLevel:
		SetLevel(InfoLevel)
	case slowLevel:
		SetLevel(SlowLevel)
	case errorLevel:
		SetLevel(ErrorLevel)
	case fatalLevel:
		SetLevel(FatalLevel)
	}
}

func shouldLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

func setupWithConsole(c LogConf) {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		writeConsole = true
		setupLogLevel(c)

		infoLogger = newLogWriter(log.New(os.Stdout, "", flags))
		errorLogger = newLogWriter(log.New(os.Stderr, "", flags))
		fatalLogger = newLogWriter(log.New(os.Stderr, "", flags))
		slowLogger = newLogWriter(log.New(os.Stderr, "", flags))
		stackLogger = NewLessWriter(errorLogger, options.logStackCooldownMills)
		statLogger = infoLogger
	})
}

func setupWithFile(c LogConf) error {
	var opts []LogOption
	var err error

	if len(c.Path) == 0 {
		return ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}

	accessFile := path.Join(c.Path, accessFilename)
	errorFile := path.Join(c.Path, errorFilename)
	fatalFile := path.Join(c.Path, fatalFilename)
	slowFile := path.Join(c.Path, slowFilename)
	statFile := path.Join(c.Path, statFilename)

	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		handleOptions(opts)
		setupLogLevel(c)

		if infoLogger, err = createOutput(accessFile); err != nil {
			return
		}
		if errorLogger, err = createOutput(errorFile); err != nil {
			return
		}
		if fatalLogger, err = createOutput(fatalFile); err != nil {
			return
		}
		if slowLogger, err = createOutput(slowFile); err != nil {
			return
		}
		if statLogger, err = createOutput(statFile); err != nil {
			return
		}

		stackLogger = NewLessWriter(errorLogger, options.logStackCooldownMills)
	})

	return err
}

func setupWithVolume(c LogConf) error {
	if len(c.ServiceName) == 0 {
		return ErrLogServiceNameNotSet
	}

	c.Path = path.Join(c.Path, c.ServiceName, sysx.Hostname())
	return setupWithFile(c)
}

func handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&options)
	}
}

func createOutput(filename string) (io.WriteCloser, error) {
	if len(filename) == 0 {
		return nil, ErrLogPathNotSet
	}

	return NewLogger(
		filename,
		DefaultRotateRule(filename, backupFileDelimiter, options.keepDays, options.gzipEnabled),
		options.gzipEnabled,
	)
}

func syncInfo(msg string) {
	if shouldLog(InfoLevel) {
		output(infoLogger, infoLevel, msg)
	}
}

func syncError(msg string, callDepth int) {
	if shouldLog(ErrorLevel) {
		outputError(errorLogger, msg, callDepth)
	}
}

func syncStack(msg string) {
	output(stackLogger, errorLevel, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
}

func syncFatal(msg string) {
	if shouldLog(FatalLevel) {
		output(fatalLogger, fatalLevel, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	}
}

func syncSlow(msg string) {
	if shouldLog(SlowLevel) {
		output(slowLogger, slowLevel, msg)
	}
}

func syncStat(msg string) {
	output(statLogger, statLevel, msg)
}

func outputError(writer io.WriteCloser, msg string, callDepth int) {
	content := formatWithCaller(msg, callDepth)
	output(writer, errorLevel, content)
}

func formatWithCaller(msg string, callDepth int) string {
	var b strings.Builder

	caller := getCaller(callDepth)
	if len(caller) > 0 {
		b.WriteString(caller)
		b.WriteByte(' ')
	}
	b.WriteString(msg)

	return b.String()
}

func getCaller(callDepth int) string {
	var b strings.Builder

	_, file, line, ok := runtime.Caller(callDepth)
	if ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		b.WriteString(short)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(line))
	}

	return b.String()
}

func output(writer io.Writer, level, msg string) {
	outputJson(writer, logEntry{
		Timestamp: getTimestamp(),
		Level:     level,
		Content:   msg,
	})
}

func outputJson(writer io.Writer, entry logEntry) {
	if content, err := json.Marshal(entry); err != nil {
		log.Println(err.Error())
	} else if atomic.LoadUint32(&initialized) == 0 || writer == nil {
		log.Println(string(content))
	} else {
		writer.Write(append(content, '\n'))
	}
}

func getTimestamp() string {
	return timex.Time().Format(timeFormat)
}
