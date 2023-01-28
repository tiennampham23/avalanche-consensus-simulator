package chain

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tiennampham23/avalanche-consensus-simulator/model"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/snow/consensus"
	"math/rand"
)

type Config struct {
	P2PConfig           p2p.Config
	ConsensusParameters consensus.Parameters
}

type BlockChain struct {
	*BlockChainState
	client    *p2p.Client
	cfg       Config
	isRunning bool
}

func InitBlockChain(ctx context.Context, cfg Config, discovery *p2p.Discovery) (*BlockChain, error) {
	blockChainState := InitBlockChainState()
	blockchain := &BlockChain{
		BlockChainState: blockChainState,
		cfg:             cfg,
	}
	getDataFromBlockIndexCb := func(index int) ([]byte, error) {
		return blockchain.getBlockDataByIndex(index)
	}
	client, err := p2p.InitClient(ctx, cfg.P2PConfig, discovery, getDataFromBlockIndexCb)
	if err != nil {
		return nil, err
	}
	blockchain.client = client

	return blockchain, nil
}

func (c *BlockChain) Sync(ctx context.Context) error {
	if c.isRunning {
		return nil
	}

	c.isRunning = true
	for i, block := range c.Blocks {
		snowBallConsensus, err := consensus.NewConsensus(
			c.cfg.ConsensusParameters,
			block.Data,
		)
		if err != nil {
			return err
		}
		getBlockDataFromRandomKCb := func(k int) ([][]byte, error) {
			return c.getBlockDataFromKPeersByIndex(ctx, i, k)
		}
		setDataCb := func(data []byte) error {
			return block.SetData(data)
		}
		err = snowBallConsensus.Sync(ctx, setDataCb, getBlockDataFromRandomKCb)
		if err != nil {
			return errors.Wrap(err, "unable to sync the consensus")
		}
	}

	c.isRunning = false
	return nil
}

func (c *BlockChain) getBlockDataFromKPeersByIndex(ctx context.Context, index int, k int) ([][]byte, error) {
	peers, err := c.client.Peers()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get the peers from the discovery")
	}
	lengthPeers := len(peers)
	var preferencesFromOtherPeers [][]byte

	var count int
	for _, i := range rand.Perm(lengthPeers) {
		randomPeer := peers[i]
		if randomPeer == nil {
			continue
		}
		preference, err := c.getDataFromOtherPeerByIndex(ctx, randomPeer, index)
		if err != nil || len(preference) == 0 {
			continue
		}
		preferencesFromOtherPeers = append(preferencesFromOtherPeers, preference)

		count++
		// get the preferences of the k random peers from the peers
		if count >= k {
			break
		}
	}
	return preferencesFromOtherPeers, nil
}
func (c *BlockChain) getDataFromOtherPeerByIndex(ctx context.Context, peer *p2p.Peer, index int) ([]byte, error) {
	req := model.GetBlockDataByIndexRequest{
		Index: index,
	}

	blockDataResponse, err := c.client.GetBlockData(ctx, peer, req)
	if err != nil {
		return nil, err
	}
	if len(blockDataResponse) == 0 {
		return nil, errors.New("the response is empty")
	}

	return blockDataResponse, nil
}

func (c *BlockChain) getBlockDataByIndex(index int) ([]byte, error) {
	if index < 0 {
		return nil, errors.New("Index is smaller than 0")
	}
	if index >= len(c.Blocks) {
		return nil, errors.New("Index is larger than the length of blocks")
	}

	block := c.Blocks[index]

	b, err := json.Marshal(block.Data)
	if err != nil {
		return nil, err
	}

	return b, nil
}
