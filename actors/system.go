package actors

import (
	"fmt"
	"sync"
)

type System struct {
	mailboxes map[PID]*Mailbox
	counter   int
	mu        sync.Mutex
}

func NewSystem() *System {
	return &System{
		mailboxes: make(map[PID]*Mailbox),
		counter:   0,
	}
}

func (s *System) Spawn(actor Actor) PID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	id := fmt.Sprintf("actor %d", s.counter)

	pid := PID{
		id: id,
	}
	ac := actorContext{
		self: pid,
		sys:  s,
	}
	mb := NewMailbox(actor, ac)
	mb.start()

	s.mailboxes[pid] = mb

	mb.send(Started{})

	return pid
}

func (s *System) Send(pid PID, msg any) {
	s.mu.Lock()
	mb := s.mailboxes[pid]
	s.mu.Unlock()
	if mb == nil {
		return
	}
	mb.send(msg)
}

func (s *System) Stop(pid PID) {
	s.mu.Lock()
	mb := s.mailboxes[pid]
	delete(s.mailboxes, pid)
	s.mu.Unlock()
	if mb == nil {
		return
	}
	mb.send(Stopping{})
}
