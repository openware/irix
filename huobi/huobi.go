package huobi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	exchange "github.com/openware/irix"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/common/crypto"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/request"
)

const (
	huobiAPIURL      = "https://api.huobi.pro"
	huobiURL         = "https://api.hbdm.com/"
	huobiFuturesURL  = "https://api.hbdm.com"
	huobiAPIVersion  = "1"
	huobiAPIVersion2 = "2"

	// Spot endpoints
	huobiMarketHistoryKline    = "market/history/kline"
	huobiMarketDetail          = "market/detail"
	huobiMarketDetailMerged    = "market/detail/merged"
	huobi24HrMarketSummary     = "/market/detail?"
	huobiMarketDepth           = "market/depth"
	huobiMarketTrade           = "market/trade"
	huobiMarketTickers         = "market/tickers"
	huobiMarketTradeHistory    = "market/history/trade"
	huobiSymbols               = "common/symbols"
	huobiCurrencies            = "common/currencys"
	huobiTimestamp             = "common/timestamp"
	huobiAccounts              = "account/accounts"
	huobiAccountBalance        = "account/accounts/%s/balance"
	huobiAccountDepositAddress = "account/deposit/address"
	huobiAccountWithdrawQuota  = "account/withdraw/quota"
	huobiAggregatedBalance     = "subuser/aggregate-balance"
	huobiOrderPlace            = "order/orders/place"
	huobiOrderCancel           = "order/orders/%s/submitcancel"
	huobiOrderCancelBatch      = "order/orders/batchcancel"
	huobiBatchCancelOpenOrders = "order/orders/batchCancelOpenOrders"
	huobiGetOrder              = "order/orders/getClientOrder"
	huobiGetOrderMatch         = "order/orders/%s/matchresults"
	huobiGetOrders             = "order/orders"
	huobiGetOpenOrders         = "order/openOrders"
	huobiGetOrdersMatch        = "orders/matchresults"
	huobiMarginTransferIn      = "dw/transfer-in/margin"
	huobiMarginTransferOut     = "dw/transfer-out/margin"
	huobiMarginOrders          = "margin/orders"
	huobiMarginRepay           = "margin/orders/%s/repay"
	huobiMarginLoanOrders      = "margin/loan-orders"
	huobiMarginAccountBalance  = "margin/accounts/balance"
	huobiWithdrawCreate        = "dw/withdraw/api/create"
	huobiWithdrawCancel        = "dw/withdraw-virtual/%s/cancel"
	huobiStatusError           = "error"
	huobiMarginRates           = "margin/loan-info"
)

// HUOBI is the overarching type across this package
type HUOBI struct {
	exchange.Base
	AccountID string
}

// GetMarginRates gets margin rates
func (h *HUOBI) GetMarginRates(symbol currency.Pair) (MarginRatesData, error) {
	var resp MarginRatesData
	vals := url.Values{}
	if !symbol.IsEmpty() {
		symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
		if err != nil {
			return resp, err
		}
		vals.Set("symbol", symbolValue)
	}
	return resp, h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiMarginRates, vals, nil, &resp, false)
}

// GetSpotKline returns kline data
// KlinesRequestParams contains symbol currency.Pair, period and size
func (h *HUOBI) GetSpotKline(arg KlinesRequestParams) ([]KlineItem, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(arg.Symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)
	vals.Set("period", arg.Period)

	if arg.Size != 0 {
		vals.Set("size", strconv.Itoa(arg.Size))
	}

	type response struct {
		Response
		Data []KlineItem `json:"data"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketHistoryKline, vals), &result)
	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.Data, err
}

// Get24HrMarketSummary returns 24hr market summary for a given market symbol
func (h *HUOBI) Get24HrMarketSummary(symbol currency.Pair) (MarketSummary24Hr, error) {
	var result MarketSummary24Hr
	params := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return result, err
	}
	params.Set("symbol", symbolValue)
	return result, h.SendHTTPRequest(exchange.RestSpot, huobi24HrMarketSummary+params.Encode(), &result)
}

// GetTickers returns the ticker for the specified symbol
func (h *HUOBI) GetTickers() (Tickers, error) {
	var result Tickers
	return result, h.SendHTTPRequest(exchange.RestSpot, "/"+huobiMarketTickers, &result)
}

// GetMarketDetailMerged returns the ticker for the specified symbol
func (h *HUOBI) GetMarketDetailMerged(symbol currency.Pair) (DetailMerged, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return DetailMerged{}, err
	}
	vals.Set("symbol", symbolValue)

	type response struct {
		Response
		Tick DetailMerged `json:"tick"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketDetailMerged, vals), &result)
	if result.ErrorMessage != "" {
		return result.Tick, errors.New(result.ErrorMessage)
	}
	return result.Tick, err
}

// GetDepth returns the depth for the specified symbol
func (h *HUOBI) GetDepth(obd OrderBookDataRequestParams) (Orderbook, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(obd.Symbol, asset.Spot)
	if err != nil {
		return Orderbook{}, err
	}
	vals.Set("symbol", symbolValue)

	if obd.Type != OrderBookDataRequestParamsTypeNone {
		vals.Set("type", string(obd.Type))
	}

	type response struct {
		Response
		Depth Orderbook `json:"tick"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketDepth, vals), &result)
	if result.ErrorMessage != "" {
		return result.Depth, errors.New(result.ErrorMessage)
	}
	return result.Depth, err
}

// GetTrades returns the trades for the specified symbol
func (h *HUOBI) GetTrades(symbol currency.Pair) ([]Trade, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)

	type response struct {
		Response
		Tick struct {
			Data []Trade `json:"data"`
		} `json:"tick"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketTrade, vals), &result)
	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.Tick.Data, err
}

// GetLatestSpotPrice returns latest spot price of symbol
//
// symbol: string of currency pair
func (h *HUOBI) GetLatestSpotPrice(symbol currency.Pair) (float64, error) {
	list, err := h.GetTradeHistory(symbol, 1)

	if err != nil {
		return 0, err
	}
	if len(list) == 0 {
		return 0, errors.New("the length of the list is 0")
	}

	return list[0].Trades[0].Price, nil
}

// GetTradeHistory returns the trades for the specified symbol
func (h *HUOBI) GetTradeHistory(symbol currency.Pair, size int64) ([]TradeHistory, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)

	if size > 0 {
		vals.Set("size", strconv.FormatInt(size, 10))
	}

	type response struct {
		Response
		TradeHistory []TradeHistory `json:"data"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketTradeHistory, vals), &result)
	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.TradeHistory, err
}

// GetMarketDetail returns the ticker for the specified symbol
func (h *HUOBI) GetMarketDetail(symbol currency.Pair) (Detail, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return Detail{}, err
	}
	vals.Set("symbol", symbolValue)

	type response struct {
		Response
		Tick Detail `json:"tick"`
	}

	var result response

	err = h.SendHTTPRequest(exchange.RestSpot, common.EncodeURLValues("/"+huobiMarketDetail, vals), &result)
	if result.ErrorMessage != "" {
		return result.Tick, errors.New(result.ErrorMessage)
	}
	return result.Tick, err
}

// GetSymbols returns an array of symbols supported by Huobi
func (h *HUOBI) GetSymbols() ([]Symbol, error) {
	type response struct {
		Response
		Symbols []Symbol `json:"data"`
	}

	var result response

	err := h.SendHTTPRequest(exchange.RestSpot, "/v"+huobiAPIVersion+"/"+huobiSymbols, &result)
	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.Symbols, err
}

// GetCurrencies returns a list of currencies supported by Huobi
func (h *HUOBI) GetCurrencies() ([]string, error) {
	type response struct {
		Response
		Currencies []string `json:"data"`
	}

	var result response

	err := h.SendHTTPRequest(exchange.RestSpot, "/v"+huobiAPIVersion+"/"+huobiCurrencies, &result)
	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.Currencies, err
}

// GetTimestamp returns the Huobi server time
func (h *HUOBI) GetTimestamp() (int64, error) {
	type response struct {
		Response
		Timestamp int64 `json:"data"`
	}

	var result response

	err := h.SendHTTPRequest(exchange.RestSpot, "/v"+huobiAPIVersion+"/"+huobiTimestamp, &result)
	if result.ErrorMessage != "" {
		return 0, errors.New(result.ErrorMessage)
	}
	return result.Timestamp, err
}

// GetAccounts returns the Huobi user accounts
func (h *HUOBI) GetAccounts() ([]Account, error) {
	result := struct {
		Accounts []Account `json:"data"`
	}{}
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiAccounts, url.Values{}, nil, &result, false)
	return result.Accounts, err
}

// GetAccountBalance returns the users Huobi account balance
func (h *HUOBI) GetAccountBalance(accountID string) ([]AccountBalanceDetail, error) {
	result := struct {
		AccountBalanceData AccountBalance `json:"data"`
	}{}
	endpoint := fmt.Sprintf(huobiAccountBalance, accountID)
	v := url.Values{}
	v.Set("account-id", accountID)
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, endpoint, v, nil, &result, false)
	return result.AccountBalanceData.AccountBalanceDetails, err
}

// GetAggregatedBalance returns the balances of all the sub-account aggregated.
func (h *HUOBI) GetAggregatedBalance() ([]AggregatedBalance, error) {
	result := struct {
		AggregatedBalances []AggregatedBalance `json:"data"`
	}{}
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot,
		http.MethodGet,
		huobiAggregatedBalance,
		nil,
		nil,
		&result,
		false,
	)
	return result.AggregatedBalances, err
}

// SpotNewOrder submits an order to Huobi
func (h *HUOBI) SpotNewOrder(arg *SpotNewOrderRequestParams) (int64, error) {
	symbolValue, err := h.FormatSymbol(arg.Symbol, asset.Spot)
	if err != nil {
		return 0, err
	}

	data := struct {
		AccountID int    `json:"account-id,string"`
		Amount    string `json:"amount"`
		Price     string `json:"price"`
		Source    string `json:"source"`
		Symbol    string `json:"symbol"`
		Type      string `json:"type"`
	}{
		AccountID: arg.AccountID,
		Amount:    strconv.FormatFloat(arg.Amount, 'f', -1, 64),
		Symbol:    symbolValue,
		Type:      string(arg.Type),
	}

	// Only set price if order type is not equal to buy-market or sell-market
	if arg.Type != SpotNewOrderRequestTypeBuyMarket && arg.Type != SpotNewOrderRequestTypeSellMarket {
		data.Price = strconv.FormatFloat(arg.Price, 'f', -1, 64)
	}

	if arg.Source != "" {
		data.Source = arg.Source
	}

	result := struct {
		OrderID int64 `json:"data,string"`
	}{}
	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot,
		http.MethodPost,
		huobiOrderPlace,
		nil,
		data,
		&result,
		false,
	)
	return result.OrderID, err
}

// CancelExistingOrder cancels an order on Huobi
func (h *HUOBI) CancelExistingOrder(orderID int64) (int64, error) {
	resp := struct {
		OrderID int64 `json:"data,string"`
	}{}
	endpoint := fmt.Sprintf(huobiOrderCancel, strconv.FormatInt(orderID, 10))
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, endpoint, url.Values{}, nil, &resp, false)
	return resp.OrderID, err
}

// CancelOrderBatch cancels a batch of orders -- to-do
func (h *HUOBI) CancelOrderBatch(_ []int64) ([]CancelOrderBatch, error) {
	type response struct {
		Response
		Data []CancelOrderBatch `json:"data"`
	}

	var result response
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, huobiOrderCancelBatch, url.Values{}, nil, &result, false)

	if result.ErrorMessage != "" {
		return nil, errors.New(result.ErrorMessage)
	}
	return result.Data, err
}

// CancelOpenOrdersBatch cancels a batch of orders -- to-do
func (h *HUOBI) CancelOpenOrdersBatch(accountID string, symbol currency.Pair) (CancelOpenOrdersBatch, error) {
	params := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return CancelOpenOrdersBatch{}, err
	}
	params.Set("account-id", accountID)
	var result CancelOpenOrdersBatch

	data := struct {
		AccountID string `json:"account-id"`
		Symbol    string `json:"symbol"`
	}{
		AccountID: accountID,
		Symbol:    symbolValue,
	}

	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, huobiBatchCancelOpenOrders, url.Values{}, data, &result, false)
	if result.Data.FailedCount > 0 {
		return result, fmt.Errorf("there were %v failed order cancellations", result.Data.FailedCount)
	}

	return result, err
}

// GetOrder returns order information for the specified order
func (h *HUOBI) GetOrder(orderID int64) (OrderInfo, error) {
	resp := struct {
		Order OrderInfo `json:"data"`
	}{}
	urlVal := url.Values{}
	urlVal.Set("clientOrderId", strconv.FormatInt(orderID, 10))
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet,
		huobiGetOrder,
		urlVal,
		nil,
		&resp,
		false)
	return resp.Order, err
}

// GetOrderMatchResults returns matched order info for the specified order
func (h *HUOBI) GetOrderMatchResults(orderID int64) ([]OrderMatchInfo, error) {
	resp := struct {
		Orders []OrderMatchInfo `json:"data"`
	}{}
	endpoint := fmt.Sprintf(huobiGetOrderMatch, strconv.FormatInt(orderID, 10))
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, endpoint, url.Values{}, nil, &resp, false)
	return resp.Orders, err
}

// GetOrders returns a list of orders
func (h *HUOBI) GetOrders(symbol currency.Pair, types, start, end, states, from, direct, size string) ([]OrderInfo, error) {
	resp := struct {
		Orders []OrderInfo `json:"data"`
	}{}

	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)
	vals.Set("states", states)

	if types != "" {
		vals.Set("types", types)
	}

	if start != "" {
		vals.Set("start-date", start)
	}

	if end != "" {
		vals.Set("end-date", end)
	}

	if from != "" {
		vals.Set("from", from)
	}

	if direct != "" {
		vals.Set("direct", direct)
	}

	if size != "" {
		vals.Set("size", size)
	}

	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiGetOrders, vals, nil, &resp, false)
	return resp.Orders, err
}

// GetOpenOrders returns a list of orders
func (h *HUOBI) GetOpenOrders(symbol currency.Pair, accountID, side string, size int64) ([]OrderInfo, error) {
	resp := struct {
		Orders []OrderInfo `json:"data"`
	}{}

	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)
	vals.Set("accountID", accountID)
	if len(side) > 0 {
		vals.Set("side", side)
	}
	vals.Set("size", strconv.FormatInt(size, 10))

	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiGetOpenOrders, vals, nil, &resp, false)
	return resp.Orders, err
}

// GetOrdersMatch returns a list of matched orders
func (h *HUOBI) GetOrdersMatch(symbol currency.Pair, types, start, end, from, direct, size string) ([]OrderMatchInfo, error) {
	resp := struct {
		Orders []OrderMatchInfo `json:"data"`
	}{}

	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)

	if types != "" {
		vals.Set("types", types)
	}

	if start != "" {
		vals.Set("start-date", start)
	}

	if end != "" {
		vals.Set("end-date", end)
	}

	if from != "" {
		vals.Set("from", from)
	}

	if direct != "" {
		vals.Set("direct", direct)
	}

	if size != "" {
		vals.Set("size", size)
	}

	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiGetOrdersMatch, vals, nil, &resp, false)
	return resp.Orders, err
}

// MarginTransfer transfers assets into or out of the margin account
func (h *HUOBI) MarginTransfer(symbol currency.Pair, currency string, amount float64, in bool) (int64, error) {
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return 0, err
	}
	data := struct {
		Symbol   string `json:"symbol"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	}{
		Symbol:   symbolValue,
		Currency: currency,
		Amount:   strconv.FormatFloat(amount, 'f', -1, 64),
	}

	path := huobiMarginTransferIn
	if !in {
		path = huobiMarginTransferOut
	}

	resp := struct {
		TransferID int64 `json:"data"`
	}{}
	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, path, nil, data, &resp, false)
	return resp.TransferID, err
}

// MarginOrder submits a margin order application
func (h *HUOBI) MarginOrder(symbol currency.Pair, currency string, amount float64) (int64, error) {
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return 0, err
	}
	data := struct {
		Symbol   string `json:"symbol"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	}{
		Symbol:   symbolValue,
		Currency: currency,
		Amount:   strconv.FormatFloat(amount, 'f', -1, 64),
	}

	resp := struct {
		MarginOrderID int64 `json:"data"`
	}{}
	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, huobiMarginOrders, nil, data, &resp, false)
	return resp.MarginOrderID, err
}

// MarginRepayment repays a margin amount for a margin ID
func (h *HUOBI) MarginRepayment(orderID int64, amount float64) (int64, error) {
	data := struct {
		Amount string `json:"amount"`
	}{
		Amount: strconv.FormatFloat(amount, 'f', -1, 64),
	}

	resp := struct {
		MarginOrderID int64 `json:"data"`
	}{}

	endpoint := fmt.Sprintf(huobiMarginRepay, strconv.FormatInt(orderID, 10))
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, endpoint, nil, data, &resp, false)
	return resp.MarginOrderID, err
}

// GetMarginLoanOrders returns the margin loan orders
func (h *HUOBI) GetMarginLoanOrders(symbol currency.Pair, currency, start, end, states, from, direct, size string) ([]MarginOrder, error) {
	vals := url.Values{}
	symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
	if err != nil {
		return nil, err
	}
	vals.Set("symbol", symbolValue)
	vals.Set("currency", currency)

	if start != "" {
		vals.Set("start-date", start)
	}

	if end != "" {
		vals.Set("end-date", end)
	}

	if states != "" {
		vals.Set("states", states)
	}

	if from != "" {
		vals.Set("from", from)
	}

	if direct != "" {
		vals.Set("direct", direct)
	}

	if size != "" {
		vals.Set("size", size)
	}

	resp := struct {
		MarginLoanOrders []MarginOrder `json:"data"`
	}{}
	err = h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiMarginLoanOrders, vals, nil, &resp, false)
	return resp.MarginLoanOrders, err
}

// GetMarginAccountBalance returns the margin account balances
func (h *HUOBI) GetMarginAccountBalance(symbol currency.Pair) ([]MarginAccountBalance, error) {
	resp := struct {
		Balances []MarginAccountBalance `json:"data"`
	}{}
	vals := url.Values{}
	if !symbol.IsEmpty() {
		symbolValue, err := h.FormatSymbol(symbol, asset.Spot)
		if err != nil {
			return resp.Balances, err
		}
		vals.Set("symbol", symbolValue)
	}
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiMarginAccountBalance, vals, nil, &resp, false)
	return resp.Balances, err
}

// Withdraw withdraws the desired amount and currency
func (h *HUOBI) Withdraw(c currency.Code, address, addrTag string, amount, fee float64) (int64, error) {
	resp := struct {
		WithdrawID int64 `json:"data"`
	}{}

	data := struct {
		Address  string `json:"address"`
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
		Fee      string `json:"fee,omitempty"`
		AddrTag  string `json:"addr-tag,omitempty"`
	}{
		Address:  address,
		Currency: c.Lower().String(),
		Amount:   strconv.FormatFloat(amount, 'f', -1, 64),
	}

	if fee > 0 {
		data.Fee = strconv.FormatFloat(fee, 'f', -1, 64)
	}

	if c == currency.XRP {
		data.AddrTag = addrTag
	}

	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, huobiWithdrawCreate, nil, data, &resp.WithdrawID, false)
	return resp.WithdrawID, err
}

// CancelWithdraw cancels a withdraw request
func (h *HUOBI) CancelWithdraw(withdrawID int64) (int64, error) {
	resp := struct {
		WithdrawID int64 `json:"data"`
	}{}
	vals := url.Values{}
	vals.Set("withdraw-id", strconv.FormatInt(withdrawID, 10))

	endpoint := fmt.Sprintf(huobiWithdrawCancel, strconv.FormatInt(withdrawID, 10))
	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, endpoint, vals, nil, &resp, false)
	return resp.WithdrawID, err
}

// QueryDepositAddress returns the deposit address for a specified currency
func (h *HUOBI) QueryDepositAddress(cryptocurrency string) (DepositAddress, error) {
	resp := struct {
		DepositAddress []DepositAddress `json:"data"`
	}{}

	vals := url.Values{}
	vals.Set("currency", cryptocurrency)

	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiAccountDepositAddress, vals, nil, &resp, true)
	if err != nil {
		return DepositAddress{}, err
	}
	if len(resp.DepositAddress) == 0 {
		return DepositAddress{}, errors.New("deposit address data isn't populated")
	}
	return resp.DepositAddress[0], nil
}

// QueryWithdrawQuotas returns the users cryptocurrency withdraw quotas
func (h *HUOBI) QueryWithdrawQuotas(cryptocurrency string) (WithdrawQuota, error) {
	resp := struct {
		WithdrawQuota WithdrawQuota `json:"data"`
	}{}

	vals := url.Values{}
	vals.Set("currency", cryptocurrency)

	err := h.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, huobiAccountWithdrawQuota, vals, nil, &resp, true)
	if err != nil {
		return WithdrawQuota{}, err
	}
	return resp.WithdrawQuota, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (h *HUOBI) SendHTTPRequest(ep exchange.URL, path string, result interface{}) error {
	endpoint, err := h.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	var tempResp json.RawMessage
	var errCap errorCapture
	err = h.SendPayload(context.Background(), &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        &tempResp,
		Verbose:       h.Verbose,
		HTTPDebugging: h.HTTPDebugging,
		HTTPRecording: h.HTTPRecording,
	})
	if err != nil {
		return err
	}
	if err := json.Unmarshal(tempResp, &errCap); err == nil {
		if errCap.Code != 200 && errCap.ErrMsg != "" {
			return errors.New(errCap.ErrMsg)
		}
	}
	return json.Unmarshal(tempResp, result)
}

// SendAuthenticatedHTTPRequest sends authenticated requests to the HUOBI API
func (h *HUOBI) SendAuthenticatedHTTPRequest(ep exchange.URL, method, endpoint string, values url.Values, data, result interface{}, isVersion2API bool) error {
	if !h.AllowAuthenticatedRequest() {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, h.Name)
	}
	ePoint, err := h.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	if values == nil {
		values = url.Values{}
	}

	now := time.Now()
	values.Set("AccessKeyId", h.API.Credentials.Key)
	values.Set("SignatureMethod", "HmacSHA256")
	values.Set("SignatureVersion", "2")
	values.Set("Timestamp", now.UTC().Format("2006-01-02T15:04:05"))

	if isVersion2API {
		endpoint = "/v" + huobiAPIVersion2 + "/" + endpoint
	} else {
		endpoint = "/v" + huobiAPIVersion + "/" + endpoint
	}

	payload := fmt.Sprintf("%s\napi.huobi.pro\n%s\n%s",
		method, endpoint, values.Encode())

	headers := make(map[string]string)

	if method == http.MethodGet {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	} else {
		headers["Content-Type"] = "application/json"
	}

	hmac := crypto.GetHMAC(crypto.HashSHA256, []byte(payload), []byte(h.API.Credentials.Secret))
	values.Set("Signature", crypto.Base64Encode(hmac))
	urlPath := ePoint + common.EncodeURLValues(endpoint, values)

	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	// Time difference between your timestamp and standard should be less than 1 minute.
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(time.Minute))
	defer cancel()
	interim := json.RawMessage{}
	err = h.SendPayload(ctx, &request.Item{
		Method:        method,
		Path:          urlPath,
		Headers:       headers,
		Body:          bytes.NewReader(body),
		Result:        &interim,
		AuthRequest:   true,
		Verbose:       h.Verbose,
		HTTPDebugging: h.HTTPDebugging,
		HTTPRecording: h.HTTPRecording,
	})
	if err != nil {
		return err
	}

	if isVersion2API {
		var errCap ResponseV2
		if err = json.Unmarshal(interim, &errCap); err == nil {
			if errCap.Code != 200 && errCap.Message != "" {
				return errors.New(errCap.Message)
			}
		}
	} else {
		var errCap Response
		if err = json.Unmarshal(interim, &errCap); err == nil {
			if errCap.Status == huobiStatusError && errCap.ErrorMessage != "" {
				return errors.New(errCap.ErrorMessage)
			}
		}
	}
	return json.Unmarshal(interim, result)
}

// GetFee returns an estimate of fee based on type of transaction
func (h *HUOBI) GetFee(feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	if feeBuilder.FeeType == exchange.OfflineTradeFee || feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		fee = calculateTradingFee(feeBuilder.Pair, feeBuilder.PurchasePrice, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func calculateTradingFee(c currency.Pair, price, amount float64) float64 {
	if c.IsCryptoFiatPair() {
		return 0.001 * price * amount
	}
	return 0.002 * price * amount
}
