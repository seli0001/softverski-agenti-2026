package cluster

import (
	"agenti/actors"
	"fmt"
	"time"
)

const activeSize = 3
const passiveSize = 10
const activeRandomWalkLenght = 3

type Membership struct {
	selfAddress string
	contact     string
	active      map[string]bool
	passive     map[string]bool
	waiting     map[string]bool
}

type PingPeers struct{}

type Timeout struct{}

type GetPeers struct {
	ReplyTo actors.PID
}

type Peers struct {
	Addresses []string
}

func NewMembership(contactAddress string) *Membership {
	return &Membership{
		contact: contactAddress,
		active:  make(map[string]bool),
		passive: make(map[string]bool),
		waiting: make(map[string]bool),
	}
}

func (m *Membership) addToPassive(address string) {
	if address == m.selfAddress {
		return
	}
	if m.active[address] {
		return
	}
	var toRemove string
	if len(m.passive) >= passiveSize {
		for peer := range m.passive {
			toRemove = peer
			break
		}
		delete(m.passive, toRemove)
	}
	m.passive[address] = true
}

func (m *Membership) addToActive(address string, ctx actors.Context) {
	var toRemove string
	if len(m.active) >= activeSize {
		for peer := range m.active {
			toRemove = peer
			break
		}
		delete(m.active, toRemove)
		m.addToPassive(toRemove)
		ctx.Send(actors.PID{
			Id:      "membership",
			Address: toRemove,
		}, Disconnect{
			Sender: ctx.Self().Address,
		})
	}
	m.active[address] = true
	delete(m.passive, address)
}

func (m *Membership) Receive(ctx actors.Context, msg any) {
	switch ms := msg.(type) {
	case actors.Started:
		m.selfAddress = ctx.Self().Address
		ctx.SendLater(ctx.Self(), PingPeers{}, 5*time.Second)
		if m.contact != "" {
			receiver := actors.PID{
				Address: m.contact,
				Id:      "membership",
			}
			m.active[m.contact] = true
			ctx.Send(receiver,
				Join{
					Address: ctx.Self().Address,
				})
		} else {
			fmt.Println("First one here")
		}
	case Join:
		peerlist := []string{}
		for peer := range m.active {
			peerlist = append(peerlist, peer)
			ctx.Send(actors.PID{
				Address: peer,
				Id:      "membership",
			}, ForwardJoin{
				NewJoinerAddress: ms.Address,
				Sender:           ctx.Self().Address,
				TTL:              activeRandomWalkLenght,
			})
		}
		for peer := range m.passive {
			peerlist = append(peerlist, peer)
		}
		m.addToActive(ms.Address, ctx)
		ctx.Send(actors.PID{
			Address: ms.Address,
			Id:      "membership",
		}, PeerList{
			Addresses: append(peerlist, ctx.Self().Address),
		})
		fmt.Println(m.active)
	case PeerList:
		for _, peer := range ms.Addresses {
			if peer == ctx.Self().Address {
				continue
			}
			m.addToPassive(peer)
		}
		fmt.Println(m.passive)
	case ForwardJoin:
		m.addToPassive(ms.NewJoinerAddress)
		fmt.Println(m.passive)

		toSend := ""

		for peer := range m.active {
			if peer == ms.NewJoinerAddress || peer == ms.Sender {
				continue
			}
			toSend = peer
			break
		}
		if toSend == "" || ms.TTL == 0 {
			ctx.Send(actors.PID{
				Address: ms.NewJoinerAddress,
				Id:      "membership",
			},
				NeighborRequest{
					Sender:       ctx.Self().Address,
					HighPriority: false,
				})
			return
		}
		if ms.TTL > 0 {
			ctx.Send(
				actors.PID{
					Address: toSend,
					Id:      "membership",
				}, ForwardJoin{
					TTL:              ms.TTL - 1,
					NewJoinerAddress: ms.NewJoinerAddress,
					Sender:           ctx.Self().Address,
				})
		}

	case Ping:
		ctx.Send(actors.PID{Address: ms.Sender, Id: "membership"}, Pong{
			Sender: ctx.Self().Address,
		})
	case Pong:
		delete(m.waiting, ms.Sender)
	case PingPeers:
		for peer := range m.active {
			ctx.Send(actors.PID{Address: peer, Id: "membership"}, Ping{
				Sender: ctx.Self().Address,
			})
			m.waiting[peer] = true
		}
		fmt.Println("active:", m.active, "passive:", m.passive)
		ctx.SendLater(ctx.Self(), Timeout{}, 2*time.Second)
		ctx.SendLater(ctx.Self(), PingPeers{}, 5*time.Second)

	case Timeout:
		for peer := range m.waiting {
			delete(m.active, peer)
			fmt.Println("Alive: ", m.active)
			ctx.Send(actors.PID{Address: peer, Id: "membership"}, Disconnect{Sender: ctx.Self().Address})
		}
		m.waiting = map[string]bool{}
		if len(m.active) < activeSize && len(m.passive) > 0 {
			var toAdd string
			var highPriority bool
			for peer := range m.passive {
				toAdd = peer
				break
			}
			if len(m.active) == 0 {
				highPriority = true
			}
			delete(m.passive, toAdd)
			ctx.Send(actors.PID{
				Address: toAdd,
				Id:      "membership",
			},
				NeighborRequest{
					Sender:       ctx.Self().Address,
					HighPriority: highPriority,
				})
		}
	case NeighborRequest:
		accepted := false
		if len(m.active) < activeSize || ms.HighPriority {
			m.addToActive(ms.Sender, ctx)
			accepted = true
		}

		ctx.Send(actors.PID{Address: ms.Sender, Id: "membership"}, NeighborResponse{
			Sender:   ctx.Self().Address,
			Accepted: accepted,
		})
	case NeighborResponse:
		if !ms.Accepted {
			m.addToPassive(ms.Sender)
		} else {
			m.addToActive(ms.Sender, ctx)
		}
	case Disconnect:
		delete(m.active, ms.Sender)
		m.addToPassive(ms.Sender)
	case GetPeers:
		var peerList []string
		for peer := range m.active {
			peerList = append(peerList, peer)
		}
		ctx.Send(ms.ReplyTo, Peers{
			Addresses: peerList,
		})
	}
}
