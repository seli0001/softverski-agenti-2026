package actors

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type System struct {
	mailboxes    map[PID]*Mailbox
	watchers     map[PID][]PID
	children     map[PID][]PID
	parents      map[PID]PID
	serverConfig *tls.Config
	clientConfig *tls.Config
	address      string
	counter      int
	mu           sync.Mutex
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

func NewRemoteSystemTLS(address string, certFile, keyFile, caFile string) *System {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}
	bytes, err2 := os.ReadFile(caFile)
	if err2 != nil {
		panic(err2)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(bytes)

	if !ok {
		panic("failed to parse root certificate")
	}

	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	clientConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "actor-node",
	}

	s := &System{
		mailboxes:    make(map[PID]*Mailbox),
		watchers:     make(map[PID][]PID),
		counter:      0,
		mu:           sync.Mutex{},
		children:     make(map[PID][]PID),
		parents:      make(map[PID]PID),
		clientConfig: clientConfig,
		serverConfig: serverConfig,
		address:      address,
	}
	go s.listen()
	return s
}

func (s *System) listen() {
	gob.Register(Envelope{})
	var conn net.Listener
	var err error
	if s.serverConfig != nil {
		conn, err = tls.Listen("tcp", s.address, s.serverConfig)
	} else {
		conn, err = net.Listen("tcp", s.address)

	}
	if err != nil {
		panic(err)
	}
	for {
		c, err := conn.Accept()
		if err != nil {
			return
		}
		var msg Envelope
		errD := gob.NewDecoder(c).Decode(&msg)
		if errD != nil {
			c.Close()
			continue
		}
		s.mu.Lock()
		recipient := s.mailboxes[msg.Recipient]
		s.mu.Unlock()
		if recipient == nil {
			continue
		}
		recipient.send(msg.Message)
		c.Close()
	}
}

func (s *System) Spawn(producer func() Actor) PID {
	s.mu.Lock()
	s.counter++
	id := fmt.Sprintf("actor %d", s.counter)
	s.mu.Unlock()
	return s.spawnLogic(producer, id)
}

func (s *System) spawnLogic(producer func() Actor, id string) PID {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *System) SpawnNamed(producer func() Actor, name string) PID {
	return s.spawnLogic(producer, name)
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
	var conn net.Conn
	var err error
	if s.clientConfig != nil {
		conn, err = tls.Dial("tcp", pid.Address, s.clientConfig)
	} else {
		conn, err = net.Dial("tcp", pid.Address)
	}
	if err != nil {
		return
	}
	defer conn.Close()

	gob.NewEncoder(conn).Encode(Envelope{
		Recipient: pid,
		Message:   msg,
	})
}

func (s *System) SendLater(pid PID, msg any, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		s.Send(pid, msg)
	}()
}

func (s *System) Stop(pid PID) {
	s.mu.Lock()
	mb := s.mailboxes[pid]
	delete(s.mailboxes, pid)
	w := s.watchers[pid]
	delete(s.watchers, pid)
	children := s.children[pid]
	delete(s.children, pid)
	delete(s.parents, pid)
	s.mu.Unlock()
	for _, child := range children {
		s.Stop(child)
	}

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
