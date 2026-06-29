package main

import (
	"agenti/actors"
	"agenti/persistence"
	"fmt"
	"time"
)

type Increment struct{}

type Incremented struct{}

type Counter struct {
	id      string
	journal persistence.Journal
	count   int
}

func (c *Counter) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		for _, e := range c.journal.Events(c.id) {
			c.apply(e)
		}
		fmt.Println("Oporavljen:", c.count)
	case Increment:
		e := Incremented{}
		c.journal.Append(c.id, e)
		c.apply(e)
		fmt.Println("count: ", c.count)
	}
}

func (c *Counter) apply(event any) {
	switch event.(type) {
	case Incremented:
		c.count++
	}
}

func main() {
	j := persistence.NewInMemoryJournal()
	sys := actors.NewSystem()
	c1 := sys.Spawn(func() actors.Actor {
		return &Counter{
			id:      "counter",
			journal: j,
		}
	})

	sys.Send(c1, Increment{})
	sys.Send(c1, Increment{})
	sys.Send(c1, Increment{})
	time.Sleep(1 * time.Second)
	sys.Stop(c1)
	time.Sleep(1 * time.Second)

	sys.Spawn(func() actors.Actor {
		return &Counter{
			id:      "counter",
			journal: j,
		}
	})
	time.Sleep(1 * time.Second)

}
