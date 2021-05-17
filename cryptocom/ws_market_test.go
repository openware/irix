package cryptocom

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestPublicOrderBook(t *testing.T) {
	testTable := []struct{
		depth int
		instruments []string
		validationError bool
		shouldError bool
	}{
		{0, nil, true, false},
		{10, []string{}, true, false},
		{10, []string{"BTC"}, true, false},
		{10, []string{"BTC_USDT", "ETCH"}, true, false},
		{1000, []string{"BTC_USDT"}, true, false},
		{10, []string{"BTC_USDT"}, false, true},
		{150, []string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			public.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePublicOrderBook(c.depth, c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			public.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		public.AssertExpectations(t)
		public.AssertNumberOfCalls(t, "WriteMessage", 1)
		private.AssertNotCalled(t, "WriteMessage")
		req := public.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("book.%s.%d", s, c.depth)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPublicTicker(t *testing.T) {
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
			public.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePublicTickers(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			public.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		public.AssertExpectations(t)
		public.AssertNumberOfCalls(t, "WriteMessage", 1)
		private.AssertNotCalled(t, "WriteMessage")
		req := public.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("ticker.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
func TestPublicTrade(t *testing.T) {
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
			public.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribePublicTrades(c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			public.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		public.AssertExpectations(t)
		public.AssertNumberOfCalls(t, "WriteMessage", 1)
		private.AssertNotCalled(t, "WriteMessage")
		req := public.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("trade.%s", s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}

func TestPublicCandlestick(t *testing.T) {
	testTable := []struct{
		interval Interval
		instruments []string
		validationError bool
		shouldError bool
	}{
		{ 0, nil, true, false},
		{ 100, nil, true, false},
		{Minute1, []string{}, true, false},
		{Minute5, []string{"BTC"}, true, false},
		{Day, []string{"BTC_USDT", "ETCH"}, true, false},
		{Minute1, []string{"BTC_USDT"}, false, false},
		{Minute5, []string{"BTC_USDT"}, false, false},
		{Minute15, []string{"BTC_USDT"}, false, false},
		{Minute30, []string{"BTC_USDT"}, false, false},
		{Hour1, []string{"BTC_USDT"}, false, false},
		{Hour4, []string{"BTC_USDT"}, false, false},
		{Hour6, []string{"BTC_USDT"}, false, false},
		{Hour12, []string{"BTC_USDT"}, false, false},
		{Day, []string{"BTC_USDT"}, false, false},
		{Week, []string{"BTC_USDT"}, false, false},
		{Week2, []string{"BTC_USDT"}, false, false},
		{Month, []string{"BTC_USDT"}, false, false},
	}
	for _, c := range testTable {
		cli, public, private := mockWsClient()
		if !c.validationError {
			var err error
			if c.shouldError {
				err = errors.New("disconnected")
			}
			public.
				On("WriteMessage", mock.Anything, mock.Anything).
				Return(err)
		}
		err := cli.SubscribeCandlestick(c.interval, c.instruments...)
		if c.validationError {
			assert.NotNil(t, err)
			public.AssertNumberOfCalls(t, "WriteMessage", 0)
			continue
		}
		public.AssertExpectations(t)
		public.AssertNumberOfCalls(t, "WriteMessage", 1)
		private.AssertNotCalled(t, "WriteMessage")
		req := public.Calls[0].Arguments[1].([]byte)
		var pr Request
		_ = json.Unmarshal(req, &pr)
		formatted := format(c.instruments, func(s string) string {
			return fmt.Sprintf("candlestick.%s.%s", c.interval.Encode(), s)
		})
		assert.Equal(t, subscribe, pr.Method)
		channels := pr.Params["channels"].([]interface{})
		for k, f := range formatted {
			assert.Equal(t, f, channels[k])
		}
	}
}
