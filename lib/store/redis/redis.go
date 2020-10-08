package redis

import (
	"github.com/go-redis/redis"
	"goa/lib/breaker"
	"time"
)

const (
	ClusterMode    = "cluster"
	StandaloneMode = "standalone"

	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
	slowThreshold   = 100 * time.Millisecond
)

type (
	Redis struct {
		Addr     string
		Mode     string
		Password string
		brk      breaker.Breaker
	}

	Client interface {
		redis.Cmdable
	}

	Pipeliner = redis.Pipeliner

	Pair struct {
		Key   string
		Score int64
	}
)

func NewRedis(addr, mode, password string) *Redis {
	return &Redis{
		Addr:     addr,
		Mode:     mode,
		Password: password,
		brk:      breaker.NewBreaker(),
	}
}
