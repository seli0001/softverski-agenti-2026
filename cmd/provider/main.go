package main

import (
	"agenti/actors"
	"agenti/cluster"
	"encoding/gob"
	"os"
)

func main() {
	myAddress := os.Args[1]

	gob.Register(cluster.Register{})
	gob.Register(cluster.Heartbeat{})
	gob.Register(cluster.Members{})

	sys := actors.NewRemoteSystem(myAddress)
	if len(os.Args) > 2 {
		sys.SpawnNamed(func() actors.Actor { return cluster.NewClient(os.Args[2]) }, "member")
	} else {
		sys.SpawnNamed(func() actors.Actor {
			return cluster.NewProvider()
		}, "provider")
	}

	select {}
}
