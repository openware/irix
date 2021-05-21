package cryptocom

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type formater func(string) string

// Input: ["ETH_BTC", "ETH_CRO"]
func format(markets []string, fn formater) []string {
	channels := []string{}
	for _, v := range markets {
		channels = append(channels, fn(v))
	}

	return channels
}

// SubscribePublicTrades is subscription trade channel
// Example: SubscribeTrades("ETH_BTC", "ETH_CRO")
func (c *Client) SubscribePublicTrades(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("trade.%s", s)
	})

	return c.subscribePublicChannels(channels, true)
}
func (c *Client) validateMarkets(markets ...string) error {
	if len(markets) == 0 {
		return errors.New("set at least one market to subscribe")
	}
	for _, mrk := range markets {
		if err := validInstrument(mrk); err != nil {
			return err
		}
	}
	return nil
}

// SubscribePublicOrderBook is subscription orderbook channel
// Example: SubscribeOrderBook(depth, "ETH_BTC", "ETH_CRO")
// depth: Number of bids and asks to return. Allowed values: 10 or 150
func (c *Client) SubscribePublicOrderBook(depth int, markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	if depth < 10 || depth > 150 {
		return errors.New("depth value is out of range. Allowed values are between 10 and/or 150")
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("book.%s.%d", s, depth)
	})

	return c.subscribePublicChannels(channels, true)
}

// SubscribePublicTickers is subscription ticker channel
func (c *Client) SubscribePublicTickers(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("ticker.%s", s)
	})

	return c.subscribePublicChannels(channels, true)
}
// SubscribeCandlestick is subscription to candlestick channel
func (c *Client) SubscribeCandlestick(interval Interval, markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	if interval < Minute1 || interval > Month {
		return errors.New("invalid interval value")
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("candlestick.%s.%s", interval.Encode(), s)
	})

	return c.subscribePublicChannels(channels, true)
}

// SubscribePrivateOrders is subscription private order user.order.markets channel
func (c *Client) SubscribePrivateOrders(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("user.order.%s", s)
	})

	return c.subscribePrivateChannels(channels, true)
}

// SubscribePrivateTrades is subscription private user.trade channel
func (c *Client) SubscribePrivateTrades(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("user.trade.%s", s)
	})

	return c.subscribePrivateChannels(channels, true)
}
// SubscribePrivateMarginOrders is subscription private order user.margin.order.markets channel
func (c *Client) SubscribePrivateMarginOrders(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("user.margin.order.%s", s)
	})

	return c.subscribePrivateChannels(channels, true)
}

// SubscribePrivateMarginTrades is subscription private user.margin.trade channel
func (c *Client) SubscribePrivateMarginTrades(markets ...string) error {
	if err := c.validateMarkets(markets...); err != nil {
		return err
	}
	channels := format(markets, func(s string) string {
		return fmt.Sprintf("user.margin.trade.%s", s)
	})

	return c.subscribePrivateChannels(channels, true)
}

// SubscribePrivateBalanceUpdates subscribe to user.balance channel
func (c *Client) SubscribePrivateBalanceUpdates() error {
	channels := []string{"user.balance"}
	return c.subscribePrivateChannels(channels, true)
}
// SubscribePrivateMarginBalanceUpdates subscribe to user.margin.balance channel
func (c *Client) SubscribePrivateMarginBalanceUpdates() error {
	channels := []string{"user.margin.balance"}
	return c.subscribePrivateChannels(channels, true)
}

// For MARKET BUY orders, amount is notional (https://exchange-docs.crypto.com/#private-create-order).
func (c *Client) CreateLimitOrder(
	reqID int,
	ask string,
	bid string,
	orderSide string,
	price decimal.Decimal,
	amount decimal.Decimal,
	uuid uuid.UUID,
) error {
	r := c.createOrderLimitRequest(
		reqID,
		ask,
		bid,
		orderSide,
		price,
		amount,
		uuid,
	)
	return c.sendPrivateRequest(r)
}

func (c *Client) CreateMarketOrder(
	reqID int,
	ask string,
	bid string,
	orderSide string,
	amount decimal.Decimal,
	uuid uuid.UUID,
) error {
	r := c.createOrderMarketRequest(
		reqID,
		ask,
		bid,
		orderSide,
		amount,
		uuid,
	)
	return c.sendPrivateRequest(r)
}

func (c *Client) WsCancelOrder(reqID int, remoteID, market string) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.cancelOrder(
			reqID,
			market,
			remoteID,
		)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsCancelAllOrders(reqID int, market string) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.cancelAllOrder(reqID, market)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsGetOrderHistory(reqID int, in *TradeParams) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.privateGetOrderHistory(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsGetOpenOrders(reqID int, in *OpenOrderParam) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getOpenOrders(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsGetOrderDetails(reqID int, remoteID string) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getOrderDetail(reqID, remoteID)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsGetTrades(reqID int, in *TradeParams) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getPrivateTrades(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

// WsGetInstruments Get markets/instruments via websocket
func (c *Client) WsGetInstruments() error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req = c.getInstruments()
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsSetCancelOnDisconnect(scope string) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.setCancelOnDisconnect(scope)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsGetCancelOnDisconnect() error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getCancelOnDisconnect()
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) WsCreateWithdrawal(reqID int, in WithdrawParams) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.createWithdrawal(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}
func (c *Client) WsGetWithdrawalHistory(reqID int, in *WithdrawHistoryParam) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getWithdrawalHistory(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}
func (c *Client) WsGetAccountSummary(market string) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.getAccountSummary(market)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}
func (c *Client) WsCreateOrder(reqID int, in CreateOrderParam) error {
	var (
		req *Request
	)
	return tryOrError(func() (err error) {
		req, err = c.createOrder(reqID, in)
		return
	}, func() error {
		return c.sendPrivateRequest(req)
	})
}

func (c *Client) respondHeartBeat(isPrivate bool, id int) {
	r, _ := c.heartbeat(id)

	if isPrivate {
		c.sendPrivateRequest(r)
	} else {
		c.sendPublicRequest(r)
	}
}
