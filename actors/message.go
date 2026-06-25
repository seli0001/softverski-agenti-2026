package actors

type Message = interface{}

type PID struct {
	id string
}

func (p PID) ID() string {
	return p.id
}
