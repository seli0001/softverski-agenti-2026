package actors

import (
	"fmt"
	"sync"
)

type System struct {
	mailboxes map[PID]*Mailbox
	watchers  map[PID][]PID
	children  map[PID][]PID
	parents   map[PID]PID
	counter   int
	mu        sync.Mutex
}

func NewSystem() *System {
	return &System{
		mailboxes: make(map[PID]*Mailbox),
		watchers:  make(map[PID][]PID),
		counter:   0,
		mu:        sync.Mutex{},
		children:  make(map[PID][]PID),
		parents:   make(map[PID]PID),
	}
}

func (s *System) Spawn(producer func() Actor) PID {
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
	mb := NewMailbox(producer, ac)
	mb.start()
	mb.onFailure = func(r any) {
		s.handleFailure(pid, r)
	}

	s.mailboxes[pid] = mb

	mb.send(Started{})

	return pid
}

func (s *System) SpawnChildren(parent PID, producer func() Actor) PID {
	child := s.Spawn(producer)
	s.mu.Lock()
	s.children[parent] = append(s.children[parent], child)
	s.parents[child] = parent
	s.mu.Unlock()
	return child
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

func (s *System) Become(pid PID, behavior func(Context, any)) {
	s.mu.Lock()
	mb := s.mailboxes[pid]
	s.mu.Unlock()
	if mb != nil {
		mb.behavior = behavior
	}
}

func (s *System) handleFailure(child PID, reason any) {
	s.mu.Lock()
	parent := s.parents[child]
	parentMB := s.mailboxes[parent]
	childMB := s.mailboxes[child]
	s.mu.Unlock()

	directive := Resume
	if parentMB != nil {
		if sup, ok := parentMB.actor.(Supervisor); ok {
			directive = sup.Decide(reason)
		}
	}

	switch directive {
	case Resume:
	case Stop:
		s.Stop(child)
	case Restart:
		if childMB != nil {
			childMB.Restart()
			childMB.send(Started{})
		}
	}
}
