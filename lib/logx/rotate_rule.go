package logx

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	dateFormat  = "2000-01-01"
	hoursPerDay = 24
)

type (
	RotateRule interface {
		BackupFilename() string
		MarkRotated()
		OutdatedFiles() []string
		ShallRotate() bool
	}

	DailyRotateRule struct {
		rotatedTime string
		filename    string
		delimiter   string
		days        int
		gzip        bool
	}
)

func DefaultRotateRule(filename, delimiter string, days int, gzip bool) RotateRule {
	return &DailyRotateRule{
		rotatedTime: getNowDate(),
		filename:    filename,
		delimiter:   delimiter,
		days:        days,
		gzip:        gzip,
	}
}

func (r DailyRotateRule) BackupFilename() string {
	return fmt.Sprintf("%s%s%s", r.filename, r.delimiter, getNowDate())
}

func (r DailyRotateRule) MarkRotated() {
	r.rotatedTime = getNowDate()
}

func (r DailyRotateRule) OutdatedFiles() []string {
	if r.days <= 0 {
		return nil
	}

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s*.gz", r.filename, r.delimiter)
	} else {
		pattern = fmt.Sprintf("%s%s*", r.filename, r.delimiter)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("获取过期日志文件失败：%s", err)
		return nil
	}

	var b strings.Builder
	boundary := time.Now().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(dateFormat)
	fmt.Fprintf(&b, "%s%s%s", r.filename, r.delimiter, boundary)
	if r.gzip {
		b.WriteString(".gz")
	}
	boundaryFile := b.String()

	var outdates []string
	for _, file := range files {
		// 对比文件名，判断是否为过期文件
		if file < boundaryFile {
			outdates = append(outdates, file)
		}
	}

	return outdates
}

func (r DailyRotateRule) ShallRotate() bool {
	return len(r.rotatedTime) > 0 && getNowDate() != r.rotatedTime
}

func getNowDate() string {
	return time.Now().Format(dateFormat)
}
