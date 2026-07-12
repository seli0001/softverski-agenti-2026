package cluster

type Join struct {
	Address string
}

type PeerList struct {
	Addresses []string
}

type ForwardJoin struct {
	NewJoinerAddress string
	Sender           string
	TTL              int
}

type Ping struct {
	Sender string
}

type Pong struct {
	Sender string
}

type NeighborRequest struct {
	Sender       string
	HighPriority bool
}

type NeighborResponse struct {
	Sender   string
	Accepted bool
}

type Disconnect struct {
	Sender string
}
