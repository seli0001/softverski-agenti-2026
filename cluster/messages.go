package cluster

type Join struct {
	Address string
}

type PeerList struct {
	Addresses []string
}

type ForwardJoin struct {
	NewJoinerAddress string
}
