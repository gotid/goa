package logx

import "log"

// 日志书写器
type logWriter struct {
	logger *log.Logger
}

func newLogWriter(logger *log.Logger) logWriter {
	return logWriter{
		logger: logger,
	}
}

func (w logWriter) Close() error {
	return nil
}

func (w logWriter) Write(data []byte) (int, error) {
	w.logger.Print(string(data))
	return len(data), nil
}
