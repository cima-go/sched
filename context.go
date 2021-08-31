package sched

type Context interface {
	Task() *Task
	Cancel() error
}

type ctx struct {
	s Scheduler
	t *Task
	// states
	cancelled bool
}

func (c *ctx) Task() *Task {
	return c.t
}

func (c *ctx) Cancel() error {
	c.cancelled = true
	return c.s.Cancel(c.t.Job)
}
