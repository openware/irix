package cryptocom

import (
	"errors"
	"fmt"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/order"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (c *Client) authRequest() *Request {
	nonce := generateNonce()
	r := &Request{
		Id:     int(nonce),
		Method: publicAuth,
		ApiKey: c.key,
		Nonce:  nonce,
	}

	c.generateSignature(r)
	return r
}

func validChannel(channel string) error {
	allowedFormats := []string{
		"user.order.",
		"user.trade.",
		"user.balance",
		"user.margin.balance",
		"user.margin.order.",
		"user.margin.trade.",
		"book.",
		"ticker.",
		"trade.",
		"candlestick.",
	}
	err := errors.New("invalid format")
	for _, f := range allowedFormats {
		idx := strings.Index(channel, f)
		if idx >= 0 {
			return nil
		}
	}
	return err

}
func (c *Client) subscribe(channels []string) (req *Request, err error) {
	for _, ch := range channels {
		if err = validChannel(ch); err != nil {
			return
		}
	}
	nonce := generateNonce()
	req = &Request{
		Id:     int(nonce),
		Method: subscribe,
		Params: KVParams{"channels": channels},
		Nonce:  nonce,
	}
	return
}
func (c *Client) unsubscribe(channels []string) (req *Request, err error) {
	for _, ch := range channels {
		if err = validChannel(ch); err != nil {
			return
		}
	}
	nonce := generateNonce()
	req = &Request{
		Id:     int(nonce),
		Method: unsubscribe,
		Params: KVParams{"channels": channels},
		Nonce:  nonce,
	}
	return
}

func (c *Client) heartbeat(reqId int) (*Request, error) {
	if reqId <= 0 {
		return nil, errors.New("invalid id")
	}
	return &Request{
		Id:     reqId,
		Method: publicRespondHeartbeat,
	}, nil
}

func (c *Client) createOrderLimitRequest(
	reqID int,
	ask string,
	bid string,
	orderSide string,
	price decimal.Decimal,
	volume decimal.Decimal,
	uuid uuid.UUID) *Request {
	side := order.Sell
	if strings.ToUpper(orderSide) == "BUY" {
		side = order.Buy
	}
	priceF, _ := price.Float64()
	qtyF, _ := volume.Float64()
	o, _ := c.createOrder(reqID, CreateOrderParam{
		Market:        fmt.Sprintf("%s_%s", ask, bid),
		Side:           side,
		OrderType:     order.Limit,
		Price:         priceF,
		Quantity:      qtyF,
		ClientOrderID: uuid.String(),
	})
	return o
}

func (c *Client) createOrderMarketRequest(
	reqID int,
	ask string,
	bid string,
	orderSide string,
	volume decimal.Decimal,
	uuid uuid.UUID) *Request {
	var volumeKey string
	if strings.ToUpper(orderSide) == "BUY" {
		volumeKey = "notional"
	} else {
		volumeKey = "quantity"
	}

	return &Request{
		Id:     reqID,
		Method: privateCreateOrder,
		Params: KVParams{
			"instrument_name": strings.ToUpper(ask) + "_" + strings.ToUpper(bid),
			"side":            strings.ToUpper(orderSide),
			"type":            "MARKET",
			volumeKey:         volume,
			"client_oid":      uuid,
		},
		Nonce: generateNonce(),
	}
}

func (c *Client) cancelOrder(reqID int, market, remoteID string) (req *Request, err error) {
	if err = tryOrError(func() error {
		return validInstrument(market)
	}, func() error {
		if remoteID == "" {
			return errors.New("order id required")
		}
		return nil
	}); err != nil {
		return
	}

	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id:     reqID,
		Method: privateCancelOrder,
		Params: KVParams{
			"instrument_name": market,
			"order_id":        remoteID,
		},
		Nonce: nonce,
	}
	return
}

// Market: "ETH_BTC"
func (c *Client) cancelAllOrder(reqID int, market string) (req *Request, err error) {
	if err = validInstrument(market); err != nil {
		return
	}
	id := reqID
	nonce := generateNonce()
	if id == 0 {
		id = int(nonce)
	}
	req = &Request{
		Id:     id,
		Method: privateCancelAllOrders,
		Params: KVParams{
			"instrument_name": market,
		},
		Nonce: generateNonce(),
	}
	return
}

func (c *Client) getOrderDetail(reqID int, remoteID string) (req *Request, err error) {
	regex := regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9-_]+)$")
	if remoteID == "" || !regex.MatchString(remoteID) {
		err = errors.New("invalid order id")
		return
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id:     reqID,
		Method: privateGetOrderDetail,
		Params: KVParams{
			"order_id": remoteID,
		},
		Nonce: nonce,
	}
	return
}

func (c *Client) privateGetTrades(reqID int, params *TradeParams) (*Request, error) {
	pr, err := params.Encode()
	if err != nil {
		return nil, err
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	r := &Request{
		Id:     reqID,
		Method: privateGetTrades,
		Params: pr,
		Nonce:  nonce,
	}
	return r, nil
}
func (c *Client) privateGetOrderHistory(reqID int, params *TradeParams) (*Request, error) {
	pr, err := params.Encode()
	if err != nil {
		return nil, err
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	r := &Request{
		Id:     reqID,
		Method: privateGetOrderHistory,
		Params: pr,
		Nonce:  nonce,
	}
	return r, nil
}

func (c *Client) getInstruments() *Request {
	return &Request{
		Id:     1,
		Method: publicGetInstruments,
		Nonce:  generateNonce(),
	}
}

func (c *Client) getOrderBook(reqID int, instrument string, depth int) (req *Request, err error) {
	if err = validInstrument(instrument); err != nil {
		return
	}
	// max depth based on docs
	if depth < 0 || depth > 150 {
		err = errors.New("invalid depth value")
		return
	}
	params := KVParams{
		"instrument_name": instrument,
	}
	if depth == 0 {
		depth = 150
	}
	if depth > 0 {
		params["depth"] = strconv.Itoa(depth)
	}
	req = &Request{
		Id:     reqID,
		Method: publicGetBook,
		Nonce:  generateNonce(),
		Params: params,
	}
	return
}

func (c *Client) getCandlestick(instrumentName string, period Interval, depth int) (req *Request, err error) {
	if err = validInstrument(instrumentName); err != nil {
		return
	}
	if period < Minute1 || period > Month {
		err = errors.New("invalid interval")
		return
	}
	if depth < 0 || depth > 1000 {
		err = errors.New("invalid interval")
		return
	}
	params := KVParams{
		"instrument_name": instrumentName,
		"interval":        period.Encode(),
	}
	if depth > 0 {
		params["depth"] = depth
	}
	req = &Request{
		Method: publicGetCandlestick,
		Params: params,
	}
	return
}
func (c *Client) getTicker(instrumentName string) (req *Request, err error) {
	params := KVParams{}
	if instrumentName != "" {
		if err = validInstrument(instrumentName); err != nil {
			return
		}
		params["instrument_name"] = instrumentName
	}
	req = &Request{
		Method: publicGetTicker,
		Params: params,
	}
	return
}
func (c *Client) getPublicTrades(instrumentName string) (req *Request, err error) {
	params := KVParams{}
	if instrumentName != "" {
		if err = validInstrument(instrumentName); err != nil {
			return
		}
		params["instrument_name"] = instrumentName
	}
	req = &Request{
		Method: publicGetTrades,
		Params: params,
	}
	return
}
func (c *Client) getDepositAddress(currency string) (req *Request, err error) {
	if err = isValidCurrency(currency); err != nil {
		return
	}
	params := KVParams{
		"currency": currency,
	}
	nonce := generateNonce()
	req = &Request{
		Id:     int(nonce),
		Method: privateGetDepositAddress,
		Params: params,
		Nonce:  nonce,
	}
	return
}

func (c *Client) getAccountSummary(instrumentName string) (req *Request, err error) {
	params := KVParams{}
	if instrumentName != "" {
		// TODO: do small validation. ask the team how the validation done in the backend
		code := currency.NewCode(instrumentName).String()
		if err = isValidCurrency(code); err != nil {
			return
		}
		params["currency"] = code
	}
	req = &Request{
		Method: privateGetAccountSummary,
		Params: params,
	}
	return
}
func (c *Client) setCancelOnDisconnect(scope string) (req *Request, err error) {
	if scope != ScopeConnection && scope != ScopeAccount {
		return nil, errors.New("invalid scope value")
	}
	nonce := generateNonce()
	req = &Request{
		Id: int(nonce),
		Method: privateSetCancelOnDisconnect,
		Nonce: nonce,
		Params: KVParams{
			"scope": scope,
		},
	}
	return
}

func (c *Client) getCancelOnDisconnect() (req *Request, err error) {
	nonce := generateNonce()
	req = &Request{
		Id: int(nonce),
		Method: privateGetCancelOnDisconnect,
		Nonce: nonce,
	}
	return
}

func (c *Client) createOrder(reqID int, param CreateOrderParam) (req *Request, err error){
	params, err := param.Encode()
	if err != nil {
		return
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id:        reqID,
		Method:    privateCreateOrder,
		Nonce:     nonce,
		Params:    params,
	}
	return
}

func (c *Client) getOpenOrders(reqID int, param *OpenOrderParam) (req *Request, err error) {
	pr, err := param.Encode()
	if err != nil {
		return nil, err
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id: reqID,
		Method: privateGetOpenOrders,
		Params: pr,
		Nonce: nonce,
	}
	return
}
func (c *Client) createWithdrawal(reqID int, params WithdrawParams) (req *Request, err error) {
	pr, err := params.Encode()
	if err != nil {
		return
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id: reqID,
		Method: privateCreateWithdrawal,
		Params: pr,
		Nonce: nonce,
	}
	return
}
func (c *Client) getWithdrawalHistory(reqID int, params *WithdrawHistoryParam) (req *Request, err error) {
	pr, err := params.Encode()
	if err != nil {
		return
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id: reqID,
		Method: privateGetWithdrawalHistory,
		Params: pr,
		Nonce: nonce,
	}
	return
}
func (c *Client) getDepositHistory(reqID int, params *DepositHistoryParam) (req *Request, err error) {
	pr, err := params.Encode()
	if err != nil {
		return
	}
	nonce := generateNonce()
	if reqID == 0 {
		reqID = int(nonce)
	}
	req = &Request{
		Id: reqID,
		Method: privateGetDepositHistory,
		Params: pr,
		Nonce: nonce,
	}
	return
}