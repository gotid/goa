package logx

import (
	"fmt"
	"goa/lib/timex"
	"io"
	"time"
)

type durationLogger logEntry

func WithDuration(d time.Duration) Logger {
	return &durationLogger{
		Duration: timex.MillisecondDuration(d),
	}
}

func (l *durationLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLogger, infoLevel, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLogger, infoLevel, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLogger, errorLevel, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLogger, errorLevel, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Slow(v ...interface{}) {
	if shouldLog(SlowLevel) {
		l.write(slowLogger, slowLevel, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(SlowLevel) {
		l.write(slowLogger, slowLevel, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) WithDuration(d time.Duration) Logger {
	l.Duration = timex.MillisecondDuration(d)
	return l
}

func (l *durationLogger) write(writer io.Writer, level, content string) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = content
	outputJson(writer, logEntry(*l))
}
