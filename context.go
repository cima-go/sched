package sched

type Context interface {
	Cancel() error
}

type ctx struct {
	s Scheduler
	j *Job
}

func (c *ctx) Cancel() error {
	return c.s.Cancel(c.j)
}
