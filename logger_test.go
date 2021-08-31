package sched_test

import (
	"log"
)

type stdoutLogger struct {
}

func (s *stdoutLogger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func (s *stdoutLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)

}

func (s *stdoutLogger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)

}

func (s *stdoutLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
