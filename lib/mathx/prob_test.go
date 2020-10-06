package mathx

import (
	"fmt"
	"testing"
)

func TestProb_TrueOnProb(t *testing.T) {
	prob := NewProb()
	fmt.Println(prob.TrueOnProb(0.3))
}
