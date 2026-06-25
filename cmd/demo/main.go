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
	ping := sys.Spawn(Player{
		name: "ping",
	})
	pong := sys.Spawn(Player{
		name: "pong",
	})

	sys.Send(ping, Ball{From: pong, counter: 6})
	time.Sleep(1 * time.Second)
	sys.Stop(ping)
	time.Sleep(1 * time.Second)

}
