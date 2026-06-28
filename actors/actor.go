package actors

type Actor interface {
	Receive(ctx Context, msg any)
}

type Context interface {
	Self() PID
	Send(to PID, msg any)
	Watch(pid PID)
	Become(behavior func(Context, any))
}

type actorContext struct {
	self PID
	sys  *System
}

func (c actorContext) Self() PID {
	return c.self
}

func (c actorContext) Send(to PID, msg any) {
	c.sys.Send(to, msg)
}

func (c actorContext) Watch(pid PID) {
	c.sys.Watch(c.self, pid)
}

func (c actorContext) Become(behavior func(Context, any)) {
	c.sys.Become(c.self, behavior)
}

type Started struct{}

type Stopping struct{}

type Terminated struct {
	PID PID
}
