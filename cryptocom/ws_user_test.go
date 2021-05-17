package cryptocom

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestPrivateUserOrder(t *testing.T) {
	testTable := []struct{
		instruments []string
		validationError bool
		shouldError bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
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
	testTable := []struct{
		instruments []string
		validationError bool
		shouldError bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
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
	testTable := []struct{
		instruments []string
		validationError bool
		shouldError bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
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
	testTable := []struct{
		instruments []string
		validationError bool
		shouldError bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"BTC"}, true, false},
		{[]string{"BTC_USDT", "ETCH"}, true, false},
		{[]string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
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
	testTable := []struct {
		shouldError bool
	}{
		{true},
		{false},
	}
	for _, c := range testTable {
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
	testTable := []struct {
		shouldError bool
	}{
		{true},
		{false},
	}
	for _, c := range testTable {
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
