package p2p

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"net/http"
	"sync"
	"time"
)

const (
	DiscoveryHost = "0.0.0.0"
	DiscoveryPort = 8080
)

type Discovery struct {
	mu          sync.Mutex
	Address     string
	Peers       []*Peer
	restyClient *resty.Client
}

type RegisterPeerRequest struct {
	Peer *Peer `json:"peer"`
}

type RegisterPeerResponse struct {
	Peer []*Peer `json:"peers"`
}

func (d *Discovery) Router(r *gin.Engine) {
	r.POST("/register-peer", d.RegisterPeer)
	r.GET("/peers", d.GetPeers)
}

func InitDiscovery() *Discovery {
	restyClient := resty.
		New().
		SetRetryCount(5).
		SetRetryWaitTime(2 * time.Second).
		AddRetryCondition(func(response *resty.Response, err error) bool {
			return response.StatusCode() == http.StatusTooManyRequests ||
				response.StatusCode() == http.StatusGatewayTimeout ||
				response.StatusCode() == http.StatusServiceUnavailable ||
				response.StatusCode() == http.StatusBadGateway ||
				response.StatusCode() == http.StatusInternalServerError
		})
	address := fmt.Sprintf("%s:%d", DiscoveryHost, DiscoveryPort)
	return &Discovery{
		Address:     address,
		Peers:       make([]*Peer, 0),
		restyClient: restyClient,
	}
}

func (d *Discovery) RegisterPeer(c *gin.Context) {
	var req RegisterPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, nil)
	}
	isAvailablePeer := true
	for _, peer := range d.Peers {
		if peer.ID == req.Peer.ID {
			isAvailablePeer = false
		}
	}
	if isAvailablePeer {
		d.Peers = append(d.Peers, req.Peer)
	}
	c.JSON(200, RegisterPeerResponse{
		Peer: d.Peers,
	})
}

func (d *Discovery) GetPeers(c *gin.Context) {
	c.JSON(200, map[string]interface{}{
		"peers": d.Peers,
	})
}

func (d *Discovery) HealthCheckPeers() error {
	log.Debug("healthy check start")
	var wg sync.WaitGroup
	peers := make([]*Peer, 0)

	for _, p := range d.Peers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := d.restyClient.R().Get(fmt.Sprintf("http://%s/liveliness", p.Address))
			if err != nil {
				return
			}
			peers = append(peers, p)
		}()
	}
	wg.Wait()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Peers = peers

	return nil
}
