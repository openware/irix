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
		for _, r := range result.Result.Instruments {
			res = append(res, r)
		}
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
