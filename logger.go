package sched

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type noopLogger struct {
}

func (n *noopLogger) Debugf(format string, args ...interface{}) {
}

func (n *noopLogger) Infof(format string, args ...interface{}) {
}

func (n *noopLogger) Warnf(format string, args ...interface{}) {
}

func (n *noopLogger) Errorf(format string, args ...interface{}) {
}
