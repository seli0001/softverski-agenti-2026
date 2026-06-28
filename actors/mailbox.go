package actors

type Mailbox struct {
	ch    chan any
	actor Actor
	ctx   Context
}

func NewMailbox(actor Actor, ctx Context) *Mailbox {
	return &Mailbox{
		ch:    make(chan any, 100),
		ctx:   ctx,
		actor: actor,
	}
}

func (m *Mailbox) send(msg any) {
	m.ch <- msg
}

func (m *Mailbox) start() {
	go func() {
		for msg := range m.ch {
			m.actor.Receive(m.ctx, msg)
			if _, ok := msg.(Stopping); ok {
				return
			}
		}
	}()
}
