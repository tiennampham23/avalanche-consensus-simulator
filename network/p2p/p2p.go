package p2p

import (
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
)

func InitP2P(host string) (host.Host, error) {
	privateKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	// libp2p will find the free port if we pass 0 to the parameters automatically
	server, err := libp2p.New(libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", host, 0)), libp2p.Identity(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "unable to init a new p2p client")
	}

	return server, nil
}
