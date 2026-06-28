package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Crash struct{}

type Child struct{}

func (c Child) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		fmt.Println("started")
	case Crash:
		fmt.Println("crashed")
		panic("CRASH!")
	}
}

type Parent struct{}

func (p Parent) Receive(ctx actors.Context, msg any) {
	if _, ok := msg.(actors.Started); ok {
		child := ctx.Spawn(func() actors.Actor { return Child{} })
		ctx.Send(child, Crash{})
	}
}

func (p Parent) Decide(reason any) actors.Directive {
	fmt.Println("Child crashed, reason:", reason)
	return actors.Restart
}

func main() {
	sys := actors.NewSystem()
	sys.Spawn(func() actors.Actor { return Parent{} })
	time.Sleep(1 * time.Second)
}
