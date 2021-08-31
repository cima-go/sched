package sched

import (
	"time"
)

type Option func(s *sched)

func WithLogger(log Logger) Option {
	return func(s *sched) {
		s.logger = log
	}
}

func WithTicker(dur time.Duration) Option {
	return func(s *sched) {
		s.ticker = dur
	}
}
