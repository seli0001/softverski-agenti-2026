package main

import (
	"agenti/actors"
	"agenti/cluster"
	"agenti/crdt"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

type AntiEntropy struct {
	counter     *crdt.GCounter
	timeToStop  int
	selfAddress string
}

type SyncTick struct{}

type SyncState struct {
	State map[string]int
}

func (a *AntiEntropy) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		a.timeToStop = 10
		a.selfAddress = ctx.Self().Address
		a.counter = crdt.NewGCounter(a.selfAddress)
		ctx.SendLater(ctx.Self(), SyncTick{}, 3*time.Second)
	case SyncTick:
		if a.timeToStop > 0 {
			a.counter.Increment()
			a.timeToStop--
		}
		ctx.Send(actors.PID{Address: a.selfAddress, Id: "membership"}, cluster.GetPeers{ReplyTo: ctx.Self()})
		ctx.SendLater(ctx.Self(), SyncTick{}, 3*time.Second)
	case cluster.Peers:
		for _, address := range m.Addresses {
			ctx.Send(actors.PID{Address: address, Id: "counter"}, SyncState{
				State: a.counter.GetState(),
			})
		}
	case SyncState:
		a.counter.MergeMap(m.State)
		fmt.Println("value:", a.counter.Value())
	}
}

func main() {
	myAddress := os.Args[1]
	var contactAddress string
	if len(os.Args) > 2 {
		contactAddress = os.Args[2]
	} else {
		contactAddress = ""
	}
	gob.Register(cluster.Join{})
	gob.Register(cluster.PeerList{})
	gob.Register(cluster.ForwardJoin{})
	gob.Register(cluster.Ping{})
	gob.Register(cluster.Pong{})
	gob.Register(cluster.NeighborRequest{})
	gob.Register(cluster.NeighborResponse{})
	gob.Register(cluster.Disconnect{})
	gob.Register(SyncState{})

	sys := actors.NewRemoteSystem(myAddress)

	sys.SpawnNamed(func() actors.Actor {
		return cluster.NewMembership(contactAddress)
	}, "membership")
	sys.SpawnNamed(func() actors.Actor {
		return &AntiEntropy{}
	}, "counter")

	select {}
}
