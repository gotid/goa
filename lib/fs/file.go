package fs

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func CloseOnExec(file *os.File) {
	if file != nil {
		syscall.CloseOnExec(int(file.Fd()))
	}
}

func GzipFile(filename string) error {
	src, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(fmt.Sprintf("%s.gz", filename))
	if err != nil {
		return err
	}
	defer out.Close()

	dst := gzip.NewWriter(out)
	if _, err = io.Copy(dst, src); err != nil {
		return err
	} else if err = dst.Close(); err != nil {
		return err
	}

	return os.Remove(filename)
}

func MkdirIfNotExist(dir string) error {
	if len(dir) == 0 {
		return nil
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}

	return nil
}

func CreateIfNotExist(name string) (*os.File, error) {
	if FileExist(name) {
		return nil, fmt.Errorf("%s 文件已存在", name)
	}

	return os.Create(name)
}

func RemoveIfExists(name string) error {
	if !FileExist(name) {
		return nil
	}

	return os.Remove(name)
}

func RemoveOrQuit(name string) error {
	if !FileExist(name) {
		return nil
	}

	// 接收控制台指令
	fmt.Printf("%s 已存在，是否覆盖？\n回车覆盖，Ctrl-C取消...", aurora.BgRed(aurora.Bold(name)))
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	return os.Remove(name)
}

func FileExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func FilenameWithoutExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}
