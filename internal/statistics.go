package internal

import (
	"time"
)

type statistics struct {
	name       string
	interval   time.Duration
	startTime  time.Time
	lastTime   time.Time
	lastCount  int64
	totalCount int64
}

func newStatistics(name string, interval time.Duration) *statistics {
	return &statistics{
		name:      name,
		interval:  interval,
		startTime: time.Now(),
		lastTime:  time.Now(),
	}
}

func (s *statistics) update() {
	s.lastCount++
	now := time.Now()
	if now.Sub(s.lastTime) >= time.Second {
		s.totalCount += s.lastCount
		sub := now.Sub(s.startTime)
		log.Debugf("%s: %s | %d/%d | %f", s.name, sub, s.lastCount, s.totalCount,
			float64(s.totalCount)/sub.Seconds())
		s.lastTime = now
		s.totalCount = 0
	}
}
