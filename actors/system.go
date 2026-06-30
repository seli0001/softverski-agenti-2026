package actors

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

type System struct {
	mailboxes map[PID]*Mailbox
	watchers  map[PID][]PID
	children  map[PID][]PID
	parents   map[PID]PID
	address   string
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

func NewRemoteSystem(address string) *System {
	s := &System{
		mailboxes: make(map[PID]*Mailbox),
		watchers:  make(map[PID][]PID),
		counter:   0,
		children:  make(map[PID][]PID),
		address:   address,
		parents:   make(map[PID]PID),
	}
	go s.listen()
	return s
}

func (s *System) listen() {
	gob.Register(Envelope{})
	conn, err := net.Listen("tcp", s.address)
	if err != nil {
		panic(err)
	}
	for {
		c, err := conn.Accept()
		if err != nil {
			continue
		}
		var msg Envelope
		gob.NewDecoder(c).Decode(&msg)
		s.mu.Lock()
		recipient := s.mailboxes[msg.Recipient]
		s.mu.Unlock()
		if recipient == nil {
			continue
		}
		recipient.send(msg.Message)
	}
}

func (s *System) Spawn(producer func() Actor) PID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	id := fmt.Sprintf("actor %d", s.counter)
	pid := PID{
		Id:      id,
		Address: s.address,
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
	var mb *Mailbox
	if pid.Address == s.address {
		mb = s.mailboxes[pid]
	}
	s.mu.Unlock()
	if mb == nil {
		if pid.Address == s.address {
			return
		}
		s.sendRemote(pid, msg)
		return
	}
	mb.send(msg)
}

func (s *System) sendRemote(pid PID, msg any) {
	gob.Register(Envelope{})
	conn, err := net.Dial("tcp", pid.Address)
	if err != nil {
		return
	}
	defer conn.Close()

	gob.NewEncoder(conn).Encode(Envelope{
		Recipient: pid,
		Message:   msg,
	})
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
