package p2p

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type Discovery struct {
	peerChan chan peer.AddrInfo
}

func (d *Discovery) HandlePeerFound(peerInfo peer.AddrInfo) {
	d.peerChan <- peerInfo
}

func InitDiscovery(host host.Host, serviceName string) (chan peer.AddrInfo, error) {
	discovery := &Discovery{
		peerChan: make(chan peer.AddrInfo),
	}

	ser := mdns.NewMdnsService(host, serviceName, discovery)
	if err := ser.Start(); err != nil {
		return nil, err
	}

	return discovery.peerChan, nil
}
