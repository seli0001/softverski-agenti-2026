package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Parent struct{}

func (p Parent) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		ctx.Spawn(func() actors.Actor { return Child{name: "seli1"} })
		ctx.Spawn(func() actors.Actor { return Child{name: "seli2"} })
		fmt.Println("parent started -> new child")
	case actors.Stopping:
		fmt.Println("parent stopped")
	}
}

type Child struct {
	name string
}

func (c Child) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		fmt.Println(c.name, "is started")
	case actors.Stopping:
		fmt.Println(c.name, " stopped")
	}
}

func main() {
	sys := actors.NewSystem()
	pid := sys.Spawn(func() actors.Actor { return Parent{} })
	time.Sleep(1 * time.Second)
	sys.Stop(pid)
	time.Sleep(1 * time.Second)
}
