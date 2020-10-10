package cache

import (
	"github.com/z-sdk/goa/lib/logx"
	"sync/atomic"
	"time"
)

const statInterval = time.Minute // 缓存统计周期

// Stat 缓存统计
type Stat struct {
	name    string
	Total   uint64 // 请求数
	Hit     uint64 // 命中数
	Miss    uint64 // 错过数
	DbFails uint64 // 查库失败次数
}

func NewCacheStat(name string) *Stat {
	stat := &Stat{name: name}
	go stat.Loop()
	return stat
}

func (s *Stat) Loop() {
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			total := atomic.SwapUint64(&s.Total, 0)
			if total == 0 {
				continue
			}

			hit := atomic.SwapUint64(&s.Hit, 0)
			percent := 100 * float32(hit) / float32(total)
			miss := atomic.SwapUint64(&s.Miss, 0)
			dbf := atomic.SwapUint64(&s.DbFails, 0)
			logx.Statf("dbcache(%s) - qpm: %d, hit_ratio: %.1f%%, hit: %d, miss: %d, db_fails: %d",
				s.name, total, percent, hit, miss, dbf)
		}
	}
}

func (s *Stat) IncrTotal() {
	atomic.AddUint64(&s.Total, 1)
}

func (s *Stat) IncrHit() {
	atomic.AddUint64(&s.Hit, 1)
}

func (s *Stat) IncrMiss() {
	atomic.AddUint64(&s.Miss, 1)
}

func (s *Stat) IncrDbFails() {
	atomic.AddUint64(&s.DbFails, 1)
}
