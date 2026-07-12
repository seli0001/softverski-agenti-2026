package crdt

type GCounter struct {
	state  map[string]int
	nodeId string
}

func NewGCounter(nodeId string) *GCounter {
	return &GCounter{
		state:  make(map[string]int),
		nodeId: nodeId,
	}
}

func (g *GCounter) Increment() {
	g.state[g.nodeId]++
}

func (g *GCounter) Value() int {
	value := 0
	for _, v := range g.state {
		value += v
	}
	return value
}

func (g *GCounter) Merge(other GCounter) {
	for k, v := range other.state {
		g.state[k] = max(g.state[k], v)
	}
}
