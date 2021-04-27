package yobit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/openware/pkg/common/crypto"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
	exchange "github.com/openware/irix"
	"github.com/openware/pkg/request"
)

const (
	apiPublicURL                  = "https://yobit.net/api"
	apiPrivateURL                 = "https://yobit.net/tapi"
	apiPublicVersion              = "3"
	publicInfo                    = "info"
	publicTicker                  = "ticker"
	publicDepth                   = "depth"
	publicTrades                  = "trades"
	privateAccountInfo            = "getInfo"
	privateTrade                  = "Trade"
	privateActiveOrders           = "ActiveOrders"
	privateOrderInfo              = "OrderInfo"
	privateCancelOrder            = "CancelOrder"
	privateTradeHistory           = "TradeHistory"
	privateGetDepositAddress      = "GetDepositAddress"
	privateWithdrawCoinsToAddress = "WithdrawCoinsToAddress"
	privateCreateCoupon           = "CreateYobicode"
	privateRedeemCoupon           = "RedeemYobicode"

	yobitAuthRate   = 0
	yobitUnauthRate = 0
)

// Yobit is the overarching type across the Yobit package
type Yobit struct {
	exchange.Base
}

// GetInfo returns the Yobit info
func (y *Yobit) GetInfo() (Info, error) {
	resp := Info{}
	path := fmt.Sprintf("/%s/%s/", apiPublicVersion, publicInfo)

	return resp, y.SendHTTPRequest(exchange.RestSpot, path, &resp)
}

// GetTicker returns a ticker for a specific currency
func (y *Yobit) GetTicker(symbol string) (map[string]Ticker, error) {
	type Response struct {
		Data map[string]Ticker
	}

	response := Response{}
	path := fmt.Sprintf("/%s/%s/%s", apiPublicVersion, publicTicker, symbol)

	return response.Data, y.SendHTTPRequest(exchange.RestSpot, path, &response.Data)
}

// GetDepth returns the depth for a specific currency
func (y *Yobit) GetDepth(symbol string) (Orderbook, error) {
	type Response struct {
		Data map[string]Orderbook
	}

	response := Response{}
	path := fmt.Sprintf("/%s/%s/%s", apiPublicVersion, publicDepth, symbol)

	return response.Data[symbol],
		y.SendHTTPRequest(exchange.RestSpot, path, &response.Data)
}

// GetTrades returns the trades for a specific currency
func (y *Yobit) GetTrades(symbol string) ([]Trade, error) {
	type respDataHolder struct {
		Data map[string][]Trade
	}

	var dataHolder respDataHolder
	path := "/" + apiPublicVersion + "/" + publicTrades + "/" + symbol
	err := y.SendHTTPRequest(exchange.RestSpot, path, &dataHolder.Data)
	if err != nil {
		return nil, err
	}

	if tr, ok := dataHolder.Data[symbol]; ok {
		return tr, nil
	}
	return nil, nil
}

// GetAccountInformation returns a users account info
func (y *Yobit) GetAccountInformation() (AccountInfo, error) {
	result := AccountInfo{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateAccountInfo, url.Values{}, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// Trade places an order and returns the order ID if successful or an error
func (y *Yobit) Trade(pair, orderType string, amount, price float64) (int64, error) {
	req := url.Values{}
	req.Add("pair", pair)
	req.Add("type", strings.ToLower(orderType))
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("rate", strconv.FormatFloat(price, 'f', -1, 64))

	result := TradeOrderResponse{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateTrade, req, &result)
	if err != nil {
		return int64(result.OrderID), err
	}
	if result.Error != "" {
		return int64(result.OrderID), errors.New(result.Error)
	}
	return int64(result.OrderID), nil
}

// GetOpenOrders returns the active orders for a specific currency
func (y *Yobit) GetOpenOrders(pair string) (map[string]ActiveOrders, error) {
	req := url.Values{}
	req.Add("pair", pair)

	result := map[string]ActiveOrders{}

	return result, y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateActiveOrders, req, &result)
}

// GetOrderInformation returns the order info for a specific order ID
func (y *Yobit) GetOrderInformation(orderID int64) (map[string]OrderInfo, error) {
	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(orderID, 10))

	result := map[string]OrderInfo{}

	return result, y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateOrderInfo, req, &result)
}

// CancelExistingOrder cancels an order for a specific order ID
func (y *Yobit) CancelExistingOrder(orderID int64) error {
	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(orderID, 10))

	result := CancelOrder{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateCancelOrder, req, &result)
	if err != nil {
		return err
	}
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
}

// GetTradeHistory returns the trade history
func (y *Yobit) GetTradeHistory(tidFrom, count, tidEnd, since, end int64, order, pair string) (map[string]TradeHistory, error) {
	req := url.Values{}
	req.Add("from", strconv.FormatInt(tidFrom, 10))
	req.Add("count", strconv.FormatInt(count, 10))
	req.Add("from_id", strconv.FormatInt(tidFrom, 10))
	req.Add("end_id", strconv.FormatInt(tidEnd, 10))
	req.Add("order", order)
	req.Add("since", strconv.FormatInt(since, 10))
	req.Add("end", strconv.FormatInt(end, 10))
	req.Add("pair", pair)

	result := TradeHistoryResponse{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateTradeHistory, req, &result)
	if err != nil {
		return nil, err
	}
	if result.Success == 0 {
		return nil, errors.New(result.Error)
	}

	return result.Data, nil
}

// GetCryptoDepositAddress returns the deposit address for a specific currency
func (y *Yobit) GetCryptoDepositAddress(coin string) (DepositAddress, error) {
	req := url.Values{}
	req.Add("coinName", coin)

	result := DepositAddress{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateGetDepositAddress, req, &result)
	if err != nil {
		return result, err
	}
	if result.Success != 1 {
		return result, fmt.Errorf("%s", result.Error)
	}
	return result, nil
}

// WithdrawCoinsToAddress initiates a withdrawal to a specified address
func (y *Yobit) WithdrawCoinsToAddress(coin string, amount float64, address string) (WithdrawCoinsToAddress, error) {
	req := url.Values{}
	req.Add("coinName", coin)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("address", address)

	result := WithdrawCoinsToAddress{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateWithdrawCoinsToAddress, req, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// CreateCoupon creates an exchange coupon for a sepcific currency
func (y *Yobit) CreateCoupon(currency string, amount float64) (CreateCoupon, error) {
	req := url.Values{}
	req.Add("currency", currency)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	var result CreateCoupon

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateCreateCoupon, req, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// RedeemCoupon redeems an exchange coupon
func (y *Yobit) RedeemCoupon(coupon string) (RedeemCoupon, error) {
	req := url.Values{}
	req.Add("coupon", coupon)

	result := RedeemCoupon{}

	err := y.SendAuthenticatedHTTPRequest(exchange.RestSpotSupplementary, privateRedeemCoupon, req, &result)
	if err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (y *Yobit) SendHTTPRequest(ep exchange.URL, path string, result interface{}) error {
	endpoint, err := y.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	return y.SendPayload(context.Background(), &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       y.Verbose,
		HTTPDebugging: y.HTTPDebugging,
		HTTPRecording: y.HTTPRecording,
	})
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request to Yobit
func (y *Yobit) SendAuthenticatedHTTPRequest(ep exchange.URL, path string, params url.Values, result interface{}) (err error) {
	if !y.AllowAuthenticatedRequest() {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, y.Name)
	}
	endpoint, err := y.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	if params == nil {
		params = url.Values{}
	}

	n := y.Requester.GetNonce(false).String()

	params.Set("nonce", n)
	params.Set("method", path)

	encoded := params.Encode()
	hmac := crypto.GetHMAC(crypto.HashSHA512, []byte(encoded), []byte(y.API.Credentials.Secret))

	if y.Verbose {
		log.Debugf(log.ExchangeSys, "Sending POST request to %s calling path %s with params %s\n",
			endpoint,
			path,
			encoded)
	}

	headers := make(map[string]string)
	headers["Key"] = y.API.Credentials.Key
	headers["Sign"] = crypto.HexEncodeToString(hmac)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	return y.SendPayload(context.Background(), &request.Item{
		Method:        http.MethodPost,
		Path:          endpoint,
		Headers:       headers,
		Body:          strings.NewReader(encoded),
		Result:        result,
		AuthRequest:   true,
		NonceEnabled:  true,
		Verbose:       y.Verbose,
		HTTPDebugging: y.HTTPDebugging,
		HTTPRecording: y.HTTPRecording,
	})
}

// GetFee returns an estimate of fee based on type of transaction
func (y *Yobit) GetFee(feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getWithdrawalFee(feeBuilder.Pair.Base)
	case exchange.InternationalBankDepositFee:
		fee = getInternationalBankDepositFee(feeBuilder.FiatCurrency,
			feeBuilder.BankTransactionType)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.FiatCurrency,
			feeBuilder.Amount,
			feeBuilder.BankTransactionType)
	case exchange.OfflineTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func calculateTradingFee(price, amount float64) (fee float64) {
	return 0.002 * price * amount
}

func getWithdrawalFee(c currency.Code) float64 {
	return WithdrawalFees[c]
}

func getInternationalBankWithdrawalFee(c currency.Code, amount float64, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64

	switch bankTransactionType {
	case exchange.PerfectMoney:
		if c == currency.USD {
			fee = 0.02 * amount
		}
	case exchange.Payeer:
		switch c {
		case currency.USD:
			fee = 0.03 * amount
		case currency.RUR:
			fee = 0.006 * amount
		}
	case exchange.AdvCash:
		switch c {
		case currency.USD:
			fee = 0.04 * amount
		case currency.RUR:
			fee = 0.03 * amount
		}
	case exchange.Qiwi:
		if c == currency.RUR {
			fee = 0.04 * amount
		}
	case exchange.Capitalist:
		if c == currency.USD {
			fee = 0.06 * amount
		}
	}

	return fee
}

// getInternationalBankDepositFee; No real fees for yobit deposits, but want to be explicit on what each payment type supports
func getInternationalBankDepositFee(c currency.Code, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64
	switch bankTransactionType {
	case exchange.PerfectMoney:
		if c == currency.USD {
			fee = 0
		}
	case exchange.Payeer:
		switch c {
		case currency.USD:
			fee = 0
		case currency.RUR:
			fee = 0
		}
	case exchange.AdvCash:
		switch c {
		case currency.USD:
			fee = 0
		case currency.RUR:
			fee = 0
		}
	case exchange.Qiwi:
		if c == currency.RUR {
			fee = 0
		}
	case exchange.Capitalist:
		switch c {
		case currency.USD:
			fee = 0
		case currency.RUR:
			fee = 0
		}
	}

	return fee
}
