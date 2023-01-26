package main

import (
	"context"
	"github.com/tiennampham23/avalanche-consensus-simulator/chain"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/node"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"github.com/tiennampham23/avalanche-consensus-simulator/snow/consensus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	protocolID  = "avalanche-consensus-simulator/1.0.0"
	serviceName = "avalanche-consensus"
	host        = "127.0.0.1"
	port        = 0
	k           = 2
	alpha       = 2
	beta        = 2
	numOfBlocks = 4
)

func main() {
	log.Build()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup

	for j := 0; j < 20; j++ {
		wg.Add(1)
		go func(j int) {
			doneChan := make(chan bool, 1)

			defer wg.Done()
			go func() {
				<-sigs
				doneChan <- true
			}()
			ctx := context.Background()
			p2pConfig := p2p.Config{
				Name:       serviceName,
				ProtocolID: protocolID,
				Host:       host,
				Port:       port,
			}
			parameters := consensus.Parameters{
				K:     k,
				Alpha: alpha,
				Beta:  beta,
			}
			n, err := node.InitNode(ctx, chain.Config{
				P2PConfig:           p2pConfig,
				ConsensusParameters: parameters,
			})
			if err != nil {
				log.Fatal(err)
			}

			for i := 0; i < numOfBlocks; i++ {
				data := make([]byte, 0)
				data = append(data, byte(i))
				newBlock := &chain.Block{
					Data: data,
				}
				err := n.Add(newBlock)
				if err != nil {
					log.Fatal(err)
				}
			}

			err = n.Sync(ctx)
			if err != nil {
				log.Fatal(err)
			}

			for i, block := range n.Blocks {
				log.Infof("client: %d, index: %d, block: %v", j, i, block.Data)
			}
			<-doneChan
		}(j)
	}
	wg.Wait()

}
