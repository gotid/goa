package stringx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	cases := []struct {
		input   string
		ignores []rune
		expect  string
	}{
		{``, nil, ``},
		{`abcd`, nil, `abcd`},
		{`ab,cd,ef`, []rune{','}, `abcdef`},
		{`ab, cd,ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef `, []rune{',', ' '}, `abcdef`},
	}

	for _, each := range cases {
		t.Run(each.input, func(t *testing.T) {
			actual := Filter(each.input, func(r rune) bool {
				for _, ignore := range each.ignores {
					if ignore == r {
						return true
					}
				}
				return false
			})
			assert.Equal(t, each.expect, actual)
		})
	}
}

func TestA(t *testing.T) {
	fields := strings.FieldsFunc("ab, cd, ef ", func(r rune) bool {
		return r == ',' || r == ' '
	})
	fmt.Println(fields)
}
