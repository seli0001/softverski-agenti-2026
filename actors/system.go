package actors

import (
	"fmt"
	"sync"
)

type System struct {
	mailboxes map[PID]*Mailbox
	watchers  map[PID][]PID
	counter   int
	mu        sync.Mutex
}

func NewSystem() *System {
	return &System{
		mailboxes: make(map[PID]*Mailbox),
		watchers:  make(map[PID][]PID),
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
	w := s.watchers[pid]
	delete(s.watchers, pid)
	s.mu.Unlock()

	if w != nil {
		for _, watcher := range w {
			s.Send(watcher, Terminated{PID: pid})
		}
	}

	if mb == nil {
		return
	}
	mb.send(Stopping{})
}

func (s *System) Watch(watcher, pid PID) {
	s.mu.Lock()
	s.watchers[pid] = append(s.watchers[pid], watcher)
	s.mu.Unlock()
}
