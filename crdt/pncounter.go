package crdt

type PNCounter struct {
	positive *GCounter
	negative *GCounter
}

func NewPNCounter(nodeId string) *PNCounter {
	return &PNCounter{
		positive: NewGCounter(nodeId),
		negative: NewGCounter(nodeId),
	}
}

func (g *PNCounter) Increment() {
	g.positive.Increment()
}

func (g *PNCounter) Decrement() {
	g.negative.Increment()
}

func (g *PNCounter) Value() int {
	return g.positive.Value() - g.negative.Value()
}

func (g *PNCounter) Merge(other *PNCounter) {
	g.positive.Merge(*other.positive)
	g.negative.Merge(*other.negative)

}
