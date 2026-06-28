package actors

type Mailbox struct {
	ch       chan any
	actor    Actor
	ctx      Context
	behavior func(Context, any)
}

func NewMailbox(actor Actor, ctx Context) *Mailbox {
	return &Mailbox{
		ch:       make(chan any, 100),
		ctx:      ctx,
		actor:    actor,
		behavior: actor.Receive,
	}
}

func (m *Mailbox) send(msg any) {
	m.ch <- msg
}

func (m *Mailbox) start() {
	go func() {
		for msg := range m.ch {
			m.behavior(m.ctx, msg)
			if _, ok := msg.(Stopping); ok {
				return
			}
		}
	}()
}
