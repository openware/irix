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

func (c *Client) CancelOrder(reqID int, remoteID, market string) error {
	r, _ := c.cancelOrder(
		reqID,
		market,
		remoteID,
	)
	return c.sendPrivateRequest(r)
}

func (c *Client) CancelAllOrders(reqID int, market string) error {
	r, _ := c.cancelAllOrder(reqID, market)
	return c.sendPrivateRequest(r)
}

func (c *Client) GetOrderDetails(reqID int, remoteID string) error {
	r, _ := c.getOrderDetail(reqID, remoteID)
	return c.sendPrivateRequest(r)
}

func (c *Client) respondHeartBeat(isPrivate bool, id int) {
	r, _ := c.heartbeat(id)

	if isPrivate {
		c.sendPrivateRequest(r)
	} else {
		c.sendPublicRequest(r)
	}
}
