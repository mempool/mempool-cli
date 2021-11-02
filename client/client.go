package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	API_URL = "https://mempool.space/api/v1/ws"
)

type MempoolInfo struct {
	Size  int `json:"size"`
	Bytes int `json:"bytes"`
}

type Block struct {
	ID        string    `json:"id"`
	Height    int       `json:"height"`
	TxCount   int       `json:"tx_count"`
	Size      int       `json:"size"`
	Time      int       `json:"timestamp"`
	Weight    int       `json:"weight"`
	FeeRange  []float64 `json:"feeRange"`
	MedianFee float64   `json:"medianFee"`
}

type MempoolBlock struct {
	BlockSize    int       `json:"blockSize"`
	BlockWeight  float64   `json:"blockVSize"`
	NTx          int       `json:"nTx"`
	MinWeigthFee float64   `json:"minWeigthFee"`
	MaxWeigthFee float64   `json:"maxWeigthFee"`
	MedianFee    float64   `json:"medianFee"`
	FeeRange     []float64 `json:"feeRange"`
	HasMyTx      bool      `json:"hasMytx"`
}

type TrackTx struct {
	Tracking    bool   `json:"tracking"`
	BlockHeight int    `json:"blockHeight"`
	Message     string `json:"message"`
	TX          struct {
		Status struct {
			Confirmed bool
		}
	} `json:"tx"`
}

type Response struct {
	MempoolInfo *MempoolInfo `json:"mempoolInfo"`

	Block  *Block  `json:"block"`
	Blocks []Block `json:"blocks"`

	MempoolBlocks []MempoolBlock `json:"mempool-blocks"`
	TrackTx       TrackTx        `json:"track-tx"`

	TxPerSecond     float64 `json:"txPerSecond"`
	VBytesPerSecond int     `json:"vBytesPerSecond"`
	Conversions     struct {
		BTC float64 `json:"BTC"`
		USD float64 `json:"USD"`
	} `json:"conversions"`
}

type Client struct {
	conn *websocket.Conn
}

func New() (*Client, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("wss://mempool.space/ws", nil)
	if err != nil {
		return nil, err
	}

	conn.WriteMessage(websocket.TextMessage, []byte(
		`{"action": "init"}`,
	))

	return &Client{conn: conn}, nil
}

func (c *Client) Read() (*Response, error) {
	var resp Response
	if err := c.conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Want() error {
	return c.conn.WriteMessage(websocket.TextMessage, []byte(
		`{"action":"want","data":["stats","blocks","mempool-blocks"]}`,
	))
}

func (c *Client) Track(txId string) error {
	return c.conn.WriteMessage(websocket.TextMessage, []byte(
		fmt.Sprintf(`{"action":"track-tx","txId":"%s"}`, txId),
	))
}

type Fees []struct {
	FPV float64 `json:"fpv"`
}

func (f Fees) Len() int           { return len(f) }
func (f Fees) Less(i, j int) bool { return f[i].FPV < f[j].FPV }
func (f Fees) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	},
}

func Get(ctx context.Context, path string, v interface{}) error {
	req, err := http.NewRequest("GET", API_URL+path, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if s := r.StatusCode; s != 200 {
		return fmt.Errorf("status %d", s)
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func GetMempoolFee(ctx context.Context, n int) (Fees, error) {
	var fees Fees
	if err := Get(ctx, fmt.Sprintf("transactions/mempool/%d", n), &fees); err != nil {
		return nil, err
	}
	return fees, nil
}

func GetBlockFee(ctx context.Context, n int) (Fees, error) {
	var fees Fees
	if err := Get(ctx, fmt.Sprintf("transactions/height/%d", n), &fees); err != nil {
		return nil, err
	}
	return fees, nil
}
