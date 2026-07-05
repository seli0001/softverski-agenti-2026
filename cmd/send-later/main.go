package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Tick struct{}

type SendLaterActor struct {
	counter int
}

func (s *SendLaterActor) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		ctx.SendLater(ctx.Self(), Tick{}, 2*time.Second)
	case Tick:
		s.counter++
		fmt.Println("nova poruka", s.counter)
		ctx.SendLater(ctx.Self(), Tick{}, 2*time.Second)
	}
}

func main() {
	sys := actors.NewSystem()
	sys.Spawn(func() actors.Actor {
		return &SendLaterActor{}
	})
	select {}
}
