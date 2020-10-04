package fs

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
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
