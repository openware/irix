package zb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	exchange "github.com/openware/irix"
	"github.com/openware/irix/faker"
	"github.com/openware/irix/portfolio/withdraw"
	"github.com/openware/irix/stream"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/common/convert"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/kline"
	"github.com/openware/pkg/order"
)

// Please supply you own test keys here for due diligence testing.
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
	testCurrency            = "btc_usdt"
)

var z ZB
var wsSetupRan bool

func setupWsAuth(t *testing.T) {
	if wsSetupRan {
		return
	}
	if !z.Websocket.IsEnabled() &&
		!z.API.AuthenticatedWebsocketSupport ||
		!z.ValidateAPICredentials() ||
		!canManipulateRealOrders {
		t.Skip(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := z.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		t.Fatal(err)
	}
	go z.wsReadData()
	wsSetupRan = true
}

func TestSpotNewOrder(t *testing.T) {
	t.Parallel()

	if !z.ValidateAPICredentials() || !canManipulateRealOrders {
		t.Skip()
	}

	arg := SpotNewOrderRequestParams{
		Symbol: testCurrency,
		Type:   SpotNewOrderRequestParamsTypeSell,
		Amount: 0.01,
		Price:  10246.1,
	}
	_, err := z.SpotNewOrder(arg)
	if err != nil {
		t.Errorf("ZB SpotNewOrder: %s", err)
	}
}

func TestCancelExistingOrder(t *testing.T) {
	t.Parallel()

	if !z.ValidateAPICredentials() || !canManipulateRealOrders {
		t.Skip()
	}

	err := z.CancelExistingOrder(20180629145864850, testCurrency)
	if err != nil {
		t.Errorf("ZB CancelExistingOrder: %s", err)
	}
}

func TestGetLatestSpotPrice(t *testing.T) {
	t.Parallel()
	_, err := z.GetLatestSpotPrice(testCurrency)
	if err != nil {
		t.Errorf("ZB GetLatestSpotPrice: %s", err)
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := z.GetTicker(testCurrency)
	if err != nil {
		t.Errorf("ZB GetTicker: %s", err)
	}
}

func TestGetTickers(t *testing.T) {
	t.Parallel()
	_, err := z.GetTickers()
	if err != nil {
		t.Errorf("ZB GetTicker: %s", err)
	}
}

func TestGetOrderbook(t *testing.T) {
	t.Parallel()
	_, err := z.GetOrderbook(testCurrency)
	if err != nil {
		t.Errorf("ZB GetTicker: %s", err)
	}
}

func TestGetMarkets(t *testing.T) {
	t.Parallel()
	_, err := z.GetMarkets()
	if err != nil {
		t.Errorf("ZB GetMarkets: %s", err)
	}
}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:  1,
		FeeType: exchange.CryptocurrencyTradeFee,
		Pair: currency.NewPairWithDelimiter(currency.LTC.String(),
			currency.BTC.String(),
			"-"),
		PurchasePrice:       1,
		FiatCurrency:        currency.USD,
		BankTransactionType: exchange.WireTransfer,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	t.Parallel()
	var feeBuilder = setFeeBuilder()
	z.GetFeeByType(feeBuilder)
	if !z.ValidateAPICredentials() {
		if feeBuilder.FeeType != exchange.OfflineTradeFee {
			t.Errorf("Expected %v, received %v", exchange.OfflineTradeFee, feeBuilder.FeeType)
		}
	} else {
		if feeBuilder.FeeType != exchange.CryptocurrencyTradeFee {
			t.Errorf("Expected %v, received %v", exchange.CryptocurrencyTradeFee, feeBuilder.FeeType)
		}
	}
}

func TestGetFee(t *testing.T) {
	var feeBuilder = setFeeBuilder()

	// CryptocurrencyTradeFee Basic
	if resp, err := z.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
		t.Error(err)
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0.0015), resp)
	}

	// CryptocurrencyTradeFee High quantity
	feeBuilder = setFeeBuilder()
	feeBuilder.Amount = 1000
	feeBuilder.PurchasePrice = 1000
	if resp, err := z.GetFee(feeBuilder); resp != float64(2000) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(2000), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee IsMaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = true
	if resp, err := z.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0.002), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee Negative purchase price
	feeBuilder = setFeeBuilder()
	feeBuilder.PurchasePrice = -1000
	if resp, err := z.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := z.GetFee(feeBuilder); resp != float64(0.005) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0.005), resp)
		t.Error(err)
	}

	// CryptocurrencyWithdrawalFee Invalid currency
	feeBuilder = setFeeBuilder()
	feeBuilder.Pair.Base = currency.NewCode("hello")
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := z.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := z.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	if resp, err := z.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	if resp, err := z.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	expectedResult := exchange.AutoWithdrawCryptoText + " & " + exchange.NoFiatWithdrawalsText
	withdrawPermissions := z.FormatWithdrawPermissions()
	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Received: %s", expectedResult, withdrawPermissions)
	}
}

func TestGetActiveOrders(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	var getOrdersRequest = order.GetOrdersRequest{
		Type: order.AnyType,
		Pairs: []currency.Pair{currency.NewPair(currency.XRP,
			currency.USDT)},
		AssetType: asset.Spot,
	}

	_, err := z.GetActiveOrders(&getOrdersRequest)
	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
}

func TestGetOrderHistory(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	var getOrdersRequest = order.GetOrdersRequest{
		Type:      order.AnyType,
		Side:      order.Buy,
		AssetType: asset.Spot,
		Pairs: []currency.Pair{currency.NewPair(currency.LTC,
			currency.BTC)},
	}

	_, err := z.GetOrderHistory(&getOrdersRequest)
	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------

func TestSubmitOrder(t *testing.T) {
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip(fmt.Sprintf("Can place orders: %v",
			canManipulateRealOrders))
	}
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}

	var orderSubmission = &order.Submit{
		Pair: currency.Pair{
			Delimiter: "_",
			Base:      currency.XRP,
			Quote:     currency.USDT,
		},
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1,
		ClientID:  "meowOrder",
		AssetType: asset.Spot,
	}
	response, err := z.SubmitOrder(orderSubmission)
	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
	if z.ValidateAPICredentials() && response.OrderID == "" {
		t.Error("expected order id")
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}

	currencyPair := currency.NewPair(currency.XRP, currency.USDT)
	var orderCancellation = &order.Cancel{
		ID:            "1",
		WalletAddress: faker.BitcoinTestAddress,
		AccountID:     "1",
		Pair:          currencyPair,
		AssetType:     asset.Spot,
	}

	err := z.CancelOrder(orderCancellation)
	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
}

func TestCancelAllExchangeOrders(t *testing.T) {
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}

	currencyPair := currency.NewPair(currency.XRP, currency.USDT)
	var orderCancellation = &order.Cancel{
		ID:            "1",
		WalletAddress: faker.BitcoinTestAddress,
		AccountID:     "1",
		Pair:          currencyPair,
		AssetType:     asset.Spot,
	}

	resp, err := z.CancelAllOrders(orderCancellation)

	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
	if len(resp.Status) > 0 {
		t.Errorf("%v orders failed to cancel", len(resp.Status))
	}
}

func TestGetAccountInfo(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() {
		_, err := z.UpdateAccountInfo(asset.Spot)
		if err != nil {
			t.Error("GetAccountInfo() error", err)
		}
	} else {
		_, err := z.UpdateAccountInfo(asset.Spot)
		if err == nil {
			t.Error("GetAccountInfo() Expected error")
		}
	}
}

func TestModifyOrder(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}
	_, err := z.ModifyOrder(&order.Modify{AssetType: asset.Spot})
	if err == nil {
		t.Error("ModifyOrder() Expected error")
	}
}

func TestWithdraw(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	withdrawCryptoRequest := withdraw.Request{
		Crypto: withdraw.CryptoRequest{
			Address:   faker.BitcoinTestAddress,
			FeeAmount: 1,
		},
		Amount:      -1,
		Currency:    currency.BTC,
		Description: "WITHDRAW IT ALL",
	}

	_, err := z.WithdrawCryptocurrencyFunds(&withdrawCryptoRequest)
	if z.ValidateAPICredentials() && err != nil {
		t.Error(err)
	} else if !z.ValidateAPICredentials() && err == nil {
		t.Error("expecting an error when no keys are set")
	}
}

func TestWithdrawFiat(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = withdraw.Request{}
	_, err := z.WithdrawFiatFunds(&withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}

func TestWithdrawInternationalBank(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = withdraw.Request{}
	_, err := z.WithdrawFiatFundsToInternationalBank(&withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}

func TestGetDepositAddress(t *testing.T) {
	if mockTests {
		t.Skip("skipping authenticated function for mock testing")
	}
	if z.ValidateAPICredentials() {
		_, err := z.GetDepositAddress(currency.BTC, "")
		if err != nil {
			t.Error("GetDepositAddress() error PLEASE MAKE SURE YOU CREATE DEPOSIT ADDRESSES VIA ZB.COM",
				err)
		}
	} else {
		_, err := z.GetDepositAddress(currency.BTC, "")
		if err == nil {
			t.Error("GetDepositAddress() Expected error")
		}
	}
}

// TestZBInvalidJSON ZB sends poorly formed JSON. this tests the JSON fixer
// Then JSON decode it to test if successful
func TestZBInvalidJSON(t *testing.T) {
	data := `{"success":true,"code":1000,"channel":"getSubUserList","message":"[{"isOpenApi":false,"memo":"Memo","userName":"hello@imgoodthanksandyou.com@good","userId":1337,"isFreez":false}]","no":"0"}`
	fixedJSON := z.wsFixInvalidJSON([]byte(data))
	var response WsGetSubUserListResponse
	err := json.Unmarshal(fixedJSON, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Message[0].UserID != 1337 {
		t.Fatal("Expected extracted JSON USERID to equal 1337")
	}

	data = `{"success":true,"code":1000,"channel":"createSubUserKey","message":"{"apiKey":"thisisnotareallykeyyousillybilly","apiSecret":"lol"}","no":"123"}`
	fixedJSON = z.wsFixInvalidJSON([]byte(data))
	var response2 WsRequestResponse
	err = json.Unmarshal(fixedJSON, &response2)
	if err != nil {
		t.Error(err)
	}
}

// TestWsTransferFunds ws test
func TestWsTransferFunds(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsDoTransferFunds(currency.BTC,
		0.0001,
		"username1",
		"username2",
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestGetSubUserList ws test
func TestGetSubUserList(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsGetSubUserList()
	if err != nil {
		t.Fatal(err)
	}
}

// TestAddSubUser ws test
func TestAddSubUser(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsAddSubUser("1", "123456789101112aA!")
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsCreateSuUserKey ws test
func TestWsCreateSuUserKey(t *testing.T) {
	setupWsAuth(t)
	subUsers, err := z.wsGetSubUserList()
	if err != nil {
		t.Fatal(err)
	}
	if len(subUsers.Message) == 0 {
		t.Skip("User ID required for test to continue. Create a subuser first")
	}
	userID := subUsers.Message[0].UserID
	_, err = z.wsCreateSubUserKey(true, true, true, true, "subu", strconv.FormatInt(userID, 10))
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsSubmitOrder ws test
func TestWsSubmitOrder(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsSubmitOrder(currency.NewPairWithDelimiter(currency.LTC.String(), currency.BTC.String(), "").Lower(), 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsCancelOrder ws test
func TestWsCancelOrder(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsCancelOrder(currency.NewPairWithDelimiter(currency.LTC.String(), currency.BTC.String(), "").Lower(), 1234)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsGetAccountInfo ws test
func TestWsGetAccountInfo(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsGetAccountInfoRequest()
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsGetOrder ws test
func TestWsGetOrder(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsGetOrder(currency.NewPairWithDelimiter(currency.LTC.String(), currency.BTC.String(), "").Lower(), 1234)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsGetOrders ws test
func TestWsGetOrders(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsGetOrders(currency.NewPairWithDelimiter(currency.LTC.String(), currency.BTC.String(), "").Lower(), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWsGetOrdersIgnoreTradeType ws test
func TestWsGetOrdersIgnoreTradeType(t *testing.T) {
	setupWsAuth(t)
	_, err := z.wsGetOrdersIgnoreTradeType(currency.NewPairWithDelimiter(currency.LTC.String(), currency.BTC.String(), "").Lower(), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWsMarketConfig(t *testing.T) {
	pressXToJSON := []byte(`{
    "code":1000,
    "data":{
        "btc_usdt":{
            "amountScale":4,
            "priceScale":2
            },
        "bcc_usdt":{
            "amountScale":3,
            "priceScale":2
            }
    },
    "success":true,
    "channel":"markets",
    "message":"操作成功。"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsTicker(t *testing.T) {
	pressXToJSON := []byte(`{
    "channel": "ltcbtc_ticker",
    "date": "1472800466093",
    "no": "1337",
    "ticker": {
        "buy": "3826.94",
        "high": "3838.22",
        "last": "3826.94",
        "low": "3802.0",
        "sell": "3828.25",
        "vol": "90151.83"
    }
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsOrderbook(t *testing.T) {
	pressXToJSON := []byte(`{
    "asks": [
        [
            3846.94,
            0.659
        ]
    ],
    "bids": [
        [
            3826.94,
            4.843
        ]
    ],
    "channel": "ltcbtc_depth",
    "no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsTrades(t *testing.T) {
	pressXToJSON := []byte(`{"data":[{"date":1581473835,"amount":"13.620","price":"242.89","trade_type":"bid","type":"buy","tid":703896035},{"date":1581473835,"amount":"0.156","price":"242.89","trade_type":"bid","type":"buy","tid":703896036}],"dataType":"trades","channel":"ethusdt_trades"}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsPlaceOrderJSON(t *testing.T) {
	pressXToJSON := []byte(`{"message":"操作成功。","no":"1337","data":"{"entrustId":201711133673}","code":1000,"channel":"btcusdt_order","success":true}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsCancelOrderJSON(t *testing.T) {
	pressXToJSON := []byte(`{
    "success": true,
    "code": 1000,
    "channel": "ltcbtc_cancelorder",
    "message": "操作成功。",
    "no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsGetOrderJSON(t *testing.T) {
	pressXToJSON := []byte(`{
    "success": true,
    "code": 1000,
    "data": {
        "currency": "ltc_btc",
        "id": "20160902387645980",
        "price": 100,
        "status": 0,
        "total_amount": 0.01,
        "trade_amount": 0,
        "trade_date": 1472814905567,
        "trade_money": 0,
        "type": 1
    },
    "channel": "ltcbtc_getorder",
    "message": "操作成功。",
    "no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsGetOrdersJSON(t *testing.T) {
	pressXToJSON := []byte(`{
    "success": true,
    "code": 1000,
    "data": [
        {
           "currency": "ltc_btc",
           "id": "20160901385862136",
           "price": 3700,
           "status": 0,
           "total_amount": 1.845,
           "trade_amount": 0,
           "trade_date": 1472706387742,
           "trade_money": 0,
           "type": 1
        }
    ],
    "channel": "ltcbtc_getorders",
    "message": "操作成功。",
    "no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsGetOrderIgnoreTypeJSON(t *testing.T) {
	pressXToJSON := []byte(`{
    "success": true,
    "code": 1000,
    "data": [
        {
            "currency": "ltc_btc",
            "id": "20160901385862136",
            "price": 3700,
            "status": 0,
            "total_amount": 1.845,
            "trade_amount": 0,
            "trade_date": 1472706387742,
            "trade_money": 0,
            "type": 1
        }
    ],
    "channel": "ltcbtc_getordersignoretradetype",
    "message": "操作成功。",
    "no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsGetUserInfo(t *testing.T) {
	pressXToJSON := []byte(`{
    "message": "操作成功",
    "no": "15207605119",
    "data": {
        "coins": [
            {
                "freez": "1.35828369",
                "enName": "BTC",
                "unitDecimal": 8,
                "cnName": "BTC",
                "unitTag": "฿",
                "available": "0.72771906",
                "key": "btc"
            },
            {
                "freez": "0.011",
                "enName": "LTC",
                "unitDecimal": 8,
                "cnName": "LTC",
                "unitTag": "Ł",
                "available": "3.51859814",
                "key": "ltc"
            }
        ],
        "base": {
            "username": "15207605119",
            "trade_password_enabled": true,
            "auth_google_enabled": true,
            "auth_mobile_enabled": true
        }
    },
    "code": 1000,
    "channel": "getaccountinfo",
    "success": true
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsGetSubUsersResponse(t *testing.T) {
	pressXToJSON := []byte(`{"success": true,"code": 1000,"channel": "getSubUserList","message": "[{"isOpenApi": false,"memo": "1","userName": "15914665280@1","userId": 110980,"isFreez": false}, {"isOpenApi": false,"memo": "2","userName": "15914665280@2","userId": 110984,"isFreez": false}, {"isOpenApi": false,"memo": "test3","userName": "15914665280@3","userId": 111014,"isFreez": false}]","no": "0"}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsCreateSubUserResponse(t *testing.T) {
	pressXToJSON := []byte(`{
	"success": true,
	"code": 1000,
	"channel": "createSubUserKey",
	"message": "{"apiKey ":"41 bf75f9 - 525e-4876 - 8257 - b880a938d4d2 ","apiSecret ":"046 b4706fe88b5728991274962d7fc46b4779c0c"}",
	"no": "1337"
}`)
	err := z.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestGetSpotKline(t *testing.T) {
	arg := KlinesRequestParams{
		Symbol: testCurrency,
		Type:   kline.OneMin.Short() + "in",
		Size:   int64(z.Features.Enabled.Kline.ResultLimit),
	}
	if mockTests {
		startTime := time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)
		arg.Since = convert.UnixMillis(startTime)
		arg.Type = "1day"
	}

	_, err := z.GetSpotKline(arg)
	if err != nil {
		t.Errorf("ZB GetSpotKline: %s", err)
	}
}

func TestGetHistoricCandles(t *testing.T) {
	currencyPair, err := currency.NewPairFromString(testCurrency)
	if err != nil {
		t.Fatal(err)
	}

	startTime := time.Now().Add(-time.Hour * 1)
	endTime := time.Now()
	if mockTests {
		startTime = time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(2020, 9, 2, 0, 0, 0, 0, time.UTC)
	}

	_, err = z.GetHistoricCandles(currencyPair, asset.Spot, startTime, endTime, kline.OneDay)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.GetHistoricCandles(currencyPair, asset.Spot, startTime, endTime, kline.Interval(time.Hour*7))
	if err == nil {
		t.Fatal("unexpected result")
	}
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	currencyPair, err := currency.NewPairFromString(testCurrency)
	if err != nil {
		t.Fatal(err)
	}
	startTime := time.Now().Add(-time.Hour * 1)
	endTime := time.Now()
	if mockTests {
		startTime = time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(2020, 9, 2, 0, 0, 0, 0, time.UTC)
	}
	_, err = z.GetHistoricCandlesExtended(currencyPair, asset.Spot, startTime, endTime, kline.OneDay)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_FormatExchangeKlineInterval(t *testing.T) {
	testCases := []struct {
		name     string
		interval kline.Interval
		output   string
	}{
		{
			"OneMin",
			kline.OneMin,
			"1min",
		},
		{
			"OneHour",
			kline.OneHour,
			"1hour",
		},
		{
			"OneDay",
			kline.OneDay,
			"1day",
		},
		{
			"ThreeDay",
			kline.ThreeDay,
			"3day",
		},
		{
			"OneWeek",
			kline.OneWeek,
			"1week",
		},
		{
			"AllOther",
			kline.FifteenDay,
			"",
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			ret := z.FormatExchangeKlineInterval(test.interval)

			if ret != test.output {
				t.Fatalf("unexpected result return expected: %v received: %v", test.output, ret)
			}
		})
	}
}

func TestValidateCandlesRequest(t *testing.T) {
	_, err := z.validateCandlesRequest(currency.Pair{}, "", time.Time{}, time.Time{}, kline.Interval(-1))
	if err != nil && err.Error() != "invalid time range supplied. Start: 0001-01-01 00:00:00 +0000 UTC End 0001-01-01 00:00:00 +0000 UTC" {
		t.Error(err)
	}
	_, err = z.validateCandlesRequest(currency.Pair{}, "", time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC), time.Time{}, kline.Interval(-1))
	if err != nil && err.Error() != "invalid time range supplied. Start: 2020-01-01 01:01:01.000000001 +0000 UTC End 0001-01-01 00:00:00 +0000 UTC" {
		t.Error(err)
	}
	_, err = z.validateCandlesRequest(currency.Pair{}, asset.Spot, time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC), time.Date(2020, 1, 1, 1, 1, 1, 3, time.UTC), kline.OneHour)
	if err != nil && err.Error() != "pair not enabled" {
		t.Error(err)
	}
	var p currency.Pair
	p, err = currency.NewPairFromString(testCurrency)
	if err != nil {
		t.Fatal(err)
	}
	var item kline.Item
	item, err = z.validateCandlesRequest(p, asset.Spot, time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC), time.Date(2020, 1, 1, 1, 1, 1, 3, time.UTC), kline.OneHour)
	if err != nil {
		t.Error(err)
	}
	if !item.Pair.Equal(p) {
		t.Errorf("unexpected result, expected %v, received %v", p, item.Pair)
	}
	if item.Asset != asset.Spot {
		t.Errorf("unexpected result, expected %v, received %v", asset.Spot, item.Asset)
	}
	if item.Interval != kline.OneHour {
		t.Errorf("unexpected result, expected %v, received %v", kline.OneHour, item.Interval)
	}
	if item.Exchange != z.Name {
		t.Errorf("unexpected result, expected %v, received %v", z.Name, item.Exchange)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()

	trades, err := z.GetTrades("btc_usdt")
	if err != nil {
		t.Error(err)
	}
	if len(trades) == 0 {
		t.Error("expected results")
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()

	currencyPair, err := currency.NewPairFromString("btc_usdt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.GetRecentTrades(currencyPair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("btc_usdt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.GetHistoricTrades(currencyPair, asset.Spot, time.Now().Add(-time.Minute*15), time.Now())
	if err != nil && err != common.ErrFunctionNotSupported {
		t.Error(err)
	}
}
