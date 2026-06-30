package actors

type Message = interface{}

type PID struct {
	Id      string
	Address string
}

func (p PID) ID() string {
	return p.Id
}

type Envelope struct {
	Sender    PID
	Recipient PID
	Message   Message
}
