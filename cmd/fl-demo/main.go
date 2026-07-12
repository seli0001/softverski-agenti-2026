package main

import (
	"agenti/actors"
	"agenti/fl"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"time"
)

type GlobalModel struct {
	Round   int
	Weights []float64
}
type LocalUpdate struct {
	Round   int
	Weights []float64
	N       int
	From    string
}

type RegisterTrainer struct{ Address string }

type Trainer struct {
	aggregatorAddress string
	index             int
	X                 [][]float64
	y                 []float64
	selfAddress       string
}

func NewTrainer(aggregatorAddress string, index int) *Trainer {
	return &Trainer{aggregatorAddress: aggregatorAddress, index: index}
}

func (t *Trainer) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		t.selfAddress = ctx.Self().Address
		X, y := fl.Load("data/airlines.csv")
		t.X, t.y = fl.Slice(X, y, t.index, 3000)
		ctx.Send(actors.PID{
			Address: t.aggregatorAddress,
			Id:      "aggregator",
		}, RegisterTrainer{Address: t.selfAddress})
	case GlobalModel:
		w := fl.Train(t.X, t.y, m.Weights, 3, 0.01)
		ctx.Send(actors.PID{
			Address: t.aggregatorAddress,
			Id:      "aggregator",
		}, LocalUpdate{
			Round:   m.Round,
			Weights: w,
			N:       len(t.X),
			From:    t.selfAddress,
		})
		fmt.Println("runda ", m.Round, ":", m.Weights)
	}
}

type Aggregator struct {
	weights  []float64
	round    int
	trainers map[string]bool
	updates  map[string]LocalUpdate
	textX    [][]float64
	testY    []float64
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		trainers: make(map[string]bool),
		updates:  make(map[string]LocalUpdate),
		weights:  make([]float64, 10),
	}
}

type StartRound struct {
}

func (a *Aggregator) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		X, y := fl.Load("data/airlines.csv")
		a.textX, a.testY = fl.TestSet(X, y, 50000)
		ctx.SendLater(ctx.Self(), StartRound{}, 5*time.Second)
	case RegisterTrainer:
		a.trainers[m.Address] = true
		fmt.Println("trainer:", m.Address)
	case StartRound:
		for tr := range a.trainers {
			ctx.Send(
				actors.PID{Address: tr, Id: "trainer"},
				GlobalModel{Round: a.round, Weights: a.weights},
			)
		}
	case LocalUpdate:
		if m.Round != a.round {
			return
		}
		a.updates[m.From] = m
		if len(a.updates) < len(a.trainers) {
			return
		}
		total := 0.0
		for _, u := range a.updates {
			total += float64(u.N)
		}
		for j := range a.weights {
			s := 0.0
			for _, u := range a.updates {
				s += u.Weights[j] * float64(u.N)
			}
			a.weights[j] = s / float64(total)
		}
		fmt.Println("runda", a.round, "acc:", fl.Accuracy(a.textX, a.testY, a.weights))
		a.updates = make(map[string]LocalUpdate)
		a.round++
		if a.round < 10 {
			ctx.Send(ctx.Self(), StartRound{})
		}
	}
}

func main() {
	myAddress := os.Args[1]

	gob.Register(GlobalModel{})
	gob.Register(RegisterTrainer{})
	gob.Register(LocalUpdate{})

	sys := actors.NewRemoteSystem(myAddress)
	if len(os.Args) > 2 {
		index, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		sys.SpawnNamed(func() actors.Actor { return NewTrainer(os.Args[2], index) }, "trainer")
	} else {
		sys.SpawnNamed(func() actors.Actor {
			return NewAggregator()
		}, "aggregator")
	}

	select {}
}
