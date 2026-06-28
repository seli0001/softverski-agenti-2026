package actors

type Mailbox struct {
	ch        chan any
	actor     Actor
	ctx       Context
	behavior  func(Context, any)
	onFailure func(any)
	producer  func() Actor
}

func NewMailbox(producer func() Actor, ctx Context) *Mailbox {
	actor := producer()
	return &Mailbox{
		ch:       make(chan any, 100),
		ctx:      ctx,
		actor:    actor,
		producer: producer,
		behavior: actor.Receive,
	}
}

func (m *Mailbox) send(msg any) {
	m.ch <- msg
}

func (m *Mailbox) start() {
	go func() {
		for msg := range m.ch {
			m.invoke(msg)
			if _, ok := msg.(Stopping); ok {
				return
			}
		}
	}()
}

func (m *Mailbox) invoke(msg any) {
	defer func() {
		if r := recover(); r != nil {
			if m.onFailure != nil {
				m.onFailure(r)
			}
		}
	}()
	m.behavior(m.ctx, msg)
}

func (m *Mailbox) Restart() {
	m.actor = m.producer()
	m.behavior = m.actor.Receive
}
