package main

import (
	"agenti/actors"
	remoteserver "agenti/cmd/remote-server"
	"encoding/gob"
	"time"
)

func main() {
	gob.Register(remoteserver.HelloWorld{})
	sys := actors.NewSystem()
	sys.Send(actors.PID{
		Id:      "actor 1",
		Address: "127.0.0.1:9001",
	}, remoteserver.HelloWorld{
		Message: "hello world",
	})
	time.Sleep(1 * time.Second)
}
