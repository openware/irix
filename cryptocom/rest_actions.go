package cryptocom

import (
	"bytes"
	"encoding/json"
)

const apiVersionSuffix = "/v2/"

func (c *Client) RestGetOrderDetails(reqID int, remoteID string) (Response, error) {
	r := c.restGetOrderDetailsRequest(reqID, remoteID)
	return c.send(r)
}

func (c *Client) RestGetBalance(reqID int) (Response, error) {
	r := c.restGetBalanceRequest(reqID)
	return c.send(r)
}

func (c *Client) RestGetTrades(reqID int, market string) (Response, error) {
	r := c.restGetTradesRequest(reqID, market)
	return c.send(r)
}

func (c *Client) RestOpenOrders(reqID int, market string, pageNumber int, pageSize int) (Response, error) {
	r := c.restOpenOrdersRequest(reqID, market, pageNumber, pageSize)
	return c.send(r)
}

func (c *Client) RestGetInstruments() (res []Instruments, err error) {
	req := c.getInstruments()
	var result InstrumentResponse
	_, err = c.rest.Send("GET", req, &result)
	if err == nil {
		res = result.Result.Instruments
	}
	return
}
func (c *Client) RestGetOrderBook(reqID int, instrumentName string, depth int) (res OrderbookResult, err error) {
	req, err := c.getOrderBook(reqID, instrumentName, depth)
	if err != nil {
		return
	}
	var result OrderbookResponse
	_, err = c.rest.Send("GET", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}
func (c *Client) RestGetCandlestick(instrumentName string, interval Interval, depth int) (res CandlestickResult, err error) {
	req, err := c.getCandlestick(instrumentName, interval, depth)
	if err != nil {
		return
	}
	var result CandlestickResponse
	_, err = c.rest.Send("GET", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}
func (c *Client) RestGetTicker(instrumentName string) (res TickerResult, err error) {
	req, err := c.getTicker(instrumentName)
	if err != nil {
		return
	}
	var result TickerResponse
	_, err = c.rest.Send("GET", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}
func (c *Client) RestGetPublicTrades(instrumentName string) (res PublicTradeResult, err error) {
	req, err := c.getPublicTrades(instrumentName)
	if err != nil {
		return
	}
	var result PublicTradeResponse
	_, err = c.rest.Send("GET", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}

func (c *Client) send(r *Request) (Response, error) {
	body, err := r.Encode()
	if err != nil {
		return Response{}, err
	}

	resp, err := c.httpClient.Post(c.restRootURL+apiVersionSuffix+r.Method, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var parsed Response
	err = json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed, err
}
