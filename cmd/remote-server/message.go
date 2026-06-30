package remote_server

import "agenti/actors"

type HelloWorld struct {
	Message string
	ReplyTo actors.PID
}
