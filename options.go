package sched

import (
	"go.uber.org/zap"
)

type Option func(s *sched)

func WithLogger(log *zap.Logger) Option {
	return func(s *sched) {
		s.logger = log
	}
}
