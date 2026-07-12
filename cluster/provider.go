package cluster

import (
	"agenti/actors"
	"fmt"
	"time"
)

type CheckMembers struct{}

type HeartbeatTick struct{}

type Provider struct {
	members map[string]bool
	waiting map[string]bool
}

func NewProvider() *Provider {
	return &Provider{
		members: make(map[string]bool),
		waiting: make(map[string]bool),
	}
}

func (p *Provider) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		ctx.SendLater(ctx.Self(), CheckMembers{}, 5*time.Second)
	case Register:
		p.members[m.Address] = true
		fmt.Println("members: ", p.members)
		p.broadcast(ctx)
	case Heartbeat:
		delete(p.waiting, m.Address)
	case CheckMembers:
		if len(p.waiting) > 0 {
			for member := range p.waiting {
				delete(p.members, member)
			}
			fmt.Println("members: ", p.members)
			p.broadcast(ctx)
		}
		p.waiting = make(map[string]bool)
		for mem := range p.members {
			p.waiting[mem] = true
		}
		ctx.SendLater(ctx.Self(), CheckMembers{}, 5*time.Second)
	}
}

func (p *Provider) broadcast(ctx actors.Context) {
	var toSend []string

	for member := range p.members {
		toSend = append(toSend, member)
	}
	for member := range p.members {
		ctx.Send(actors.PID{
			Address: member,
			Id:      "member",
		}, Members{Addresses: toSend})
	}
}

type Client struct {
	providerAddress string
	selfAddress     string
	members         []string
}

func NewClient(providerAddress string) *Client {
	return &Client{
		providerAddress: providerAddress,
		members:         make([]string, 0),
	}
}

func (c *Client) Receive(ctx actors.Context, msg any) {
	switch m := msg.(type) {
	case actors.Started:
		c.selfAddress = ctx.Self().Address
		ctx.Send(actors.PID{
			Address: c.providerAddress,
			Id:      "provider",
		}, Register{
			Address: c.selfAddress,
		})
		ctx.SendLater(ctx.Self(), HeartbeatTick{}, 2*time.Second)
	case HeartbeatTick:
		ctx.Send(actors.PID{
			Address: c.providerAddress,
			Id:      "provider",
		}, Heartbeat{
			Address: c.selfAddress,
		})
		ctx.SendLater(ctx.Self(), HeartbeatTick{}, 2*time.Second)
	case Members:
		c.members = m.Addresses
		fmt.Println("members: ", c.members)
	}
}
