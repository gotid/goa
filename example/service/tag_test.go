package service

import (
	"fmt"
	"testing"
)

func TestNewServiceContext(t *testing.T) {
	ids := []int64{1, 2, 3}
	names := GetNames(ids)
	fmt.Println(names)
}
