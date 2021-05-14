package cryptocom

import (
	"encoding/json"
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
	Nonce     int64
	Params    kvParams
}

type Response struct {
	Id      int
	Method  string
	Code    int
	Message string
	Result  map[string]interface{}
}

type RawResponse struct {
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Code    int    `json:"code"`
	Message string `json:"message"`
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
	Open   int   `json:"o"`
	High   int   `json:"h"`
	Low    int   `json:"l"`
	Close  int   `json:"c"`
	Volume int   `json:"v"`
}
type TickerResponse struct {
	Result TickerResult `json:"result"`
}
type TickerResult struct {
	Data []Ticker `json:"data"`
}
type Ticker struct {
	Instrument string  `json:"i"`
	Bid        int     `json:"b"`
	Ask        int     `json:"k"`
	Trade      float64 `json:"a"`
	Timestamp  int64   `json:"t"`
	Volume     int     `json:"v"`
	Highest    float64 `json:"h"`
	Lowest     int     `json:"l"`
	Change     float64 `json:"c"`
}
type PublicTradeResponse struct {
	Result PublicTradeResult `json:"result"`
}
type PublicTradeResult struct {
	Data []PublicTrade `json:"data"`
}
type PublicTrade struct {
	Instrument string  `json:"i"`
	Quantity   int     `json:"q"`
	Price      float64 `json:"p"`
	Side       string  `json:"s"`
	Timestamp  int64   `json:"t"`
	TradeID    int     `json:"d"`
}
type DepositAddressResponse struct {
	Result DepositAddressResult `json:"result"`
}
type DepositAddressResult struct {
	DepositAddressList []DepositAddress `json:"deposit_address_list"`
}
type DepositAddress struct {
	Currency   string `json:"currency"`
	CreateTime int64  `json:"create_time"`
	ID         string `json:"id"`
	Address    string `json:"address"`
	Status     string `json:"status"`
	Network    string `json:"network"`
}

type AccountResponse struct {
	Result AccountResult `json:"result"`
}
type AccountResult struct {
	Accounts []AccountSummary `json:"accounts"`
}
type AccountSummary struct {
	Balance   float64 `json:"balance"`
	Available float64 `json:"available"`
	Order     float64 `json:"order"`
	Stake     int     `json:"stake"`
	Currency  string  `json:"currency"`
}

type ScopeResult struct {
	Result Scope `json:"result"`
}
type Scope struct {
	Scope string `json:"scope"`
}
type OrderOption struct {
	Notional      float64
	ClientOrderID string
	TimeInForce   string
	ExecInst      string
	TriggerPrice  float64
}

type OrderResult struct {
	Result Order `json:"result"`
}
type Order struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_oid"`
}
type OrderDetailResponse struct {
	Result OrderDetail `json:"result"`
}
type OrderDetail struct {
	TradeList []Trade `json:"trade_list"`
	OrderInfo Order `json:"order_info"`
}
type Trade struct {
	Side           string  `json:"side"`
	InstrumentName string  `json:"instrument_name"`
	Fee            float64 `json:"fee"`
	TradeID        string  `json:"trade_id"`
	CreateTime     int64   `json:"create_time"`
	TradedPrice    int     `json:"traded_price"`
	TradedQuantity int     `json:"traded_quantity"`
	FeeCurrency    string  `json:"fee_currency"`
	OrderID        string  `json:"order_id"`
}
type OrderInfo struct {
	Status             string `json:"status"`
	Side               string `json:"side"`
	OrderID            string `json:"order_id"`
	ClientOid          string `json:"client_oid"`
	CreateTime         int64  `json:"create_time"`
	UpdateTime         int64  `json:"update_time"`
	Type               string `json:"type"`
	InstrumentName     string `json:"instrument_name"`
	CumulativeQuantity int    `json:"cumulative_quantity"`
	CumulativeValue    int    `json:"cumulative_value"`
	AvgPrice           int    `json:"avg_price"`
	FeeCurrency        string `json:"fee_currency"`
	TimeInForce        string `json:"time_in_force"`
	ExecInst           string `json:"exec_inst"`
}
type kvParams map[string]interface{}

func generateNonce() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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
