package cryptocom

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openware/irix/faker"
	"github.com/openware/pkg/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestPrivateUserOrder(t *testing.T) {
	testCases := []struct {
		instruments     []string
		validationError bool
		shouldError     bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePrivateOrders(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("user.order.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPrivateUserTrade(t *testing.T) {
	testCases := []struct {
		instruments     []string
		validationError bool
		shouldError     bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePrivateTrades(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("user.trade.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPrivateMarginOrder(t *testing.T) {
	testCases := []struct {
		instruments     []string
		validationError bool
		shouldError     bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePrivateMarginOrders(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("user.margin.order.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPrivateMarginTrade(t *testing.T) {
	testCases := []struct {
		instruments     []string
		validationError bool
		shouldError     bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePrivateMarginTrades(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("user.margin.trade.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPrivateUserBalance(t *testing.T) {
	testCases := []struct {
		shouldError bool
	}{
		{true},
		{false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		var err error
		if c.shouldError {
			err = errors.New("disconnected")
		}
		private.
			On("WriteMessage", mock.Anything, mock.Anything).
			Return(err)
		err = cli.SubscribePrivateBalanceUpdates()
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		assert.Equal(t, subscribe, pr.Method)
		params := pr.Params["channels"].([]interface{})
		assert.Equal(t, "user.balance", params[0])
	}
}
func TestPrivateMarginBalance(t *testing.T) {
	testCases := []struct {
		shouldError bool
	}{
		{true},
		{false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		var err error
		if c.shouldError {
			err = errors.New("disconnected")
		}
		private.
			On("WriteMessage", mock.Anything, mock.Anything).
			Return(err)
		err = cli.SubscribePrivateMarginBalanceUpdates()
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		assert.Equal(t, subscribe, pr.Method)
		params := pr.Params["channels"].([]interface{})
		assert.Equal(t, "user.margin.balance", params[0])
	}
}
func TestWsAuthenticate(t *testing.T) {
	cli, public, private := mockWsClient()
	private.
		On("WriteMessage", mock.Anything, mock.Anything).
		Return(nil)
	cli.authenticate()
	var pr Request
	private.AssertExpectations(t)
	private.AssertNumberOfCalls(t, "WriteMessage", 1)
	public.AssertNumberOfCalls(t, "WriteMessage", 0)
	req := private.Calls[0].Arguments[1].([]byte)
	_ = json.Unmarshal(req, &pr)
	assert.Equal(t, publicAuth, pr.Method)
	assert.NotEmpty(t, pr.Signature)
	assert.NotEmpty(t, pr.ApiKey)
	assert.Equal(t, "test", pr.ApiKey)
}
func TestWsSetCancelOnDisconnect(t *testing.T) {
	testCases := []struct {
		scope                 string
		shouldValidationError bool
		shouldError           bool
	}{
		{"", true, false},
		{"random", true, false},
		// valid cases
		{ScopeConnection, false, true},
		{ScopeAccount, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsSetCancelOnDisconnect(c.scope)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		_ = json.Unmarshal(req, &pr)
		assert.Equal(t, privateSetCancelOnDisconnect, pr.Method)
		assert.Equal(t, c.scope, pr.Params["scope"])
		assert.Empty(t, pr.Signature)
		assert.Empty(t, pr.ApiKey)
	}
}
func TestWsGetCancelOnDisconnect(t *testing.T) {
	cli, public, private := mockWsClient()
	private.
		On("WriteMessage", mock.Anything, mock.Anything).
		Return(nil)
	err := cli.WsGetCancelOnDisconnect()
	var pr map[string]interface{}
	assert.Nil(t, err)
	private.AssertExpectations(t)
	private.AssertNumberOfCalls(t, "WriteMessage", 1)
	public.AssertNumberOfCalls(t, "WriteMessage", 0)
	req := private.Calls[0].Arguments[1].([]byte)
	_ = json.Unmarshal(req, &pr)
	method, ok := pr["method"]
	assert.True(t, ok)
	assert.Equal(t, privateGetCancelOnDisconnect, method)
	_, ok = pr["sig"]
	assert.False(t, ok)
	_, ok = pr["params"]
	assert.False(t, ok)
	_, ok = pr["api_key"]
	assert.False(t, ok)
}

func TestWsCreateWithdrawal(t *testing.T) {
	method := privateCreateWithdrawal
	testCases := []struct {
		reqID                 int
		in                    WithdrawParams
		body                  *mockBody
		expectedParams        KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{0, WithdrawParams{}, nil, nil, true, false},
		{0, WithdrawParams{Currency: "BTC_USDT"}, nil, nil, true, false},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress}, &mockBody{400, mockResponseBody(-1, method, 10004, Withdraw{})}, KVParams{"currency": "BTC", "amount": "0.1", "address": faker.BitcoinTestAddress}, false, true},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress, WithdrawID: "withdrawid"}, &mockBody{400, mockResponseBody(-1, method, 10004, Withdraw{})}, KVParams{"currency": "BTC", "amount": "0.1", "address": faker.BitcoinTestAddress, "client_wid": "withdrawid"}, false, true},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress, WithdrawID: "withdrawid", AddressTag: "XRP"}, &mockBody{200, mockResponseBody(-1, method, 0, Withdraw{})}, KVParams{"currency": "BTC", "amount": "0.1", "address": faker.BitcoinTestAddress, "client_wid": "withdrawid", "address_tag": "XRP"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsCreateWithdrawal(c.reqID, c.in)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		assert.Equal(t, privateCreateWithdrawal, pr.Method)
		assert.Equal(t, c.expectedParams, pr.Params)

		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsGetWithdrawalHistory(t *testing.T) {
	method := privateGetWithdrawalHistory
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testCases := []struct {
		reqID                 int
		in                    *WithdrawHistoryParam
		body                  *mockBody
		expectedParams        KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{0, &WithdrawHistoryParam{StartTS: 2, EndTS: 1}, nil, nil, true, false},
		{0, &WithdrawHistoryParam{Currency: "BTC_USDT"}, nil, nil, true, false},
		{0, nil, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{}, false, true},
		{0, &WithdrawHistoryParam{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", StartTS: timeAgo[0]}, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0]}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0]}, &mockBody{500, mockResponseBody(-1, method, 10005, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1]}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1}, &mockBody{500, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1, Status: WithdrawCancelled}, &mockBody{200, mockResponseBody(-1, method, 0, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1, "status": "6"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsGetWithdrawalHistory(c.reqID, c.in)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(pr.Params)
		b2, _ := json.Marshal(c.expectedParams)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))

		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}

func TestWsGetAccountSummary(t *testing.T) {
	method := privateGetAccountSummary
	testCases := []struct {
		instrumentName        string
		shouldValidationError bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", true, false},
		{"BTC_",true, false},
		{"_", true, false},
		{"BTC_USDT", true, false},
		{"BTC_USDT", true, false},
		// invalid cases
		{"BTC_USDT", true, false},
		{"BTC_USDT", true, false},
		{"BTC", false, true},
		{"USDT", false, true},
		// valid cases
		{"",  false, false},
		{"USDT", false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsGetAccountSummary(c.instrumentName)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		assert.Equal(t, method, pr.Method)
		if c.instrumentName != "" {
			assert.Equal(t, c.instrumentName, pr.Params["currency"])
		} else {
			assert.Equal(t, KVParams{}, pr.Params)
		}


		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}

func TestWsCreateOrder(t *testing.T) {
	method := privateCreateOrder
	testCases := []struct {
		reqID                 int
		param                 CreateOrderParam
		expectedParams        KVParams
		body                  *mockBody
		shouldValidationError bool
		shouldError           bool
	}{
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Sell, OrderType: order.Limit, Price: 0.1, Quantity: 0.1, ExecInst: GoodTillCancel}, nil, nil, true, false},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Sell, OrderType: order.Limit, Price: 0.1, Quantity: 0.1, TimeInForce: PostOnly}, nil, nil, true, false},
		{0, CreateOrderParam{}, nil, nil, true, false},
		// valid cases
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Buy, OrderType: order.Limit, Price: 0.001, Quantity: 0.0001}, KVParams{"instrument_name": "BTC_USDT", "side": "BUY", "type": "LIMIT", "price": "0.001", "quantity": "0.0001"}, &mockBody{400, mockResponseBody(-1, method, 10004, Order{})}, false, true},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Sell, OrderType: order.Limit, Price: 0.001, Quantity: 0.0001, Notional: 0.0001}, KVParams{"instrument_name": "BTC_USDT", "side": "SELL", "type": "LIMIT", "price": "0.001", "quantity": "0.0001", "notional": "0.0001"}, &mockBody{400, mockResponseBody(-1, method, 10004, Order{})}, false, true},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Buy, OrderType: order.Limit, Price: 0.001, Quantity: 0.0001, Notional: 0.0001, TimeInForce: GoodTillCancel}, KVParams{"instrument_name": "BTC_USDT", "side": "BUY", "type": "LIMIT", "price": "0.001", "quantity": "0.0001", "notional": "0.0001", "time_in_force": "GOOD_TILL_CANCEL"}, &mockBody{400, mockResponseBody(-1, method, 10004, Order{})}, false, true},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Sell, OrderType: order.Limit, Price: 0.001, Quantity: 0.0001, Notional: 0.0001, TimeInForce: GoodTillCancel, ExecInst: PostOnly}, KVParams{"instrument_name": "BTC_USDT", "side": "SELL", "type": "LIMIT", "price": "0.001", "quantity": "0.0001", "notional": "0.0001", "time_in_force": "GOOD_TILL_CANCEL", "exec_inst": "POST_ONLY"}, &mockBody{400, mockResponseBody(-1, method, 10004, Order{})}, false, true},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Buy, OrderType: StopLoss, Price: 0.001, Quantity: 0.0001, Notional: 0.0001, TriggerPrice: 0.001}, KVParams{"instrument_name": "BTC_USDT", "side": "BUY", "type": "STOP_LOSS", "price": "0.001", "quantity": "0.0001", "notional": "0.0001", "trigger_price": "0.001"}, &mockBody{200, mockResponseBody(-1, method, 0, Order{"121212121212", ""})}, false, false},
		{0, CreateOrderParam{Market: "BTC_USDT", Side: order.Buy, OrderType: StopLoss, Price: 0.001, Quantity: 0.0001, Notional: 0.0001, TriggerPrice: 0.001, ClientOrderID: "someorderid"}, KVParams{"instrument_name": "BTC_USDT", "side": "BUY", "type": "STOP_LOSS", "price": "0.001", "quantity": "0.0001", "notional": "0.0001", "trigger_price": "0.001", "client_oid": "someorderid"}, &mockBody{200, mockResponseBody(-1, method, 0, Order{"121212121212", "someorderid"})}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsCreateOrder(c.reqID, c.param)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsCancelOrder(t *testing.T) {
	method := privateCancelOrder
	type input struct {
		market, orderID string
	}
	testCases := []struct {
		reqID                 int
		param                 input
		expectedParams        KVParams
		body                  *mockBody
		shouldValidationError bool
		shouldError           bool
	}{
		{0, input{"", ""}, nil, nil, true, false},
		{0, input{"BTC_USDT", ""}, nil, nil, true, false},
		// valid values
		{0, input{"BTC_USDT", "1212121212"}, KVParams{"instrument_name": "BTC_USDT", "order_id": "1212121212"}, &mockBody{400, mockResponseBody(-1, method, 10004, nil)}, false, true},
		{0, input{"BTC_USDT", "1212121212"}, KVParams{"instrument_name": "BTC_USDT", "order_id": "1212121212"}, &mockBody{200, mockResponseBody(-1, method, 0, nil)}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.CancelOrder(c.reqID, c.param.orderID, c.param.market)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsCancelAllOrder(t *testing.T) {
	method := privateCancelAllOrders
	type input struct {
		market string
	}
	testCases := []struct {
		reqID                 int
		param                 input
		expectedParams        KVParams
		body                  *mockBody
		shouldValidationError bool
		shouldError           bool
	}{
		{0, input{""}, nil, nil, true, false},
		// valid values
		{0, input{"BTC_USDT"}, KVParams{"instrument_name": "BTC_USDT"}, &mockBody{200, mockResponseBody(-1, method, 0, nil)}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.CancelAllOrders(c.reqID, c.param.market)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsOrderHistory(t *testing.T) {
	method := privateGetOrderHistory
	type input struct {
		reqID int
		body  *TradeParams
	}
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testCases := []struct {
		in                    input
		body                  *mockBody
		expectedParams        KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{input{0, &TradeParams{Market: "BTC"}}, nil, nil, true, false},
		{input{0, nil}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[0]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[0]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[1]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[1]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", EndTS: timeAgo[2]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "end_ts": timeAgo[2]}, false, true},
		{input{1212, &TradeParams{Page: 20}}, &mockBody{200, mockResponseBody(1212, method, 0, mockOrderHistory())}, KVParams{"page": 20}, false, false},
		{input{121212, &TradeParams{PageSize: 1}}, &mockBody{200, mockResponseBody(1212, method, 0, mockOrderHistory(OrderInfo{}))}, KVParams{"page_size": 1}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsGetOrderHistory(c.in.reqID, c.in.body)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsOpenOrders(t *testing.T) {
	method := privateGetOpenOrders
	testCases := []struct {
		reqID                 int
		param                 *OpenOrderParam
		expectedParams        KVParams
		body                  *mockBody
		shouldValidationError bool
		shouldError           bool
	}{
		{0, &OpenOrderParam{"-", 0, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC", 0, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC_USDT", -1, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC_USDT", 0, -1}, nil, nil, true, false},
		// valid values
		{0, nil, KVParams{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{}, KVParams{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{"BTC_USDT", 0, 0}, KVParams{"instrument_name": "BTC_USDT"}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{"BTC_USDT", 10, 0}, KVParams{"instrument_name": "BTC_USDT", "page_size": 10}, &mockBody{200, mockResponseBody(-1, method, 0, mockOpenOrders(1, OrderInfo{InstrumentName: "BTC_USDT", Status: "ACTIVE"}))}, false, false},
		{0, &OpenOrderParam{"BTC_USDT", 10, 1}, KVParams{"instrument_name": "BTC_USDT", "page_size": 10, "page": 1}, &mockBody{200, mockResponseBody(-1, method, 0, mockOpenOrders(1, OrderInfo{InstrumentName: "BTC_USDT", Status: "ACTIVE"}))}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsGetOpenOrders(c.reqID, c.param)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsGetOrderDetail(t *testing.T) {
	method := privateGetOrderDetail
	type input struct {
		reqID    int
		remoteID string
	}
	testCases := []struct {
		in                    input
		body                  *mockBody
		expectedBody          KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{input{0, "0"}, nil, nil, true, false},
		{input{0, "1212121212"}, &mockBody{400, mockResponseBody(-1, method, 10004, nil)}, KVParams{"order_id": "1212121212"}, false, true},
		{input{1213, "1212121212"}, &mockBody{200, mockResponseBody(1213, method, 0, mockOrderDetail(OrderInfo{}, Trade{}))}, KVParams{"order_id": "1212121212"}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.GetOrderDetails(c.in.reqID, c.in.remoteID)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, c.in.remoteID, pr.Params["order_id"])
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
func TestWsGetTrades(t *testing.T) {
	method := privateGetTrades
	type input struct {
		reqID int
		body  *TradeParams
	}
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testCases := []struct {
		in                    input
		body                  *mockBody
		expectedParams        KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{input{0, &TradeParams{Market: "BTC"}}, nil, nil, true, false},
		{input{0, nil}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[0]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[0]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[1]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[1]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", EndTS: timeAgo[2]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "end_ts": timeAgo[2]}, false, true},
		{input{1212, &TradeParams{Page: 20}}, &mockBody{200, mockResponseBody(1212, method, 0, mockTrades())}, KVParams{"page": 20}, false, false},
		{input{121212, &TradeParams{PageSize: 1}}, &mockBody{200, mockResponseBody(1212, method, 0, mockTrades())}, KVParams{"page_size": 1}, false, false},
	}
	for _, c := range testCases {
		cli, public, private := mockWsClient()
		if !c.shouldValidationError {
			var err error
			if c.shouldError {
				err = errors.New("some error")
			}
			private.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.WsGetTrades(c.in.reqID, c.in.body)
		var pr Request
		if c.shouldValidationError {
			assert.NotNil(t, err)
			private.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		if c.shouldError {
			assert.NotNil(t, err)
		}
		private.AssertExpectations(t)
		private.AssertNumberOfCalls(t, "WriteMessage", 1)
		public.AssertNumberOfCalls(t, "WriteMessage", 0)
		req := private.Calls[0].Arguments[1].([]byte)
		var maps map[string]interface{}
		_ = json.Unmarshal(req, &pr)
		_ = json.Unmarshal(req, &maps)
		b1, _ := json.Marshal(c.expectedParams)
		b2, _ := json.Marshal(pr.Params)
		assert.Equal(t, method, pr.Method)
		assert.Equal(t, string(b1), string(b2))
		_, ok := maps["sig"]
		assert.False(t, ok)
		_, ok = maps["api_key"]
		assert.False(t, ok)
	}
}
