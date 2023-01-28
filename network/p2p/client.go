package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tiennampham23/avalanche-consensus-simulator/model"
	"github.com/tiennampham23/avalanche-consensus-simulator/pkg/log"
	"net/http"
	"time"
)

type Client struct {
	cfg                   Config
	client                *Peer
	peers                 []*Peer
	r                     *gin.Engine
	peerChan              chan *Peer
	getBlockDataByIndexCb func(int) ([]byte, error)
	discovery             *Discovery
	resty                 *resty.Client
}

func (c *Client) ReceiveMessage(r *gin.Context) {
	r.JSON(200, "OK")
}

func (c *Client) GetDataByIndex(r *gin.Context) {
	var req model.GetBlockDataByIndexRequest
	if err := r.ShouldBindJSON(&req); err != nil {
		r.JSON(400, nil)
		return
	}
	blockData, err := c.getBlockDataByIndexCb(req.Index)
	if err != nil {
		r.JSON(400, nil)
		return
	}
	r.JSON(200, blockData)
}

func (c *Client) Liveliness(r *gin.Context) {
	r.JSON(200, nil)
}

func (c *Client) Router(r *gin.Engine) {
	r.GET("/receive-msg", c.ReceiveMessage)
	r.POST("/get-data-by-index", c.GetDataByIndex)
	r.GET("/liveliness", c.Liveliness)
}

func InitClient(ctx context.Context, cfg Config, discovery *Discovery, getBlockDataByIndexCb func(int) ([]byte, error)) (*Client, error) {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	r := gin.New()
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
	client := &Client{
		cfg:                   cfg,
		getBlockDataByIndexCb: getBlockDataByIndexCb,
		peers:                 make([]*Peer, 0),
		r:                     r,
		discovery:             discovery,
		resty:                 restyClient,
	}
	client.Router(r)

	p2pClient, err := client.InitP2P()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to start the client with host: %s, port: %d", cfg.Host, cfg.Port))
	}
	client.client = p2pClient

	log.Infof("Init P2P Client successfully, host: %s, port: %d", cfg.Host, cfg.Port)

	// Discovery other node
	peers, err := client.RegisterDiscovery(ctx, p2pClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to register the peer to the discovery")
	}
	client.peers = peers

	return client, nil
}

func (c *Client) InitP2P() (*Peer, error) {
	address := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	p := &Peer{
		Address: address,
		ID:      uuid.New().String(),
	}
	go func() {
		err := c.r.Run(address)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return p, nil
}

func (c *Client) RegisterDiscovery(ctx context.Context, peer *Peer) ([]*Peer, error) {
	resp, err := c.resty.R().SetBody(map[string]interface{}{
		"peer": peer,
	}).Post(fmt.Sprintf("http://%s/register-peer", c.discovery.Address))
	if err != nil {
		return nil, err
	}
	var response struct {
		Peers []*Peer `json:"peers"`
	}
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return nil, err
	}
	return response.Peers, nil
}

func (c *Client) GetBlockData(ctx context.Context, peer *Peer, req model.GetBlockDataByIndexRequest) ([]byte, error) {
	resp, err := c.resty.R().SetBody(req).Post(fmt.Sprintf("http://%s/get-data-by-index", peer.Address))
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil

}

func (c *Client) Peers() ([]*Peer, error) {
	resp, err := c.resty.R().Get(fmt.Sprintf("http://%s/peers", c.discovery.Address))
	if err != nil {
		return nil, err
	}
	var response struct {
		Peers []*Peer `json:"peers"`
	}
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return nil, err
	}
	c.peers = response.Peers
	return response.Peers, nil
}
