package client

import "github.com/gorilla/websocket"

type Block struct {
	Hash      string  `json:"hash"`
	Height    int     `json:"height"`
	NTx       int     `json:"nTx"`
	Size      int     `json:"size"`
	Time      int     `json:"time"`
	Weight    int     `json:"weight"`
	Fees      float64 `json:"fees"`
	MinFee    float64 `json:"minFee"`
	MaxFee    float64 `json:"maxFee"`
	MedianFee float64 `json:"medianFee"`
}

type ProjectedBlock struct {
	BlockSize    int     `json:"blockSize"`
	BlockWeight  int     `json:"blockWeight"`
	NTx          int     `json:"nTx"`
	MinFee       float64 `json:"minFee"`
	MaxFee       float64 `json:"maxFee"`
	MinWeigthFee float64 `json:"minWeigthFee"`
	MaxWeigthFee float64 `json:"maxWeigthFee"`
	MedianFee    float64 `json:"medianFee"`
	Fees         float64 `json:"fees"`
	HasMyTx      bool    `json:"hasMytx"`
}

type Response struct {
	MempoolInfo struct {
		Size  int `json:"size"`
		Bytes int `json:"bytes"`
	} `json:"mempoolInfo"`

	Block  Block   `json:"block"`
	Blocks []Block `json:"blocks"`

	ProjectedBlocks []ProjectedBlock `json:"projectedBlocks"`

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

	if err := conn.WriteMessage(websocket.TextMessage, []byte(
		`{"action":"want","data":["stats","blocks","projected-blocks"]}`,
	)); err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Read() (*Response, error) {
	var resp Response
	if err := c.conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
