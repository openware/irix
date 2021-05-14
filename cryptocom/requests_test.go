package cryptocom

import (
	"fmt"
	"github.com/openware/pkg/order"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	cl = Default("test", "test", true)
)

func TestRequest_AuthRequest(t *testing.T)  {
	t.Parallel()
	r := cl.authRequest()
	assert.NotEmpty(t,  r.Id)
	assert.Equal(t, publicAuth, r.Method)
	assert.NotEmpty(t, r.Nonce)
	assert.NotEmpty(t, r.Signature)
	assert.Equal(t, "test", r.ApiKey)
}
func TestRequest_GetInstruments(t *testing.T)  {
	t.Parallel()
	r := cl.getInstruments()
	assert.Equal(t, 1, r.Id)
	assert.Equal(t, publicGetInstruments, r.Method)
	assert.NotNil(t, r.Nonce)
}
func TestRequest_GetBook(t *testing.T)  {
	t.Parallel()
	type args struct {
		instrument string
		depth int
		shouldError bool
	}
	inputs := []args{
		{"", 0, true},
		{"_", 0, true},
		{"BTC_", 0, true},
		{"_USDT", 0, true},
		{"BTC_USDT", 151, true},
		{"BTC_USDT", -1, true},
		{"BTC_USDT", 10, false},
		{"BTC_USDT", 150, false},
		{"BTC_USDT", 0, false},
	}

	for _, arg := range inputs {
		r, err := cl.getOrderBook(1, arg.instrument, arg.depth)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, 1, r.Id)
			assert.Equal(t, publicGetBook, r.Method)
			assert.Equal(t, arg.instrument, r.Params["instrument_name"])
			if arg.depth > 0 {
				assert.Equal(t, strconv.Itoa(arg.depth), r.Params["depth"])
			} else {
				assert.Equal(t, strconv.Itoa(150), r.Params["depth"])
			}
			assert.NotNil(t, r.Nonce)
		}
	}
}

func TestRequest_GetCandlestick(t *testing.T)  {
	t.Parallel()
	type args struct {
		instrument  string
		period      Interval
		depth       int
		shouldError bool
	}
	inputs := []args{
		// invalid inputs
		{"", Minute1, 1, true},
		{"_", Minute5, 1, true},
		{"BTC_", Minute1, 1, true},
		{"_USDT", Minute1, 1, true},
		{"BTC_USDT", -10, 1, true},
		{"BTC_USDT", 100, 1, true},
		// valid inputs
		{"BTC_USDT", Minute5, 10, false},
		{"BTC_USDT", Hour1, 100, false},
		{"BTC_USDT", Week, 0, false},
	}

	for _, arg := range inputs {
		r, err := cl.getCandlestick(arg.instrument, arg.period, arg.depth)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, publicGetCandlestick, r.Method, arg)
			assert.Equal(t, arg.instrument, r.Params["instrument_name"], arg)
			assert.Equal(t, arg.period.Encode(), r.Params["interval"], arg)
			if arg.depth > 0 {
				assert.Equal(t, arg.depth, r.Params["depth"])
			}
		}
	}
}

func TestGetTicker(t *testing.T)  {
	testTable := []struct{
		instrumentName string
		shouldError bool
	}{
		{"_", true},
		{"BTC_", true},
		{"_USDT", true},
		// valid inputs
		{"", false},
		{"BTC_USDT", false},
		{"BTC_USDT", false},
		{"BTC_USDT", false},
	}
	for _, arg := range testTable {
		r, err := cl.getTicker(arg.instrumentName)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, publicGetTicker, r.Method, arg)
			if arg.instrumentName != "" {
				assert.Equal(t, arg.instrumentName, r.Params["instrument_name"], arg)
			} else {
				assert.Nil(t, r.Params["instrument_name"])
			}
		}
	}
}

func TestGetTrades(t *testing.T)  {
	testTable := []struct{
		instrumentName string
		shouldError bool
	}{
		{"_", true},
		{"BTC_", true},
		{"_USDT", true},
		// valid inputs
		{"", false},
		{"BTC_USDT", false},
		{"BTC_USDT", false},
		{"BTC_USDT", false},
	}
	for _, arg := range testTable {
		r, err := cl.getPublicTrades(arg.instrumentName)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, publicGetTrades, r.Method, arg)
			if arg.instrumentName != "" {
				assert.Equal(t, arg.instrumentName, r.Params["instrument_name"], arg)
			} else {
				assert.Nil(t, r.Params["instrument_name"])
			}
		}
	}
}

func TestGetDepositAddress(t *testing.T)  {
	testTable := []struct{
		instrumentName string
		shouldError bool
	}{
		{"_", true},
		{"BTC_", true},
		{"_USDT", true},
		{"BTC_USDT", true},
		{"", true},
		// valid inputs
		{"BTC", false},
		{"USDT", false},
	}
	for _, arg := range testTable {
		r, err := cl.getDepositAddress(arg.instrumentName)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, privateGetDepositAddress, r.Method, arg)
			if arg.instrumentName != "" {
				assert.Equal(t, arg.instrumentName, r.Params["currency"], arg)
			} else {
				assert.Nil(t, r.Params["currency"])
			}
		}
	}
}
func TestGetAccountSummary(t *testing.T)  {
	t.Parallel()
	testTable := []struct{
		instrumentName string
		shouldError bool
	}{
		{"_", true},
		{"BTC_", true},
		{"_USDT", true},
		{"BTC_USDT", true},
		// valid inputs
		{"", false},
		{"BTC", false},
		{"USDT", false},
	}
	for _, arg := range testTable {
		r, err := cl.getAccountSummary(arg.instrumentName)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, privateGetAccountSummary, r.Method, arg)
			assert.NotEmpty(t, r.ApiKey)
			assert.NotEmpty(t, r.Signature)
			if arg.instrumentName != "" {
				assert.Equal(t, arg.instrumentName, r.Params["currency"], arg)
			} else {
				assert.Nil(t, r.Params["currency"])
			}
		}
	}
}
func TestGetDepositHistory(t *testing.T)  {
	testTable := []struct{
		instrumentName string
		shouldError bool
	}{
		{"_", true},
		{"BTC_", true},
		{"_USDT", true},
		{"BTC_USDT", true},
		{"", true},
		// valid inputs
		{"BTC", false},
		{"USDT", false},
	}
	for _, arg := range testTable {
		r, err := cl.getDepositAddress(arg.instrumentName)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, privateGetDepositAddress, r.Method, arg)
			if arg.instrumentName != "" {
				assert.Equal(t, arg.instrumentName, r.Params["currency"], arg)
			} else {
				assert.Nil(t, r.Params["currency"])
			}
		}
	}
}

func TestRespondHeartbeat(t *testing.T) {
	t.Parallel()
	testTable := []struct{
		id int
		shouldError bool
	}{
		{0, true},
		{ -1, true},
		{1, false},
		{int(time.Now().UnixNano()/int64(time.Millisecond)), false},
	}
	for _, c := range testTable {
		res, err := cl.heartbeat(c.id)
		if c.shouldError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, publicRespondHeartbeat, res.Method)
			assert.Equal(t, c.id, res.Id)
		}
	}
}
func TestSetCancelOnDisconnect(t *testing.T) {
	t.Parallel()
	testTable := []struct{
		scope string
		shouldError bool
	}{
		{"", true},
		{"random", true},
		{"-", true},
		// valid inputs
		{ScopeAccount, false},
		{ScopeConnection, false},
	}
	for _, arg := range testTable {
		r, err := cl.setCancelOnDisconnect(arg.scope)
		if arg.shouldError {
			assert.NotNil(t, err, arg)
			assert.Nil(t, r, arg)
		} else {
			assert.Nil(t, err, arg)
			assert.Equal(t, privateSetCancelOnDisconnect, r.Method, arg)
			assert.NotEmpty(t, r.ApiKey)
			assert.NotEmpty(t, r.Signature)
			assert.Equal(t, arg.scope, r.Params["scope"])
		}
	}
}

func TestGetCancelOnDisconnect(t *testing.T) {
	t.Parallel()
	req, _ := cl.getCancelOnDisconnect()
	assert.Equal(t, privateGetCancelOnDisconnect, req.Method)
	assert.NotEmpty(t, req.Signature)
	assert.NotEmpty(t, req.Nonce)
}
func TestSubscribeChannel(t *testing.T) {
	t.Parallel()
	testTable := []struct{
		channel []string
		shouldError bool
	}{
		{[]string{""}, true},
		{[]string{"-"}, true},
		{[]string{"cde"}, true},
		{[]string{"user.balance", "-"}, true},
		{[]string{"user.something", "-"}, true},
		// valid inputs
		{[]string{"user.balance", "user.order.ETC_USDT"}, false},
		{[]string{"user.margin.order.BTC_USDT", "user.margin.trade.ETC_USDT"}, false},
		{[]string{"book.ETC_USDT.10", "trade.ETC_USDT"}, false},
		{[]string{"candlestick.1d.ETC_USDT"}, false},
	}
	for _, c := range testTable {
		req, err := cl.subscribe(c.channel)
		if c.shouldError {
			assert.Nil(t, req)
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, req, c)
			assert.Equal(t, subscribe, req.Method, c)
			assert.Equal(t, c.channel, req.Params["channels"], c)
		}
	}
}
func TestUnsubscribeChannel(t *testing.T) {
	t.Parallel()
	testTable := []struct{
		channel []string
		shouldError bool
	}{
		{[]string{""}, true},
		{[]string{"-"}, true},
		{[]string{"cde"}, true},
		{[]string{"user.balance", "-"}, true},
		{[]string{"user.something", "-"}, true},
		// valid inputs
		{[]string{"user.balance", "user.order.ETC_USDT"}, false},
		{[]string{"user.margin.order.BTC_USDT", "user.margin.trade.ETC_USDT"}, false},
		{[]string{"book.ETC_USDT.10", "trade.ETC_USDT"}, false},
		{[]string{"candlestick.1d.ETC_USDT"}, false},
	}
	for _, c := range testTable {
		req, err := cl.unsubscribe(c.channel)
		if c.shouldError {
			assert.Nil(t, req, c)
			assert.NotNil(t, err, c)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, req, c)
			assert.Equal(t, unsubscribe, req.Method, c)
			assert.Equal(t, c.channel, req.Params["channels"], c)
		}
	}
}

func TestClient_CreateOrder(t *testing.T)  {
	t.Parallel()
	testTable := []struct{
		instrumentName string
		side order.Side
		orderType order.Type
		price float64
		quantity float64
		orderOption *OrderOption
		shouldError bool
	}{
		{"random", order.AnySide, order.AnyType, 0, 0, nil, true},
		{"BTC_USDT", order.AnySide, order.AnyType, 0, 0,  nil, true},
		{"BTC_USDT", order.Sell, order.AnyType, -1, 0, nil, true},
		{"BTC_USDT", order.Sell, order.AnyType, 0.001, 0, nil, true},
		{"BTC_USDT", order.Buy, order.AnyType, 0.001, 0.0001, nil, true},
		// edge cases but should fail
		{"BTC_USDT", order.Buy, order.Limit, 0, 0.0001, nil, true},
		{"BTC_USDT", order.Sell, order.Limit, 0.001, 0, nil, true},
		{"BTC_USDT", order.Buy, order.Market, 0, 0, nil, true},
		{"BTC_USDT", order.Buy, order.Market, 0, 0, &OrderOption{Notional: 0}, true},
		{"BTC_USDT", order.Sell, order.Market, 0, 0, nil, true},
		{"BTC_USDT", order.Sell, order.StopLimit, 0, 0, nil, true},
		{"BTC_USDT", order.Sell, order.StopLimit, 0.1, 0, nil, true},
		{"BTC_USDT", order.Buy, order.StopLimit, 0.1, 0.1, nil, true},
		{"BTC_USDT", order.Buy, order.StopLimit, 0.1, 0.1, &OrderOption{}, true},
		// TODO: switch to constant when the PR is merged
		{"BTC_USDT", order.Sell, TakeProfitLimit, 0, 0, nil, true},
		{"BTC_USDT", order.Sell, TakeProfitLimit, 0.1, 0, nil, true},
		{"BTC_USDT", order.Buy, TakeProfitLimit, 0.1, 0.1, nil, true},
		{"BTC_USDT", order.Buy, TakeProfitLimit, 0.1, 0.1, &OrderOption{}, true},
		// TODO: make PR to pkg
		{"BTC_USDT", order.Buy, StopLoss, 0.1, 0.1, nil, true},
		{"BTC_USDT", order.Buy, StopLoss, 0.1, 0.1, &OrderOption{Notional: 0, TriggerPrice: 0}, true},
		{"BTC_USDT", order.Buy, StopLoss, 0.1, 0.1, &OrderOption{Notional: 0, TriggerPrice: 0.1}, true},
		{"BTC_USDT", order.Buy, StopLoss, 0.1, 0.1, &OrderOption{Notional: 0.1, TriggerPrice: 0}, true},
		{"BTC_USDT", order.Sell, StopLoss, 0.1, 0.1, nil, true},
		{"BTC_USDT", order.Sell, StopLoss, 0.1, 0.1, &OrderOption{Notional: 0, TriggerPrice: 0}, true},
		{"BTC_USDT", order.Sell, StopLoss, 0.1, 0, &OrderOption{Notional: 0, TriggerPrice: 0.1}, true},
		{"BTC_USDT", order.Sell, StopLoss, 0.1, 0.1, &OrderOption{Notional: 0.1, TriggerPrice: 0}, true},
		// valid cases
		{"BTC_USDT", order.Buy, order.Limit, 0.001, 0.0001, nil, false},
		{"BTC_USDT", order.Buy, order.Limit, 0.001, 0.0001, &OrderOption{Notional: 0.0001}, false},
	}
	for _, c := range testTable {
		req, err := cl.createOrder(c.instrumentName, c.side, c.orderType, c.price, c.quantity, c.orderOption)
		if c.shouldError {
			assert.NotNil(t, err, c)
			assert.Nil(t, req, c)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, req, c)
			assert.Equal(t, privateCreateOrder, req.Method, c)
			assert.Equal(t, c.instrumentName, req.Params["instrument_name"])
			assert.Equal(t, c.side.String(), req.Params["side"])
			assert.Equal(t, strings.ReplaceAll(c.orderType.String(), " ", "-"), req.Params["type"])
			assert.Equal(t, c.quantity, req.Params["quantity"])
			assert.Equal(t, c.price, req.Params["price"])
			if c.orderOption != nil {
				if c.orderOption.Notional > 0 {
					assert.Equal(t, c.orderOption.Notional, req.Params["notional"])
				}
				if c.orderOption.TriggerPrice > 0 {
					assert.Equal(t, c.orderOption.TriggerPrice, req.Params["trigger_price"])
				}
				// fully optional
				if c.orderOption.ClientOrderID != "" {
					assert.Equal(t, c.orderOption.ClientOrderID, req.Params["client_oid"])
				}
				if c.orderOption.TimeInForce != "" {
					assert.Equal(t, c.orderOption.TriggerPrice, req.Params["time_in_force"])
				}
				if c.orderOption.ExecInst != "" {
					assert.Equal(t, c.orderOption.ClientOrderID, req.Params["exec_inst"])
				}
			}
		}
	}
}
func TestClient_CancelOrder(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		instrumentName string
		orderId string
		reqID int
		shouldError bool
	}{
		{"random", "", 0, true},
		{"-", "-", 0, true},
		{"-", "-", 0, true},
		{"BTC_USDT", "", 0, true},
		// valid values
		{"BTC_USDT", fmt.Sprintf("%d", time.Now().Unix()), 0, false},
		{"BTC_USDT", fmt.Sprintf("%d", time.Now().Unix()), 1234, false},
	}

	for _, c := range testTable {
		req, err := cl.cancelOrder(c.reqID, c.orderId, c.instrumentName)
		if c.shouldError {
			assert.NotNil(t, err, c)
			assert.Nil(t, req, c)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, req, c)
			assert.Equal(t, privateCancelOrder, req.Method)
			assert.Equal(t, c.orderId, req.Params["order_id"])
			assert.Equal(t, c.instrumentName, req.Params["instrument_name"])
			if c.reqID > 0 {
				assert.Equal(t, c.reqID, req.Id)
			} else {
				assert.NotEmpty(t, req.Id)
			}
		}
	}
}

func TestClient_CancelAllOrders(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		instrumentName string
		reqID int
		shouldError bool
	}{
		{"random",  0, true},
		{"-",  0, true},
		{"-",  0, true},
		// valid values
		{"BTC_USDT",  0, false},
		{"BTC_USDT",  1234, false},
	}

	for _, c := range testTable {
		req, err := cl.cancelAllOrder(c.reqID, c.instrumentName)
		if c.shouldError {
			assert.NotNil(t, err, c)
			assert.Nil(t, req, c)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, req, c)
			assert.Equal(t, privateCancelAllOrders, req.Method)
			assert.Equal(t, c.instrumentName, req.Params["instrument_name"])
			if c.reqID > 0 {
				assert.Equal(t, c.reqID, req.Id)
			} else {
				assert.NotEmpty(t, req.Id)
			}
		}
	}
}
