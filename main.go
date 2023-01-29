package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/phayes/freeport"
	"github.com/tiennampham23/avalanche-consensus-simulator/chain"
	"github.com/tiennampham23/avalanche-consensus-simulator/network/p2p"
	"github.com/tiennampham23/avalanche-consensus-simulator/node"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"github.com/tiennampham23/avalanche-consensus-simulator/snow/consensus"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	protocolID          = "avalanche-consensus-simulator/1.0.0"
	serviceName         = "avalanche-consensus"
	host                = "127.0.0.1"
	k                   = 3
	alpha               = 2
	beta                = 2
	numOfBlocks         = 500
	possiblePreferences = 2
	numOfNodes          = 200
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
	time.Sleep(2 * time.Second)

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

	for j := 0; j < numOfNodes; j++ {
		wg.Add(1)
		go func(j int, discovery *p2p.Discovery) {
			doneChan := make(chan bool, 1)
			defer wg.Done()
			go func() {
				<-sigs
				doneChan <- true
			}()
			ctx := context.Background()
			freePort, err := freeport.GetFreePort()
			if err != nil {
				log.Fatal(err)
			}
			p2pConfig := p2p.Config{
				Name:       serviceName,
				ProtocolID: protocolID,
				Host:       host,
				Port:       freePort,
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
			time.Sleep(1 * time.Second)

			for i := 0; i < numOfBlocks; i++ {
				data := make([]byte, 0)
				l := float64(possiblePreferences) * 2
				r := rand.Intn(int(l))
				data = append(data, byte(r))

				data = append(data, byte(i))
				newBlock := &chain.Block{
					Data: data,
				}
				err := n.Add(newBlock)
				if err != nil {
					log.Fatal(err)
				}
			}
			beforeBlockChainState := ""
			for _, b := range n.Blocks {
				data := b.Data[0]
				beforeBlockChainState += fmt.Sprintf("%d", data)
			}
			log.Infof("Before sync, data of node: %d is %s", j, beforeBlockChainState)

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
	gin.SetMode(gin.ReleaseMode)
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
