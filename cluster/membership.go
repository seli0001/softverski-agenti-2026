package cluster

import (
	"agenti/actors"
	"fmt"
)

type Membership struct {
	contact string
	peers   map[string]bool
}

func NewMembership(contactAddress string) *Membership {
	return &Membership{
		contact: contactAddress,
		peers:   make(map[string]bool),
	}
}

func (m *Membership) Receive(ctx actors.Context, msg any) {
	switch ms := msg.(type) {
	case actors.Started:
		if m.contact != "" {
			receiver := actors.PID{
				Address: m.contact,
				Id:      "membership",
			}
			ctx.Send(receiver,
				Join{
					Address: ctx.Self().Address,
				})
		} else {
			fmt.Println("First one here")
		}
	case Join:
		peerlist := []string{}
		for peer := range m.peers {
			peerlist = append(peerlist, peer)
			ctx.Send(actors.PID{
				Address: peer,
				Id:      "membership",
			}, ForwardJoin{
				NewJoinerAddress: ms.Address,
			})
		}
		m.peers[ms.Address] = true
		ctx.Send(actors.PID{
			Address: ms.Address,
			Id:      "membership",
		}, PeerList{
			Addresses: append(peerlist, ctx.Self().Address),
		})
		fmt.Println(m.peers)
	case PeerList:
		for _, peer := range ms.Addresses {
			if peer == ctx.Self().Address {
				continue
			}
			m.peers[peer] = true
		}
		fmt.Println(m.peers)
	case ForwardJoin:
		m.peers[ms.NewJoinerAddress] = true
		fmt.Println(m.peers)
	}
}
