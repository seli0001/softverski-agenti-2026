package main

import (
	"agenti/actors"
	"agenti/cluster"
	"agenti/crdt"
	"agenti/fl"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"time"
)

type SyncTick struct{}

type ModelGossip struct {
	Weights []float64
	N       int
	Counter map[string]int
}

type Trainer struct {
	round       int
	weights     []float64
	counter     *crdt.GCounter
	x           [][]float64
	y           []float64
	testX       [][]float64
	testY       []float64
	selfAddress string
	index       int
}

func NewTrainer(index int) *Trainer {
	return &Trainer{index: index}
}

func (t *Trainer) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		t.selfAddress = ctx.Self().Address
		X, y := fl.Load("data/airlines.csv")
		t.x, t.y = fl.Slice(X, y, t.index, 3000)
		t.testX, t.testY = fl.TestSet(X, y, 50000)
		t.weights = make([]float64, 10)
		t.counter = crdt.NewGCounter(t.selfAddress)
		ctx.SendLater(ctx.Self(), SyncTick{}, 2*time.Second)
	case SyncTick:
		w := append([]float64{}, t.weights...)
		t.weights = fl.Train(t.x, t.y, w, 1, 0.01)
		t.counter.Increment()
		ctx.Send(
			actors.PID{
				Address: ctx.Self().Address,
				Id:      "membership",
			},
			cluster.GetPeers{
				ReplyTo: ctx.Self(),
			},
		)
		ctx.SendLater(ctx.Self(), SyncTick{}, 3*time.Second)

	case cluster.Peers:
		for _, peer := range m.Addresses {
			ctx.Send(actors.PID{Address: peer, Id: "trainer"}, ModelGossip{
				Weights: append([]float64{}, t.weights...),
				N:       len(t.x),
				Counter: t.counter.GetState(),
			})
		}
	case ModelGossip:
		total := float64(len(t.x) + m.N)
		fresh := make([]float64, len(t.weights))
		for j := 0; j < len(fresh); j++ {
			fresh[j] = (t.weights[j]*float64(len(t.x)) + m.Weights[j]*float64(m.N)) / total
		}
		t.weights = fresh
		t.counter.MergeMap(m.Counter)
		fmt.Println("acc:", fl.Accuracy(t.testX, t.testY, t.weights), "| treninga u klasteru:", t.counter.Value())
	}
}

func main() {
	myAddress := os.Args[1]
	index, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	var contactAddress string
	if len(os.Args) > 3 {
		contactAddress = os.Args[3]
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
	gob.Register(ModelGossip{})

	sys := actors.NewRemoteSystem(myAddress)

	sys.SpawnNamed(func() actors.Actor {
		return cluster.NewMembership(contactAddress)
	}, "membership")

	sys.SpawnNamed(func() actors.Actor {
		return NewTrainer(index)
	}, "trainer")

	select {}
}
