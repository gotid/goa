package timex

import (
	"fmt"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	fmt.Println(Now() - Now())
	fmt.Println(time.Since(time.Now()))

	fmt.Println(Since(Now()))

	fmt.Println(Time())
}
