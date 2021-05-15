package cryptocom

func (c *Client) RestGetOrderDetails(reqID int, remoteID string) (res *OrderDetail, err error) {
	var (
		req *Request
		response OrderDetailResponse
	)
	if err = tryOrError(func() error {
		req, err = c.getOrderDetail(reqID, remoteID)
		return err
	}, func() error {
		c.generateSignature(req)
		_, err := c.rest.Send("POST", req, &response)
		return err
	}); err != nil {
		return
	}
	res = &response.Result
	return
}

func (c *Client) RestGetTrades(reqID int, param *TradeParams) (res *TradeResult, err error) {
	var (
		req *Request
		response TradeResponse
	)
	if err = tryOrError(func() error {
		req, err = c.privateGetTrades(reqID, param)
		return err
	}, func() error {
		c.generateSignature(req)
		_, err := c.rest.Send("POST", req, &response)
		return err
	}); err != nil {
		return
	}
	res = &response.Result
	return
}
func (c *Client) RestOpenOrders(reqID int, param *OpenOrderParam) (res *OpenOrdersResult, err error) {
	var (
		req *Request
		response OpenOrdersResponse
	)
	if err = tryOrError(func() (err error) {
		req, err = c.getOpenOrders(reqID, param)
		return
	}, func() error {
		c.generateSignature(req)
		_, err := c.rest.Send("POST", req, &response)
		return err
	}); err != nil {
		return
	}
	res = &response.Result
	return
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
func (c *Client) RestGetDepositAddress(currency string) (res DepositAddressResult, err error) {
	req, err := c.getDepositAddress(currency)
	if err != nil {
		return
	}
	c.generateSignature(req)
	var result DepositAddressResponse
	_, err = c.rest.Send("POST", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}
func (c *Client) RestGetAccountSummary(currency string) (res AccountResult, err error) {
	req, err := c.getAccountSummary(currency)
	if err != nil {
		return
	}
	c.generateSignature(req)
	var result AccountResponse
	_, err = c.rest.Send("POST", req, &result)
	if err == nil {
		res = result.Result
	}
	return
}

func (c *Client) RestCreateWithdrawal(reqID int, in WithdrawParams) (res *Withdraw, err error) {
	var (
		req *Request
		result WithdrawResponse
	)
	if err = tryOrError(func() (err error) {
		req, err = c.createWithdrawal(reqID, in)
		return
	}, func() (err error) {
		c.generateSignature(req)
		_, err = c.rest.Send("POST", req, &result)
		return
	}); err != nil {
		return
	}
	res = &result.Result
	return
}

func (c *Client) RestGetWithdrawalHistory(reqID int, in *WithdrawHistoryParam) (res *WithdrawHistoryResult, err error) {
	var (
		req *Request
		result WithdrawHistoryResponse
	)
	if err = tryOrError(func() (err error) {
		req, err = c.getWithdrawalHistory(reqID, in)
		return
	}, func() (err error) {
		c.generateSignature(req)
		_, err = c.rest.Send("POST", req, &result)
		return
	}); err != nil {
		return
	}
	res = &result.Result
	return
}

func (c *Client) RestGetDepositHistory(reqID int, in *DepositHistoryParam) (res *DepositHistoryResult, err error) {
	var (
		req *Request
		result DepositHistoryResponse
	)
	if err = tryOrError(func() (err error) {
		req, err = c.getDepositHistory(reqID, in)
		return
	}, func() (err error) {
		c.generateSignature(req)
		_, err = c.rest.Send("POST", req, &result)
		return
	}); err != nil {
		return
	}
	res = &result.Result
	return
}

// RestCreateOrder create order via Spot HTTP API
// it returns Order as a result and error if occurred
func (c *Client) RestCreateOrder(reqID int, in CreateOrderParam) (res *Order, err error) {
	var (
		req *Request
		result OrderResult
	)
	if err = tryOrError(func() (err error) {
		req, err = c.createOrder(reqID, in)
		return
	}, func() (err error) {
		c.generateSignature(req)
		_, err = c.rest.Send("POST", req, &result)
		return
	}); err != nil {
		return
	}
	res = &result.Result
	return
}

func (c *Client) RestCancelOrder(reqID int, market, orderID string) (res bool, err error) {
	var (
		req *Request
	)
	err = tryOrError(func() (err error) {
		req, err = c.cancelOrder(reqID, market, orderID)
		return
	}, func() (err error) {
		c.generateSignature(req)
		result, err := c.rest.Send("POST", req, nil)
		if err == nil {
			res = result.Code == 0
		}
		return
	})
	return
}

