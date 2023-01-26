package chain

import (
	"sync"
)

type Block struct {
	Data      []byte `json:"data"`
	BlockHash string `json:"blockHash"`
	BlockTime int64  `json:"blockTime"`
}

func (b *Block) SetData(data []byte) error {
	b.Data = data
	return nil
}

type BlockChainState struct {
	Blocks []*Block
	mu     sync.Mutex
}

func InitBlockChainState() *BlockChainState {
	blocks := make([]*Block, 0)
	return &BlockChainState{
		Blocks: blocks,
	}
}

func (c *BlockChainState) Add(newBlock *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Blocks = append(c.Blocks, newBlock)

	return nil
}
