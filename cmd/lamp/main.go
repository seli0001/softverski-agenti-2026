package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Toggle struct{}

type Lamp struct{}

func (l Lamp) Receive(ctx actors.Context, msg any) {
	if _, ok := msg.(actors.Started); ok {
		ctx.Become(offBehavior)
	}
}

func offBehavior(ctx actors.Context, msg any) {
	if _, ok := msg.(Toggle); ok {
		fmt.Println("lamp: turning on")
		ctx.Become(onBehavior)
	}
}

func onBehavior(ctx actors.Context, msg any) {
	if _, ok := msg.(Toggle); ok {
		fmt.Println("lamp: turning off")
		ctx.Become(offBehavior)
	}
}

func main() {
	sys := actors.NewSystem()
	lamp := sys.Spawn(Lamp{})

	sys.Send(lamp, Toggle{})
	sys.Send(lamp, Toggle{})
	sys.Send(lamp, Toggle{})
	time.Sleep(1 * time.Second)
}
