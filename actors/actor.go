package actors

import "time"

type Actor interface {
	Receive(ctx Context, msg any)
}

type Context interface {
	Self() PID
	Send(to PID, msg any)
	SendLater(pid PID, msg any, delay time.Duration)
	Watch(pid PID)
	Become(behavior func(Context, any))
	Spawn(producer func() Actor) PID
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

func (c actorContext) Spawn(producer func() Actor) PID {
	return c.sys.SpawnChildren(c.self, producer)
}

func (c actorContext) SendLater(pid PID, msg any, delay time.Duration) {
	c.sys.SendLater(pid, msg, delay)
}

type Started struct{}

type Stopping struct{}

type Terminated struct {
	PID PID
}
