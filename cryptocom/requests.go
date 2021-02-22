package cryptocom

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (c *Client) AuthRequest() *Request {
	r := &Request{
		Id:     1,
		Type:   AuthRequest,
		Method: "public/auth",
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
}

func (c *Client) subscribeRequest(channels []string) *Request {
	return &Request{
		Id:     1,
		Type:   SubscribeRequest,
		Method: "subscribe",
		Params: map[string]interface{}{"channels": channels},
		Nonce:  generateNonce(),
	}
}

func (c *Client) hearBeatRequest(reqId int) *Request {
	return &Request{
		Id:     reqId,
		Type:   HeartBeat,
		Method: "public/respond-heartbeat",
	}
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
		Type:   OrderRequest,
		Method: "private/create-order",
		Params: map[string]interface{}{
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
		Type:   OrderRequest,
		Method: "private/create-order",
		Params: map[string]interface{}{
			"instrument_name": strings.ToUpper(ask) + "_" + strings.ToUpper(bid),
			"side":            strings.ToUpper(orderSide),
			"type":            "MARKET",
			volumeKey:         volume,
			"client_oid":      uuid,
		},
		Nonce: generateNonce(),
	}
}

func (c *Client) cancelOrderRequest(reqID int, remoteID, market string) *Request {
	return &Request{
		Id:     reqID,
		Type:   OrderRequest,
		Method: "private/cancel-order",
		Params: map[string]interface{}{
			"instrument_name": market,
			"order_id":        remoteID,
		},
		Nonce: generateNonce(),
	}
}

// Market: "ETH_BTC"
func (c *Client) cancelAllOrdersRequest(reqID int, market string) *Request {
	return &Request{
		Id:     reqID,
		Type:   OrderRequest,
		Method: "private/cancel-all-orders",
		Params: map[string]interface{}{
			"instrument_name": market,
		},
		Nonce: generateNonce(),
	}
}

func (c *Client) getOrderDetailsRequest(reqID int, remoteID string) *Request {
	return &Request{
		Id:     reqID,
		Type:   OrderRequest,
		Method: "private/get-order-detail",
		Params: map[string]interface{}{
			"order_id": remoteID,
		},
		Nonce: generateNonce(),
	}
}

func (c *Client) restGetOrderDetailsRequest(reqID int, remoteID string) *Request {
	r := &Request{
		Id:     reqID,
		Type:   RestOrderRequest,
		Method: "private/get-order-detail",
		Params: map[string]interface{}{
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
		Type:   RestBalanceRequest,
		Method: "private/get-account-summary",
		Params: map[string]interface{}{},
		ApiKey: c.key,
		Nonce:  generateNonce(),
	}

	c.generateSignature(r)
	return r
}

func (c *Client) restGetTradesRequest(reqID int, market string) *Request {
	r := &Request{
		Id:     reqID,
		Type:   RestTradesRequest,
		Method: "private/get-trades",
		Params: map[string]interface{}{
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
		Type:   RestOpenOrdersRequest,
		Method: "private/get-open-orders",
		Params: map[string]interface{}{
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
