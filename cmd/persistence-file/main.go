package main

import (
	"agenti/actors"
	"agenti/persistence"
	"encoding/gob"
	"fmt"
	"time"
)

type Increment struct{}

type Incremented struct{}

type Counter struct {
	id       string
	journal  persistence.Journal
	snapshot persistence.SnapshotStore
	seq      int
	count    int
}

func (c *Counter) Receive(ctx actors.Context, msg any) {
	switch msg.(type) {
	case actors.Started:
		if state, seq, ok := c.snapshot.Load(c.id); ok {
			c.count = state.(int)
			c.seq = seq
		}
		events := c.journal.Events(c.id)
		replayed := 0
		for i := c.seq; i < len(events); i++ {
			c.apply(events[i])
			c.seq++
			replayed++
		}
		fmt.Println("Oporavljen:", c.count)
		fmt.Printf("snapshot: %d, od dogadjaja %d do %d dogadjaja", c.seq-replayed, replayed, len(events))
	case Increment:
		e := Incremented{}
		c.journal.Append(c.id, e)
		c.apply(e)
		c.seq++
		if c.seq%3 == 0 {
			c.snapshot.Save(c.id, c.seq, c.count)
			fmt.Println("snapshot ", c.seq, "-", c.count)
		}
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
	gob.Register(Incremented{})
	gob.Register(int(0))
	j := persistence.NewFileJournal("./data")
	ss := persistence.NewFileSnapshotStore("./data")
	sys := actors.NewSystem()
	c1 := sys.Spawn(func() actors.Actor {
		return &Counter{
			id:       "counter",
			journal:  j,
			snapshot: ss,
		}
	})

	for i := 0; i < 11; i++ {
		sys.Send(c1, Increment{})
	}
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
