package cryptocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
)

func mockResponseBody(id int, method string, code int, result interface{}) []byte {
	b, err := json.Marshal(map[string]interface {
	}{
		"id":     id,
		"method": method,
		"code":   code,
		"result": result,
	})
	if err != nil {
		panic(err)
	}
	return b
}
func mockOrderbook(instrumentName string, depth int, bids, asks [][]float64, t int64) OrderbookResult {
	return OrderbookResult{
		InstrumentName: instrumentName,
		Depth:          depth,
		Data: []OrderbookData{{
			Bids: bids,
			Asks: asks,
			T:    t,
		}},
	}
}
func mockCandlestick(instrumentName string, period Interval, depth int, data []Candlestick) CandlestickResult {
	return CandlestickResult{
		InstrumentName: instrumentName,
		Depth:          depth,
		Interval:       period.Encode(),
		Data:           data,
	}
}
func mockTicker(data ...Ticker) TickerResult {
	return TickerResult{
		Data: data,
	}
}
func mockPublicTrades(data ...PublicTrade) PublicTradeResult {
	return PublicTradeResult{
		Data: data,
	}
}

func mockAccounts(data ...AccountSummary) AccountResult {
	return AccountResult{Accounts: data}
}
func mockDepositAddress(data ...DepositAddress) DepositAddressResult {
	return DepositAddressResult{DepositAddressList: data}
}
func mockOrderDetail(info OrderInfo, tradeList ...Trade) OrderDetail {
	return OrderDetail{
		TradeList: tradeList,
		OrderInfo: info,
	}
}
func mockTrades(tradeList ...Trade) TradeResult {
	return TradeResult{
		TradeList: tradeList,
	}
}
func mockOrderHistory(tradeList ...OrderInfo) OrderHistoryResult {
	return OrderHistoryResult{
		OrderList: tradeList,
	}
}
func mockOpenOrders(count int, orders ...OrderInfo) OpenOrdersResult {
	return OpenOrdersResult{
		Count:     count,
		OrderList: orders,
	}
}

func mockWithdrawHistory(list ...Withdraw) WithdrawHistoryResult {
	return WithdrawHistoryResult{
		WithdrawList: list,
	}
}

func mockDepositHistory(list ...Deposit) DepositHistoryResult {
	return DepositHistoryResult{
		DepositList: list,
	}
}

func setupHttpMock(body *mockBody) (cli *Client, mockClient *httpClientMock) {
	mockClient = &httpClientMock{}
	if body != nil {
		mockResponse := &http.Response{
			StatusCode: body.code,
			Body:       ioutil.NopCloser(bytes.NewReader(body.body)),
		}
		mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
	}
	cli = &Client{
		key: "something",
		secret: "something",
		rest: newHttpClient(mockClient,
			fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
		),
	}
	return
}
