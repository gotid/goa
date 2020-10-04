package logx

import (
	"errors"
	"goa/lib/fs"
	"goa/lib/lang"
	"goa/lib/timex"
	"log"
	"os"
	"path"
	"sync"
)

const (
	bufferSize     = 100
	defaultDirMode = 0755
)

var ErrLogFileClosed = errors.New("错误：日志文件已关闭")

type (
	RotateLogger struct {
		filename  string
		backup    string
		rule      RotateRule
		compress  bool
		keepDays  int
		fp        *os.File
		channel   chan []byte
		done      chan lang.PlaceholderType
		wg        sync.WaitGroup
		closeOnce sync.Once
	}
)

func NewLogger(filename string, rule RotateRule, compress bool) (*RotateLogger, error) {
	l := &RotateLogger{
		filename: filename,
		rule:     rule,
		compress: compress,
		channel:  make(chan []byte, bufferSize),
		done:     make(chan lang.PlaceholderType),
	}

	if err := l.init(); err != nil {
		return nil, err
	}

	l.startWorker()

	return l, nil
}

func (l *RotateLogger) Write(data []byte) (n int, err error) {
	select {
	case l.channel <- data:
		return len(data), nil
	case <-l.done:
		log.Println(string(data))
		return 0, ErrLogFileClosed
	}
}

func (l *RotateLogger) Close() (err error) {
	l.closeOnce.Do(func() {
		close(l.done)
		l.wg.Wait()

		if err = l.fp.Sync(); err != nil {
			return
		}

		err = l.fp.Close()
	})

	return
}

func (l *RotateLogger) init() error {
	l.backup = l.rule.BackupFilename()

	if _, err := os.Stat(l.filename); err != nil {
		basePath := path.Dir(l.filename)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, defaultDirMode); err != nil {
				return err
			}
		}

		if l.fp, err = os.Create(l.filename); err != nil {
			return err
		}
	} else if l.fp, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, defaultDirMode); err != nil {
		return err
	}

	fs.CloseOnExec(l.fp)

	return nil
}

func (l *RotateLogger) rotate() error {
	if l.fp != nil {
		err := l.fp.Close()
		l.fp = nil
		if err != nil {
			return err
		}
	}

	_, err := os.Stat(l.filename)
	if err == nil && len(l.backup) > 0 {
		backupFilename := l.getBackupFilename()
		err = os.Rename(l.filename, backupFilename)
		if err != nil {
			return err
		}

		l.postRotate(backupFilename)
	}

	l.backup = l.rule.BackupFilename()
	if l.fp, err = os.Create(l.filename); err != nil {
		fs.CloseOnExec(l.fp)
	}

	return err
}

func (l *RotateLogger) write(v []byte) {
	if l.rule.ShallRotate() {
		if err := l.rotate(); err != nil {
			log.Println(err)
		} else {
			l.rule.MarkRotated()
		}
	}

	if l.fp != nil {
		l.fp.Write(v)
	}
}

func (l *RotateLogger) startWorker() {
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		for {
			select {
			case event := <-l.channel:
				l.write(event)
			case <-l.done:
				return
			}
		}
	}()
}

func (l *RotateLogger) getBackupFilename() string {
	if len(l.backup) == 0 {
		return l.rule.BackupFilename()
	} else {
		return l.backup
	}
}

func (l *RotateLogger) postRotate(filename string) {
	go func() {
		// 此处不能使用 threading.GoSafe，因为logx 和 threading 会循环引用
		l.maybeCompressFile(filename)
		l.maybeDeleteOutdatedFiles()
	}()
}

func (l *RotateLogger) maybeCompressFile(filename string) {
	if !l.compress {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			ErrorStack(r)
		}
	}()

	compressLogFile(filename)
}

func (l *RotateLogger) maybeDeleteOutdatedFiles() {
	files := l.rule.OutdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Errorf("删除过期日志失败: %s", file)
		}
	}
}

func compressLogFile(filename string) {
	start := timex.Now()
	Infof("压缩日志文件: %s", filename)
	if err := fs.GzipFile(filename); err != nil {
		Errorf("压缩失败: %s", err)
	} else {
		Infof("压缩日志文件: %s, 耗时 %s", filename, timex.Since(start))
	}
}
