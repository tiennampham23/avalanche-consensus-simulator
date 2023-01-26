package chain

import (
	"context"
	"encoding/json"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/snow/consensus"
	"math/rand"
)

type request struct {
	Index int `json:"index"`
}

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

func InitBlockChain(cfg Config) (*BlockChain, error) {
	blockChainState := InitBlockChainState()
	blockchain := &BlockChain{
		BlockChainState: blockChainState,
		cfg:             cfg,
	}
	getDataFromBlockIndexCb := func(req []byte) ([]byte, error) {
		return blockchain.getBlockDataByIndex(req)
	}
	client, err := p2p.InitClient(cfg.P2PConfig, getDataFromBlockIndexCb)
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
	peers := c.client.Peers()
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
func (c *BlockChain) getDataFromOtherPeerByIndex(ctx context.Context, peer *peer.AddrInfo, index int) ([]byte, error) {
	req := request{
		Index: index,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	blockDataResponse, err := c.client.GetBlockData(ctx, peer, reqData)
	if err != nil {
		return nil, err
	}
	if len(blockDataResponse) == 0 {
		return nil, errors.New("the response is empty")
	}

	var blockData []byte
	err = json.Unmarshal(blockDataResponse, &blockData)
	if err != nil {
		return nil, err
	}

	return blockData, nil
}

func (c *BlockChain) getBlockDataByIndex(reqData []byte) ([]byte, error) {
	var req request
	err := json.Unmarshal(reqData, &req)
	if err != nil {
		return nil, err
	}
	if req.Index < 0 {
		return nil, errors.New("Index is smaller than 0")
	}
	if req.Index >= len(c.Blocks) {
		return nil, errors.New("Index is larger than the length of blocks")
	}

	block := c.Blocks[req.Index]

	b, err := json.Marshal(block.Data)
	if err != nil {
		return nil, err
	}

	return b, nil
}
