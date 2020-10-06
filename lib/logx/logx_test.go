package logx

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	c := LogConf{
		Mode: "console",
		Path: "logs",
	}
	MustSetup(c)
	defer Close()

	Info("info")
	Error("error")
	ErrorStack("hello")
	Errorf("%s and %s", "hello", "world")
	Fatalf("%s fatal %s", "hello", "world")
	Slowf("%s slow %s", "hello", "world")
	Statf("%s stat %s", "hello", "world")
	WithDuration(time.Minute + time.Second).Info("hello")
	WithDuration(time.Minute + time.Second).Errorf("hello")
}
