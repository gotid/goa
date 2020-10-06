package mathx

import (
	"math/rand"
	"sync"
	"time"
)

// 维护可能性的概率
type Prob struct {
	r    *rand.Rand
	lock sync.Mutex
}

// 新建可能性概率
func NewProb() *Prob {
	return &Prob{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// 返回 v 值是否大于随机值
func (p *Prob) TrueOnProb(v float64) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	return v > p.r.Float64()
}
