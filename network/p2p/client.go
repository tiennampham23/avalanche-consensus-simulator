package p2p

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type Client struct {
	cfg                   Config
	client                host.Host
	peers                 []*peer.AddrInfo
	getBlockDataByIndexCb func([]byte) ([]byte, error)
}

func InitClient(cfg Config, getBlockDataByIndexCb func([]byte) ([]byte, error)) (*Client, error) {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}

	client := &Client{
		cfg:                   cfg,
		getBlockDataByIndexCb: getBlockDataByIndexCb,
		peers:                 make([]*peer.AddrInfo, 0),
	}

	p2pClient, err := InitP2P(client.cfg.Host)
	if err != nil {
		return nil, err
	}
	log.Infof("Init P2P Client successfully, address: %s, id: %s", p2pClient.Addrs(), p2pClient.ID())

	// This gets called every time a peer connects and opens a stream to this node.
	p2pClient.SetStreamHandler(protocol.ID(client.cfg.ProtocolID), client.handleNewStream)

	// Discovery other node
	peerChan, err := InitDiscovery(p2pClient, client.cfg.Name)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			newPeer := <-peerChan
			client.handleNewPeer(newPeer)
		}
	}()

	client.client = p2pClient

	return client, nil
}

func (c *Client) GetBlockData(ctx context.Context, peer *peer.AddrInfo, data []byte) ([]byte, error) {
	// open a stream, this stream will be handled by handleStream other end
	stream, err := c.client.NewStream(ctx, peer.ID, protocol.ID(c.cfg.ProtocolID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new stream")
	}
	defer stream.Close()

	_, err = stream.Write(data)
	if err != nil {
		return nil, err
	}
	err = stream.CloseWrite()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Client) Peers() []*peer.AddrInfo {
	return c.peers
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) handleNewPeer(peer peer.AddrInfo) {
	c.client.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
	c.peers = append(c.peers, &peer)
}

func (c *Client) handleNewStream(stream network.Stream) {
	defer stream.Close()
	reqData, err := io.ReadAll(stream)
	if err != nil {
		log.Fatal(err)
		return
	}

	respData, err := c.getBlockDataByIndexCb(reqData)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(respData) > 0 {
		_, err = stream.Write(respData)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
