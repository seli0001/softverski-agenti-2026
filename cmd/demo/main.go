package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Ball struct {
	From    actors.PID
	counter int
}

type Watcher struct {
	target actors.PID
}

func (w Watcher) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		ctx.Watch(w.target)
	case actors.Terminated:
		fmt.Println("actor terminated", m.PID)
	}
}

type Player struct {
	name string
}

func (p Player) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case Ball:
		fmt.Println("Ball", msg.(Ball))
		if m.counter > 0 {
			ctx.Send(m.From, Ball{From: ctx.Self(), counter: m.counter - 1})
		}
	case actors.Started:
		fmt.Println(p.name + " started")
	case actors.Stopping:
		fmt.Println(p.name + " stopping")
	}
}

func main() {
	sys := actors.NewSystem()
	ping := sys.Spawn(func() actors.Actor {
		return Player{
			name: "ping",
		}
	})
	pong := sys.Spawn(func() actors.Actor {
		return Player{
			name: "pong",
		}
	})

	sys.Spawn(func() actors.Actor {
		return Watcher{
			target: ping,
		}
	})

	sys.Send(ping, Ball{From: pong, counter: 6})
	time.Sleep(1 * time.Second)
	sys.Stop(ping)
	time.Sleep(1 * time.Second)

}
