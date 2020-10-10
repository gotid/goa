package main

import (
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	breakRange = 20
	workRange  = 50
)

type metric struct {
	calls int64
}

func (m *metric) addCall() {
	atomic.AddInt64(&m.calls, 1)
}

type server struct {
	state int32
}

func (s *server) start() {
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var state int32
		for {
			var v int32
			if state == 0 {
				v = r.Int31n(breakRange)
			} else {
				v = r.Int31n(workRange)
			}
			time.Sleep(time.Second * time.Duration(v+1))
			state ^= 1
			atomic.StoreInt32(&s.state, state)
		}
	}()
}

func (s *server) serve(m *metric) bool {
	m.addCall()
	return atomic.LoadInt32(&s.state) == 1
}

func newServer() *server {
	return &server{}
}

func main() {

}
