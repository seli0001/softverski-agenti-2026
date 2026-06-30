package main

import (
	"agenti/actors"
	remoteserver "agenti/cmd/remote-server"
	"encoding/gob"
	"fmt"
)

type Receiver struct{}

func (r *Receiver) Receive(ctx actors.Context, msg any) {
	switch mess := msg.(type) {
	case actors.Started:
		fmt.Println(ctx.Self().ID())
	case remoteserver.HelloWorld:
		fmt.Printf("Message received: %s\n", mess.Message)
		ctx.Send(mess.ReplyTo, remoteserver.HelloWorld{Message: "odgovor", ReplyTo: ctx.Self()})
	}
}

func main() {
	gob.Register(remoteserver.HelloWorld{})
	sys := actors.NewRemoteSystem("127.0.0.1:9001")
	sys.Spawn(func() actors.Actor {
		return &Receiver{}
	})
	select {}
}
