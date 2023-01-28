package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tiennampham23/avalanche-consensus-simulator/chain"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/node"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"github.com/tiennampham23/avalanche-consensus-simulator/snow/consensus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	protocolID  = "avalanche-consensus-simulator/1.0.0"
	serviceName = "avalanche-consensus"
	host        = "127.0.0.1"
	k           = 3
	alpha       = 2
	beta        = 1
	numOfBlocks = 4
)

func main() {
	log.Build()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	discovery, err := runDiscovery()
	if err != nil {
		log.Fatal(err)
	}

	healthyPeersTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-sigs:
				return
			case <-healthyPeersTicker.C:
				err := discovery.HealthCheckPeers()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	time.Sleep(1 * time.Second)

	for j := 0; j < 200; j++ {
		wg.Add(1)
		go func(j int, discovery *p2p.Discovery) {
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
				Port:       10000 + j,
			}
			parameters := consensus.Parameters{
				K:     k,
				Alpha: alpha,
				Beta:  beta,
			}
			n, err := node.InitNode(ctx, chain.Config{
				P2PConfig:           p2pConfig,
				ConsensusParameters: parameters,
			}, discovery)
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
			blockChainState := ""
			for _, b := range n.Blocks {
				data := b.Data[0]
				blockChainState += fmt.Sprintf("%d", data)
			}
			log.Infof("client: %d, block: %s", j, blockChainState)

			<-doneChan
		}(j, discovery)
	}
	wg.Wait()

}

func runDiscovery() (*p2p.Discovery, error) {
	discovery := p2p.InitDiscovery()
	r := gin.New()
	discovery.Router(r)
	go func() {
		err := r.Run(discovery.Address)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return discovery, nil
}
