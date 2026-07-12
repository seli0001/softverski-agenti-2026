package main

import (
	"agenti/actors"
	"agenti/cluster"
	"encoding/gob"
	"os"
)

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

	sys := actors.NewRemoteSystem(myAddress)

	sys.SpawnNamed(func() actors.Actor {
		return cluster.NewMembership(contactAddress)
	}, "membership")

	select {}
}
