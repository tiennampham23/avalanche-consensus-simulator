package node

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tiennampham23/avalanche-consensus-simulator/chain"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
)

type Node struct {
	*chain.BlockChain
}

func InitNode(ctx context.Context, config chain.Config, discovery *p2p.Discovery) (*Node, error) {
	s := &Node{}
	blockchain, err := chain.InitBlockChain(ctx, config, discovery)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "unable to init blockchain")
	}
	s.BlockChain = blockchain
	return s, nil
}
