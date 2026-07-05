package main

import (
	"agenti/actors"
	remoteserver "agenti/cmd/remote-server"
	"encoding/gob"
	"fmt"
)

type Receiver struct {
	counter int
}

func (r *Receiver) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		fmt.Println(ctx.Self().ID())
	case remoteserver.HelloWorld:
		r.counter++
		if r.counter > 3 {
			fmt.Println(r.counter)
		} else {
			ctx.Send(m.ReplyTo, remoteserver.HelloWorld{Message: "odgovor", ReplyTo: ctx.Self()})
		}
	}
}

func main() {
	gob.Register(remoteserver.HelloWorld{})
	sys := actors.NewRemoteSystemTLS("127.0.0.1:9002", "certs/node-cert.pem", "certs/node-key.pem", "certs/ca-cert.pem")
	pid := sys.Spawn(func() actors.Actor {
		return &Receiver{
			counter: 0,
		}
	})
	sys.Send(actors.PID{
		Id:      "actor 1",
		Address: "127.0.0.1:9001",
	}, remoteserver.HelloWorld{
		Message: "hello world",
		ReplyTo: pid,
	})
	fmt.Println(pid)
	select {}

}
