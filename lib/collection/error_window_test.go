package collection

import (
	"fmt"
	"testing"
)

func TestNewErrorWindow(t *testing.T) {
	errWin := NewErrorWindow()

	errWin.Add("因为1....")
	errWin.Add("因为2....")
	errWin.Add("因为3....")

	fmt.Println(errWin.count)

	for _, reason := range errWin.reasons {
		fmt.Println(reason)
	}

	fmt.Println(errWin.String())
}
