package coinbasepro

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	exchange "github.com/openware/irix"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/common/crypto"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
	"github.com/openware/pkg/order"
	"github.com/openware/pkg/request"
)

const (
	coinbaseproAPIURL                  = "https://api.pro.coinbase.com/"
	coinbaseproSandboxAPIURL           = "https://api-public.sandbox.pro.coinbase.com/"
	coinbaseproAPIVersion              = "0"
	coinbaseproProducts                = "products"
	coinbaseproOrderbook               = "book"
	coinbaseproTicker                  = "ticker"
	coinbaseproTrades                  = "trades"
	coinbaseproHistory                 = "candles"
	coinbaseproStats                   = "stats"
	coinbaseproCurrencies              = "currencies"
	coinbaseproAccounts                = "accounts"
	coinbaseproLedger                  = "ledger"
	coinbaseproHolds                   = "holds"
	coinbaseproOrders                  = "orders"
	coinbaseproFills                   = "fills"
	coinbaseproTransfers               = "transfers"
	coinbaseproReports                 = "reports"
	coinbaseproTime                    = "time"
	coinbaseproMarginTransfer          = "profiles/margin-transfer"
	coinbaseproPosition                = "position"
	coinbaseproPositionClose           = "position/close"
	coinbaseproPaymentMethod           = "payment-methods"
	coinbaseproPaymentMethodDeposit    = "deposits/payment-method"
	coinbaseproDepositCoinbase         = "deposits/coinbase-account"
	coinbaseproWithdrawalPaymentMethod = "withdrawals/payment-method"
	coinbaseproWithdrawalCoinbase      = "withdrawals/coinbase"
	coinbaseproWithdrawalCrypto        = "withdrawals/crypto"
	coinbaseproCoinbaseAccounts        = "coinbase-accounts"
	coinbaseproTrailingVolume          = "users/self/trailing-volume"
)

// CoinbasePro is the overarching type across the coinbasepro package
type CoinbasePro struct {
	exchange.Base
}

// GetProducts returns supported currency pairs on the exchange with specific
// information about the pair
func (c *CoinbasePro) GetProducts() ([]Product, error) {
	var products []Product

	return products, c.SendHTTPRequest(exchange.RestSpot, coinbaseproProducts, &products)
}

// GetOrderbook returns orderbook by currency pair and level
func (c *CoinbasePro) GetOrderbook(symbol string, level int) (interface{}, error) {
	orderbook := OrderbookResponse{}

	path := fmt.Sprintf("%s/%s/%s", coinbaseproProducts, symbol, coinbaseproOrderbook)
	if level > 0 {
		levelStr := strconv.Itoa(level)
		path = fmt.Sprintf("%s/%s/%s?level=%s", coinbaseproProducts, symbol, coinbaseproOrderbook, levelStr)
	}

	if err := c.SendHTTPRequest(exchange.RestSpot, path, &orderbook); err != nil {
		return nil, err
	}

	if level == 3 {
		ob := OrderbookL3{}
		ob.Sequence = orderbook.Sequence
		for _, x := range orderbook.Asks {
			price, err := strconv.ParseFloat((x[0].(string)), 64)
			if err != nil {
				continue
			}
			amount, err := strconv.ParseFloat((x[1].(string)), 64)
			if err != nil {
				continue
			}

			ob.Asks = append(ob.Asks, OrderL3{Price: price, Amount: amount, OrderID: x[2].(string)})
		}
		for _, x := range orderbook.Bids {
			price, err := strconv.ParseFloat((x[0].(string)), 64)
			if err != nil {
				continue
			}
			amount, err := strconv.ParseFloat((x[1].(string)), 64)
			if err != nil {
				continue
			}

			ob.Bids = append(ob.Bids, OrderL3{Price: price, Amount: amount, OrderID: x[2].(string)})
		}
		return ob, nil
	}
	ob := OrderbookL1L2{}
	ob.Sequence = orderbook.Sequence
	for _, x := range orderbook.Asks {
		price, err := strconv.ParseFloat((x[0].(string)), 64)
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat((x[1].(string)), 64)
		if err != nil {
			continue
		}

		ob.Asks = append(ob.Asks, OrderL1L2{Price: price, Amount: amount, NumOrders: x[2].(float64)})
	}
	for _, x := range orderbook.Bids {
		price, err := strconv.ParseFloat((x[0].(string)), 64)
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat((x[1].(string)), 64)
		if err != nil {
			continue
		}

		ob.Bids = append(ob.Bids, OrderL1L2{Price: price, Amount: amount, NumOrders: x[2].(float64)})
	}
	return ob, nil
}

// GetTicker returns ticker by currency pair
// currencyPair - example "BTC-USD"
func (c *CoinbasePro) GetTicker(currencyPair string) (Ticker, error) {
	tick := Ticker{}
	path := fmt.Sprintf(
		"%s/%s/%s", coinbaseproProducts, currencyPair, coinbaseproTicker)
	return tick, c.SendHTTPRequest(exchange.RestSpot, path, &tick)
}

// GetTrades listd the latest trades for a product
// currencyPair - example "BTC-USD"
func (c *CoinbasePro) GetTrades(currencyPair string) ([]Trade, error) {
	var trades []Trade
	path := fmt.Sprintf(
		"%s/%s/%s", coinbaseproProducts, currencyPair, coinbaseproTrades)
	return trades, c.SendHTTPRequest(exchange.RestSpot, path, &trades)
}

// GetHistoricRates returns historic rates for a product. Rates are returned in
// grouped buckets based on requested granularity.
func (c *CoinbasePro) GetHistoricRates(currencyPair, start, end string, granularity int64) ([]History, error) {
	var resp [][]interface{}
	var history []History
	values := url.Values{}

	if len(start) > 0 {
		values.Set("start", start)
	} else {
		values.Set("start", "")
	}

	if len(end) > 0 {
		values.Set("end", end)
	} else {
		values.Set("end", "")
	}

	allowedGranularities := [6]int64{60, 300, 900, 3600, 21600, 86400}
	validGran, _ := common.InArray(granularity, allowedGranularities)
	if !validGran {
		return nil, errors.New("Invalid granularity value: " + strconv.FormatInt(granularity, 10) + ". Allowed values are {60, 300, 900, 3600, 21600, 86400}")
	}
	if granularity > 0 {
		values.Set("granularity", strconv.FormatInt(granularity, 10))
	}

	path := common.EncodeURLValues(
		fmt.Sprintf("%s/%s/%s", coinbaseproProducts, currencyPair, coinbaseproHistory),
		values)

	if err := c.SendHTTPRequest(exchange.RestSpot, path, &resp); err != nil {
		return history, err
	}

	for _, single := range resp {
		var s History
		a, _ := single[0].(float64)
		s.Time = int64(a)
		b, _ := single[1].(float64)
		s.Low = b
		c, _ := single[2].(float64)
		s.High = c
		d, _ := single[3].(float64)
		s.Open = d
		e, _ := single[4].(float64)
		s.Close = e
		f, _ := single[5].(float64)
		s.Volume = f
		history = append(history, s)
	}

	return history, nil
}

// GetStats returns a 24 hr stat for the product. Volume is in base currency
// units. open, high, low are in quote currency units.
func (c *CoinbasePro) GetStats(currencyPair string) (Stats, error) {
	stats := Stats{}
	path := fmt.Sprintf(
		"%s/%s/%s", coinbaseproProducts, currencyPair, coinbaseproStats)

	return stats, c.SendHTTPRequest(exchange.RestSpot, path, &stats)
}

// GetCurrencies returns a list of supported currency on the exchange
// Warning: Not all currencies may be currently in use for tradinc.
func (c *CoinbasePro) GetCurrencies() ([]Currency, error) {
	var currencies []Currency

	return currencies, c.SendHTTPRequest(exchange.RestSpot, coinbaseproCurrencies, &currencies)
}

// GetServerTime returns the API server time
func (c *CoinbasePro) GetServerTime() (ServerTime, error) {
	serverTime := ServerTime{}

	return serverTime, c.SendHTTPRequest(exchange.RestSpot, coinbaseproTime, &serverTime)
}

// GetAccounts returns a list of trading accounts associated with the APIKEYS
func (c *CoinbasePro) GetAccounts() ([]AccountResponse, error) {
	var resp []AccountResponse

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, coinbaseproAccounts, nil, &resp)
}

// GetAccount returns information for a single account. Use this endpoint when
// account_id is known
func (c *CoinbasePro) GetAccount(accountID string) (AccountResponse, error) {
	resp := AccountResponse{}
	path := fmt.Sprintf("%s/%s", coinbaseproAccounts, accountID)

	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// GetAccountHistory returns a list of account activity. Account activity either
// increases or decreases your account balance. Items are paginated and sorted
// latest first.
func (c *CoinbasePro) GetAccountHistory(accountID string) ([]AccountLedgerResponse, error) {
	var resp []AccountLedgerResponse
	path := fmt.Sprintf("%s/%s/%s", coinbaseproAccounts, accountID, coinbaseproLedger)

	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// GetHolds returns the holds that are placed on an account for any active
// orders or pending withdraw requests. As an order is filled, the hold amount
// is updated. If an order is canceled, any remaining hold is removed. For a
// withdraw, once it is completed, the hold is removed.
func (c *CoinbasePro) GetHolds(accountID string) ([]AccountHolds, error) {
	var resp []AccountHolds
	path := fmt.Sprintf("%s/%s/%s", coinbaseproAccounts, accountID, coinbaseproHolds)

	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// PlaceLimitOrder places a new limit order. Orders can only be placed if the
// account has sufficient funds. Once an order is placed, account funds
// will be put on hold for the duration of the order. How much and which funds
// are put on hold depends on the order type and parameters specified.
//
// GENERAL PARAMS
// clientRef - [optional] Order ID selected by you to identify your order
// side - 	buy or sell
// productID - A valid product id
// stp - [optional] Self-trade prevention flag
//
// LIMIT ORDER PARAMS
// price - Price per bitcoin
// amount - Amount of BTC to buy or sell
// timeInforce - [optional] GTC, GTT, IOC, or FOK (default is GTC)
// cancelAfter - [optional] min, hour, day * Requires time_in_force to be GTT
// postOnly - [optional] Post only flag Invalid when time_in_force is IOC or FOK
func (c *CoinbasePro) PlaceLimitOrder(clientRef string, price, amount float64, side, timeInforce, cancelAfter, productID, stp string, postOnly bool) (string, error) {
	resp := GeneralizedOrderResponse{}
	req := make(map[string]interface{})
	req["type"] = order.Limit.Lower()
	req["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	req["size"] = strconv.FormatFloat(amount, 'f', -1, 64)
	req["side"] = side
	req["product_id"] = productID

	if cancelAfter != "" {
		req["cancel_after"] = cancelAfter
	}
	if timeInforce != "" {
		req["time_in_foce"] = timeInforce
	}
	if clientRef != "" {
		req["client_oid"] = clientRef
	}
	if stp != "" {
		req["stp"] = stp
	}
	if postOnly {
		req["post_only"] = postOnly
	}

	err := c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproOrders, req, &resp)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// PlaceMarketOrder places a new market order.
// Orders can only be placed if the account has sufficient funds. Once an order
// is placed, account funds will be put on hold for the duration of the order.
// How much and which funds are put on hold depends on the order type and
// parameters specified.
//
// GENERAL PARAMS
// clientRef - [optional] Order ID selected by you to identify your order
// side - 	buy or sell
// productID - A valid product id
// stp - [optional] Self-trade prevention flag
//
// MARKET ORDER PARAMS
// size - [optional]* Desired amount in BTC
// funds	[optional]* Desired amount of quote currency to use
// * One of size or funds is required.
func (c *CoinbasePro) PlaceMarketOrder(clientRef string, size, funds float64, side, productID, stp string) (string, error) {
	resp := GeneralizedOrderResponse{}
	req := make(map[string]interface{})
	req["side"] = side
	req["product_id"] = productID
	req["type"] = order.Market.Lower()

	if size != 0 {
		req["size"] = strconv.FormatFloat(size, 'f', -1, 64)
	}
	if funds != 0 {
		req["funds"] = strconv.FormatFloat(funds, 'f', -1, 64)
	}
	if clientRef != "" {
		req["client_oid"] = clientRef
	}
	if stp != "" {
		req["stp"] = stp
	}

	err := c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproOrders, req, &resp)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// PlaceMarginOrder places a new market order.
// Orders can only be placed if the account has sufficient funds. Once an order
// is placed, account funds will be put on hold for the duration of the order.
// How much and which funds are put on hold depends on the order type and
// parameters specified.
//
// GENERAL PARAMS
// clientRef - [optional] Order ID selected by you to identify your order
// side - 	buy or sell
// productID - A valid product id
// stp - [optional] Self-trade prevention flag
//
// MARGIN ORDER PARAMS
// size - [optional]* Desired amount in BTC
// funds - [optional]* Desired amount of quote currency to use
func (c *CoinbasePro) PlaceMarginOrder(clientRef string, size, funds float64, side, productID, stp string) (string, error) {
	resp := GeneralizedOrderResponse{}
	req := make(map[string]interface{})
	req["side"] = side
	req["product_id"] = productID
	req["type"] = "margin"

	if size != 0 {
		req["size"] = strconv.FormatFloat(size, 'f', -1, 64)
	}
	if funds != 0 {
		req["funds"] = strconv.FormatFloat(funds, 'f', -1, 64)
	}
	if clientRef != "" {
		req["client_oid"] = clientRef
	}
	if stp != "" {
		req["stp"] = stp
	}

	err := c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproOrders, req, &resp)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// CancelExistingOrder cancels order by orderID
func (c *CoinbasePro) CancelExistingOrder(orderID string) error {
	path := fmt.Sprintf("%s/%s", coinbaseproOrders, orderID)

	return c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodDelete, path, nil, nil)
}

// CancelAllExistingOrders cancels all open orders on the exchange and returns
// and array of order IDs
// currencyPair - [optional] all orders for a currencyPair string will be
// canceled
func (c *CoinbasePro) CancelAllExistingOrders(currencyPair string) ([]string, error) {
	var resp []string
	req := make(map[string]interface{})

	if len(currencyPair) > 0 {
		req["product_id"] = currencyPair
	}
	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodDelete, coinbaseproOrders, req, &resp)
}

// GetOrders lists current open orders. Only open or un-settled orders are
// returned. As soon as an order is no longer open and settled, it will no
// longer appear in the default request.
// status - can be a range of "open", "pending", "done" or "active"
// currencyPair - [optional] for example "BTC-USD"
func (c *CoinbasePro) GetOrders(status []string, currencyPair string) ([]GeneralizedOrderResponse, error) {
	var resp []GeneralizedOrderResponse
	params := url.Values{}

	for _, individualStatus := range status {
		params.Add("status", individualStatus)
	}
	if currencyPair != "" {
		params.Set("product_id", currencyPair)
	}

	path := common.EncodeURLValues(coinbaseproOrders, params)
	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// GetOrder returns a single order by order id.
func (c *CoinbasePro) GetOrder(orderID string) (GeneralizedOrderResponse, error) {
	resp := GeneralizedOrderResponse{}
	path := fmt.Sprintf("%s/%s", coinbaseproOrders, orderID)

	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// GetFills returns a list of recent fills
func (c *CoinbasePro) GetFills(orderID, currencyPair string) ([]FillResponse, error) {
	var resp []FillResponse
	params := url.Values{}

	if orderID != "" {
		params.Set("order_id", orderID)
	}
	if currencyPair != "" {
		params.Set("product_id", currencyPair)
	}
	if params.Get("order_id") == "" && params.Get("product_id") == "" {
		return resp, errors.New("no parameters set")
	}

	path := common.EncodeURLValues(coinbaseproFills, params)
	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// MarginTransfer sends funds between a standard/default profile and a margin
// profile.
// A deposit will transfer funds from the default profile into the margin
// profile. A withdraw will transfer funds from the margin profile to the
// default profile. Withdraws will fail if they would set your margin ratio
// below the initial margin ratio requirement.
//
// amount - the amount to transfer between the default and margin profile
// transferType - either "deposit" or "withdraw"
// profileID - The id of the margin profile to deposit or withdraw from
// currency - currency to transfer, currently on "BTC" or "USD"
func (c *CoinbasePro) MarginTransfer(amount float64, transferType, profileID, currency string) (MarginTransfer, error) {
	resp := MarginTransfer{}
	req := make(map[string]interface{})
	req["type"] = transferType
	req["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	req["currency"] = currency
	req["margin_profile_id"] = profileID

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproMarginTransfer, req, &resp)
}

// GetPosition returns an overview of account profile.
func (c *CoinbasePro) GetPosition() (AccountOverview, error) {
	resp := AccountOverview{}

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, coinbaseproPosition, nil, &resp)
}

// ClosePosition closes a position and allowing you to repay position as well
// repayOnly -  allows the position to be repaid
func (c *CoinbasePro) ClosePosition(repayOnly bool) (AccountOverview, error) {
	resp := AccountOverview{}
	req := make(map[string]interface{})
	req["repay_only"] = repayOnly

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproPositionClose, req, &resp)
}

// GetPayMethods returns a full list of payment methods
func (c *CoinbasePro) GetPayMethods() ([]PaymentMethod, error) {
	var resp []PaymentMethod

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, coinbaseproPaymentMethod, nil, &resp)
}

// DepositViaPaymentMethod deposits funds from a payment method. See the Payment
// Methods section for retrieving your payment methods.
//
// amount - The amount to deposit
// currency - The type of currency
// paymentID - ID of the payment method
func (c *CoinbasePro) DepositViaPaymentMethod(amount float64, currency, paymentID string) (DepositWithdrawalInfo, error) {
	resp := DepositWithdrawalInfo{}
	req := make(map[string]interface{})
	req["amount"] = amount
	req["currency"] = currency
	req["payment_method_id"] = paymentID

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproPaymentMethodDeposit, req, &resp)
}

// DepositViaCoinbase deposits funds from a coinbase account. Move funds between
// a Coinbase account and coinbasepro trading account within daily limits. Moving
// funds between Coinbase and coinbasepro is instant and free. See the Coinbase
// Accounts section for retrieving your Coinbase accounts.
//
// amount - The amount to deposit
// currency - The type of currency
// accountID - ID of the coinbase account
func (c *CoinbasePro) DepositViaCoinbase(amount float64, currency, accountID string) (DepositWithdrawalInfo, error) {
	resp := DepositWithdrawalInfo{}
	req := make(map[string]interface{})
	req["amount"] = amount
	req["currency"] = currency
	req["coinbase_account_id"] = accountID

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproDepositCoinbase, req, &resp)
}

// WithdrawViaPaymentMethod withdraws funds to a payment method
//
// amount - The amount to withdraw
// currency - The type of currency
// paymentID - ID of the payment method
func (c *CoinbasePro) WithdrawViaPaymentMethod(amount float64, currency, paymentID string) (DepositWithdrawalInfo, error) {
	resp := DepositWithdrawalInfo{}
	req := make(map[string]interface{})
	req["amount"] = amount
	req["currency"] = currency
	req["payment_method_id"] = paymentID

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproWithdrawalPaymentMethod, req, &resp)
}

// /////////////////////// NO ROUTE FOUND ERROR ////////////////////////////////
// WithdrawViaCoinbase withdraws funds to a coinbase account.
//
// amount - The amount to withdraw
// currency - The type of currency
// accountID - 	ID of the coinbase account
// func (c *CoinbasePro) WithdrawViaCoinbase(amount float64, currency, accountID string) (DepositWithdrawalInfo, error) {
// 	resp := DepositWithdrawalInfo{}
// 	req := make(map[string]interface{})
// 	req["amount"] = amount
// 	req["currency"] = currency
// 	req["coinbase_account_id"] = accountID
//
// 	return resp,
// 		c.SendAuthenticatedHTTPRequest(http.MethodPost, coinbaseproWithdrawalCoinbase, req, &resp)
// }

// WithdrawCrypto withdraws funds to a crypto address
//
// amount - The amount to withdraw
// currency - The type of currency
// cryptoAddress - 	A crypto address of the recipient
func (c *CoinbasePro) WithdrawCrypto(amount float64, currency, cryptoAddress string) (DepositWithdrawalInfo, error) {
	resp := DepositWithdrawalInfo{}
	req := make(map[string]interface{})
	req["amount"] = amount
	req["currency"] = currency
	req["crypto_address"] = cryptoAddress

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproWithdrawalCrypto, req, &resp)
}

// GetCoinbaseAccounts returns a list of coinbase accounts
func (c *CoinbasePro) GetCoinbaseAccounts() ([]CoinbaseAccounts, error) {
	var resp []CoinbaseAccounts

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, coinbaseproCoinbaseAccounts, nil, &resp)
}

// GetReport returns batches of historic information about your account in
// various human and machine readable forms.
//
// reportType - "fills" or "account"
// startDate - Starting date for the report (inclusive)
// endDate - Ending date for the report (inclusive)
// currencyPair - ID of the product to generate a fills report for.
// E.c. BTC-USD. *Required* if type is fills
// accountID - ID of the account to generate an account report for. *Required*
// if type is account
// format - 	pdf or csv (default is pdf)
// email - [optional] Email address to send the report to
func (c *CoinbasePro) GetReport(reportType, startDate, endDate, currencyPair, accountID, format, email string) (Report, error) {
	resp := Report{}
	req := make(map[string]interface{})
	req["type"] = reportType
	req["start_date"] = startDate
	req["end_date"] = endDate
	req["format"] = "pdf"

	if currencyPair != "" {
		req["product_id"] = currencyPair
	}
	if accountID != "" {
		req["account_id"] = accountID
	}
	if format == "csv" {
		req["format"] = format
	}
	if email != "" {
		req["email"] = email
	}

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodPost, coinbaseproReports, req, &resp)
}

// GetReportStatus once a report request has been accepted for processing, the
// status is available by polling the report resource endpoint.
func (c *CoinbasePro) GetReportStatus(reportID string) (Report, error) {
	resp := Report{}
	path := fmt.Sprintf("%s/%s", coinbaseproReports, reportID)

	return resp, c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, path, nil, &resp)
}

// GetTrailingVolume this request will return your 30-day trailing volume for
// all products.
func (c *CoinbasePro) GetTrailingVolume() ([]Volume, error) {
	var resp []Volume

	return resp,
		c.SendAuthenticatedHTTPRequest(exchange.RestSpot, http.MethodGet, coinbaseproTrailingVolume, nil, &resp)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (c *CoinbasePro) SendHTTPRequest(ep exchange.URL, path string, result interface{}) error {
	endpoint, err := c.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	return c.SendPayload(context.Background(), &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       c.Verbose,
		HTTPDebugging: c.HTTPDebugging,
		HTTPRecording: c.HTTPRecording,
	})
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (c *CoinbasePro) SendAuthenticatedHTTPRequest(ep exchange.URL, method, path string, params map[string]interface{}, result interface{}) (err error) {
	if !c.AllowAuthenticatedRequest() {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet,
			c.Name)
	}
	endpoint, err := c.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	payload := []byte("")

	if params != nil {
		payload, err = json.Marshal(params)
		if err != nil {
			return errors.New("sendAuthenticatedHTTPRequest: Unable to JSON request")
		}

		if c.Verbose {
			log.Debugf(log.ExchangeSys, "Request JSON: %s\n", payload)
		}
	}

	now := time.Now()
	n := strconv.FormatInt(now.Unix(), 10)
	message := n + method + "/" + path + string(payload)
	hmac := crypto.GetHMAC(crypto.HashSHA256, []byte(message), []byte(c.API.Credentials.Secret))
	headers := make(map[string]string)
	headers["CB-ACCESS-SIGN"] = crypto.Base64Encode(hmac)
	headers["CB-ACCESS-TIMESTAMP"] = n
	headers["CB-ACCESS-KEY"] = c.API.Credentials.Key
	headers["CB-ACCESS-PASSPHRASE"] = c.API.Credentials.ClientID
	headers["Content-Type"] = "application/json"

	// Timestamp must be within 30 seconds of the api service time
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(30*time.Second))
	defer cancel()
	return c.SendPayload(ctx, &request.Item{
		Method:        method,
		Path:          endpoint + path,
		Headers:       headers,
		Body:          bytes.NewBuffer(payload),
		Result:        result,
		AuthRequest:   true,
		Verbose:       c.Verbose,
		HTTPDebugging: c.HTTPDebugging,
		HTTPRecording: c.HTTPRecording,
	})
}

// GetFee returns an estimate of fee based on type of transaction
func (c *CoinbasePro) GetFee(feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		trailingVolume, err := c.GetTrailingVolume()
		if err != nil {
			return 0, err
		}
		fee = c.calculateTradingFee(trailingVolume,
			feeBuilder.Pair.Base,
			feeBuilder.Pair.Quote,
			feeBuilder.Pair.Delimiter,
			feeBuilder.PurchasePrice,
			feeBuilder.Amount,
			feeBuilder.IsMaker)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.FiatCurrency)
	case exchange.InternationalBankDepositFee:
		fee = getInternationalBankDepositFee(feeBuilder.FiatCurrency)
	case exchange.OfflineTradeFee:
		fee = getOfflineTradeFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

// getOfflineTradeFee calculates the worst case-scenario trading fee
func getOfflineTradeFee(price, amount float64) float64 {
	return 0.0025 * price * amount
}

func (c *CoinbasePro) calculateTradingFee(trailingVolume []Volume, base, quote currency.Code, delimiter string, purchasePrice, amount float64, isMaker bool) float64 {
	var fee float64
	for _, i := range trailingVolume {
		if strings.EqualFold(i.ProductID, base.String()+delimiter+quote.String()) {
			switch {
			case isMaker:
				fee = 0
			case i.Volume <= 10000000:
				fee = 0.003
			case i.Volume > 10000000 && i.Volume <= 100000000:
				fee = 0.002
			case i.Volume > 100000000:
				fee = 0.001
			}
			break
		}
	}
	return fee * amount * purchasePrice
}

func getInternationalBankWithdrawalFee(c currency.Code) float64 {
	var fee float64

	if c == currency.USD {
		fee = 25
	} else if c == currency.EUR {
		fee = 0.15
	}

	return fee
}

func getInternationalBankDepositFee(c currency.Code) float64 {
	var fee float64

	if c == currency.USD {
		fee = 10
	} else if c == currency.EUR {
		fee = 0.15
	}

	return fee
}
