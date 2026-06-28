package main

import (
	"agenti/actors"
	"fmt"
	"time"
)

type Crash struct{}

type Error struct{}

func (e Error) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case Crash:
		panic("crashing!")
	case actors.Started:
		fmt.Println("actors started")
	case string:
		fmt.Println("primio:", msg)
	}
}

func main() {
	sys := actors.NewSystem()
	e := sys.Spawn(func() actors.Actor { return Error{} })

	sys.Send(e, Crash{})
	sys.Send(e, "still working")

	time.Sleep(1 * time.Second)
}
