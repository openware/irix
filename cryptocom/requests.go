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
		Params: kvParams{"channels": channels},
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
		Params: kvParams{"channels": channels},
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
	return &Request{
		Id:     reqID,
		Method: privateCreateOrder,
		Params: kvParams{
			"instrument_name": strings.ToUpper(ask) + "_" + strings.ToUpper(bid),
			"side":            strings.ToUpper(orderSide),
			"type":            "LIMIT",
			"price":           price,
			"quantity":        volume,
			"client_oid":      uuid,
		},
		Nonce: generateNonce(),
	}
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
		Params: kvParams{
			"instrument_name": strings.ToUpper(ask) + "_" + strings.ToUpper(bid),
			"side":            strings.ToUpper(orderSide),
			"type":            "MARKET",
			volumeKey:         volume,
			"client_oid":      uuid,
		},
		Nonce: generateNonce(),
	}
}

func (c *Client) cancelOrder(reqID int, remoteID, market string) (req *Request, err error) {
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

	id := reqID
	nonce := generateNonce()
	if id == 0 {
		id = int(nonce)
	}
	req = &Request{
		Id:     id,
		Method: privateCancelOrder,
		Params: kvParams{
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
		Params: kvParams{
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
		Params: kvParams{
			"order_id": remoteID,
		},
		Nonce: nonce,
	}
	return
}

func (c *Client) restGetOrderDetailsRequest(reqID int, remoteID string) *Request {
	r := &Request{
		Id:     reqID,
		Method: privateGetOrderDetail,
		Params: kvParams{
			"order_id": remoteID,
		},
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
}

func (c *Client) restGetBalanceRequest(reqID int) *Request {
	r := &Request{
		Id:     reqID,
		Method: privateGetAccountSummary,
		Params: kvParams{},
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
}

func (c *Client) restGetTradesRequest(reqID int, market string) *Request {
	r := &Request{
		Id:     reqID,
		Method: privateGetTrades,
		Params: kvParams{
			"instrument_name": market,
		},
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
}

func (c *Client) restOpenOrdersRequest(reqID int, market string, page int, pageSize int) *Request {
	r := &Request{
		Id:     reqID,
		Method: privateGetOpenOrders,
		Params: kvParams{
			"instrument_name": market,
			"page":            strconv.Itoa(page),
			"page_size":       strconv.Itoa(pageSize),
		},
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
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
	params := kvParams{
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
	params := kvParams{
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
	params := kvParams{}
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
	params := kvParams{}
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
	if currency == "" {
		err = errors.New("invalid currency value")
		return
	}
	if err = isValidCurrency(currency); err != nil {
		return
	}
	params := kvParams{
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
	params := kvParams{}
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
		Params: kvParams{
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

func isValidCurrency(code string) (err error) {
	regex := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if code == "" || len(code) < 3 || !regex.MatchString(code) {
		err = errors.New("invalid code")
	}
	return
}
func (c *Client) createOrder(instrumentName string, side order.Side, orderType order.Type, price, quantity float64, orderOption *OrderOption) (req *Request, err error){

	if err = tryOrError(func() error {
		return validInstrument(instrumentName)
	}, func() (err error) {
		if side != order.Buy && side != order.Sell {
			err = errors.New("invalid order side")
		}
		return
	}, func() error {
		switch orderType {
		case order.Limit, order.StopLimit, order.Market, TakeProfitLimit, StopLoss, order.TakeProfit:
			return nil
		default:
			return errors.New("invalid order type")
		}
	}, func() (err error) {
		if orderType == order.Limit {
			if quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if price <= 0 {
				err = errors.New("price required")
				return
			}
			if orderOption != nil {
				if orderOption.ExecInst != "" && orderOption.ExecInst != PostOnly.String() {
					err = fmt.Errorf("exec_inst value not allowed. either leave it empty or set it to %s", PostOnly)
					return
				}
				if orderOption.TimeInForce != "" &&
					!(orderOption.TimeInForce == GoodTillCancel.String() || orderOption.TimeInForce == FillOrKill.String() || orderOption.TimeInForce == order.ImmediateOrCancel.String()) {
					err = fmt.Errorf("time_in_force value not allowed. either leave it empty or set it to %s, %s, or %s", GoodTillCancel, FillOrKill, order.ImmediateOrCancel)
					return
				}
			}
		}
		return
	}, func() (err error) {
		if orderType == order.Market {
			if side == order.Buy && (orderOption == nil || orderOption.Notional <= 0) {
				err = errors.New("notional required")
				return
			}
			if side == order.Sell && quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
		}
		return
	}, func() (err error) {
		if orderType == order.StopLimit || orderType == TakeProfitLimit {
			if price <= 0 {
				err = errors.New("price required")
				return
			}
			if quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if orderOption == nil || orderOption.TriggerPrice <= 0 {
				err = errors.New("trigger_price required")
				return
			}
		}
		return
	}, func() (err error) {
		if orderType == StopLoss || orderType == order.TakeProfit {
			if side == order.Buy && (orderOption == nil || orderOption.Notional <= 0) {
				err = errors.New("notional required")
				return
			}
			if side == order.Sell && quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if orderOption == nil || orderOption.TriggerPrice <= 0 {
				err = errors.New("trigger_price required")
				return
			}
		}
		return
	}); err != nil {
		return
	}
	// validate cases based on the requirements

	nonce := generateNonce()
	params := kvParams{
		"instrument_name": instrumentName,
		"side": side.String(),
		"type": strings.ReplaceAll(orderType.String(), " ", "-"),
		"price": price,
		"quantity": quantity,
	}
	if orderOption != nil {
		if orderOption.Notional > 0 {
			params["notional"] = orderOption.Notional
		}
		if orderOption.TriggerPrice > 0 {
			params["notional"] = orderOption.TriggerPrice
		}
		// set params only if order type is order.Limit
		if orderOption.TimeInForce != "" && orderType == order.Limit {
			params["time_in_force"] = orderOption.TimeInForce
		}
		if orderOption.ExecInst != "" && orderType == order.Limit {
			params["exec_inst"] = orderOption.ExecInst
		}
		if orderOption.ClientOrderID != "" {
			params["client_oid"] = orderOption.ClientOrderID
		}
	}
	req = &Request{
		Id:        int(nonce),
		Method:    privateCreateOrder,
		Nonce:     nonce,
		Params:    params,
	}
	return
}

func (c *Client) getOpenOrders(market string, pageSize, page int) (req *Request, err error) {
	if err = tryOrError(func() error {
		if market == "" {
			return nil
		}
		return validInstrument(market)
	}, func() error {
		if pageSize < 0 {
			return errors.New("invalid page size value. minimum is 0")
		}
		return nil
	}, func() error {
		if page < 0 {
			return errors.New("invalid page value. minimum is 0")
		}
		return nil
	}); err != nil {
		return
	}
	if pageSize == 0 {
		pageSize = 20
	}
	params := kvParams{
		"page_size": pageSize,
		"page": page,
	}
	if market != "" {
		params["instrument_name"] = market
	}
	req = &Request{
		Method: privateGetOpenOrders,
		Params: params,
	}
	return
}