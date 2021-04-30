package cryptocom

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// Types
	AuthRequest = iota + 1
	SubscribeRequest
	HeartBeat
	OrderRequest
	RestOrderRequest
	RestBalanceRequest
	RestTradesRequest
	RestOpenOrdersRequest
)

type Request struct {
	Id        int
	Method    string
	ApiKey    string
	Signature string
	Nonce     string
	Params    map[string]interface{}
}

type Response struct {
	Id      int
	Method  string
	Code    int
	Message string
	Result  map[string]interface{}
}

type RawResponse struct {
	ID      int
	Method  string
	Code    int
	Message string
}

type InstrumentResponse struct {
	Result InstrumentResult `json:"result"`
}
type InstrumentResult struct {
	Instruments []Instruments `json:"instruments"`
}
type Instruments struct {
	InstrumentName       string `json:"instrument_name"`
	QuoteCurrency        string `json:"quote_currency"`
	BaseCurrency         string `json:"base_currency"`
	PriceDecimals        int    `json:"price_decimals"`
	QuantityDecimals     int    `json:"quantity_decimals"`
	MarginTradingEnabled bool   `json:"margin_trading_enabled"`
}

type OrderbookResponse struct {
	Result OrderbookResult `json:"result"`
}
type OrderbookResult struct {
	InstrumentName string          `json:"instrument_name"`
	Depth          int             `json:"depth"`
	Data           []OrderbookData `json:"data"`
}
type Orderbook [][]float64
type OrderbookData struct {
	Bids Orderbook `json:"bids"`
	Asks Orderbook `json:"asks"`
	T    int64     `json:"t"`
}

type CandlestickResponse struct {
	Result CandlestickResult `json:"result"`
}
type CandlestickResult struct {
	InstrumentName string        `json:"instrument_name"`
	Depth          int           `json:"depth"`
	Interval       string        `json:"interval"`
	Data           []Candlestick `json:"data"`
}

type Candlestick struct {
	Time   int64 `json:"t"`
	Open   int `json:"o"`
	High   int `json:"h"`
	Low    int `json:"l"`
	Close  int `json:"c"`
	Volume int `json:"v"`
}

func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().Unix()*1000)
}

func (r *Request) Encode() ([]byte, error) {
	if r.Method == publicAuth {
		return json.Marshal(map[string]interface{}{
			"id":      r.Id,
			"method":  r.Method,
			"api_key": r.ApiKey,
			"sig":     r.Signature,
			"nonce":   r.Nonce,
		})
	}
	if r.Method == publicRespondHeartbeat {
		return json.Marshal(map[string]interface{}{
			"id":     r.Id,
			"method": r.Method,
		})
	}
	if strings.Contains(r.Method, "private/") {
		return json.Marshal(map[string]interface{}{
			"id":      r.Id,
			"method":  r.Method,
			"params":  r.Params,
			"api_key": r.ApiKey,
			"sig":     r.Signature,
			"nonce":   r.Nonce,
		})
	}
	return json.Marshal(map[string]interface{}{
		"id":     r.Id,
		"method": r.Method,
		"params": r.Params,
		"nonce":  r.Nonce,
	})

}
