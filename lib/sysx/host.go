package sysx

import (
	"github.com/z-sdk/goa/lib/stringx"
	"os"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = stringx.RandId()
	}
}

func Hostname() string {
	return hostname
}
