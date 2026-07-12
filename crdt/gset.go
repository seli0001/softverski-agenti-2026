package crdt

type GSet struct {
	state map[string]bool
}

func NewGSet() *GSet {
	return &GSet{make(map[string]bool)}
}

func (s *GSet) Add(element string) {
	s.state[element] = true
}

func (s *GSet) Contains(element string) bool {
	return s.state[element]
}

func (s *GSet) Merge(other *GSet) {
	for k, v := range other.state {
		s.state[k] = v
	}
}
