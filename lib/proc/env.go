package proc

import (
	"os"
	"strconv"
	"sync"
)

var (
	envs    = make(map[string]string)
	envLock sync.Mutex
)

func Env(key string) string {
	envLock.Lock()
	val, ok := envs[key]
	envLock.Unlock()

	if ok {
		return val
	}

	val = os.Getenv(key)
	envLock.Lock()
	envs[key] = val
	envLock.Unlock()

	return val
}

func EnvInt(key string) (int, bool) {
	val := Env(key)
	if len(val) == 0 {
		return 0, false
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return n, true
}
