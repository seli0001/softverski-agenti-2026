package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Parent struct{}

func (p Parent) Receive(ctx actors.Context, msg any) {
	if _, ok := msg.(actors.Started); ok {
		ctx.Spawn(func() actors.Actor { return Child{name: "seli1"} })
		ctx.Spawn(func() actors.Actor { return Child{name: "seli2"} })
		fmt.Println("parent started -> new child")
	}
}

type Child struct {
	name string
}

func (c Child) Receive(ctx actors.Context, msg any) {
	if _, ok := msg.(actors.Started); ok {
		fmt.Println(c.name, "is started")
	}
}

func main() {
	sys := actors.NewSystem()
	sys.Spawn(func() actors.Actor { return Parent{} })
	time.Sleep(1 * time.Second)
}
