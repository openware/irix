package cryptocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/openware/irix/faker"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)


func TestClient_RestGetInstruments(t *testing.T) {
	method := publicGetInstruments
	testCases := []struct {
		responseBody     []byte
		responseCode     int
		totalInstruments int
		shouldError      bool
	}{
		// invalid cases
		{mockResponseBody(1, method, 10004, nil), 400, 0, true},
		{mockResponseBody(1, method, 10001, nil), 500, 0, true},
		// valid cases
		{mockResponseBody(1, method, 0, InstrumentResult{Instruments: nil}), 200, 0, false},
		{mockResponseBody(1, method, 0, InstrumentResult{Instruments: []Instruments{
			{
				InstrumentName:       "BTC_USDT",
				QuoteCurrency:        "BTC",
				BaseCurrency:         "USDT",
				PriceDecimals:        7,
				QuantityDecimals:     3,
				MarginTradingEnabled: false,
			},
		}}), 200, 1, false},
	}
	for _, r := range testCases {
		cli, mockClient := setupHttpMock(&mockBody{r.responseCode, r.responseBody})
		res, err := cli.RestGetInstruments()
		mockClient.AssertExpectations(t)
		if r.shouldError {
			assert.Nil(t, res, r)
			assert.NotNil(t, err, r)
		} else {
			assert.Nil(t, err, r)
			assert.Len(t, res, r.totalInstruments, r)
			if r.totalInstruments > 0 {
				assert.NotNil(t, res, r)
			}
		}
	}
}

func TestClient_RestGetOrderbook(t *testing.T) {
	method := publicGetBook

	testCases := []struct {
		instrumentName        string
		depth                 int
		expectedDepth         int
		responseBody          []byte
		responseCode          int
		shouldErrorValidation bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", 0, 150, nil, 400, true, false},
		{"BTC_", 0, 150, nil, 400, true, false},
		{"_", 1, 1, nil, 400, true, false},
		{"", 1, 1, nil, 400, true, false},
		{"BTC_USDT", 151, 151, mockResponseBody(1, method, 10001, nil), 500, true, false},
		{"BTC_USDT", -100, -100, mockResponseBody(1, method, 10001, nil), 500, true, false},
		// invalid cases
		{"BTC_USDT", 1, 1, mockResponseBody(1, method, 10004, nil), 400, false, true},
		{"BTC_USDT", 1, 1, mockResponseBody(1, method, 10001, nil), 500, false, true},
		// valid cases
		{"BTC_USDT", 0, 150, mockResponseBody(1, method, 0, mockOrderbook("BTC_USDT", 150, [][]float64{{float64(1), float64(2), float64(2)}}, [][]float64{{float64(1), float64(2), float64(2)}}, time.Now().Unix())), 200, false, false},
		{"BTC_USDT", 1, 1, mockResponseBody(1, method, 0, mockOrderbook("BTC_USDT", 1, [][]float64{{float64(1), float64(2), float64(2)}}, [][]float64{{float64(1), float64(2), float64(2)}}, time.Now().Unix())), 200, false, false},
		{"BTC_USDT", 3, 3, mockResponseBody(1, method, 0, mockOrderbook("BTC_USDT", 3, [][]float64{{float64(1), float64(2), float64(2)}}, [][]float64{{float64(1), float64(2), float64(2)}}, time.Now().Unix())), 200, false, false},
	}
	for _, r := range testCases {
		cli, mockClient := setupHttpMock(&mockBody{r.responseCode, r.responseBody})
		res, err := cli.RestGetOrderBook(1, r.instrumentName, r.depth)
		if r.shouldErrorValidation {
			mockClient.AssertNotCalled(t, "Do")
		} else {
			req := mockClient.Calls[0].Arguments[0].(*http.Request)
			assert.Contains(t, req.URL.Path, method)
			assert.Equal(t, strconv.Itoa(r.expectedDepth), req.URL.Query().Get("depth"))
			assert.Equal(t, r.instrumentName, req.URL.Query().Get("instrument_name"))
			mockClient.AssertExpectations(t)
		}
		if r.shouldError {
			assert.NotNil(t, err, r)
		} else if !r.shouldErrorValidation && !r.shouldError {
			assert.Nil(t, err, r)
			assert.Equal(t, r.instrumentName, res.InstrumentName)
			for _, b := range res.Data {
				for _, bid := range b.Bids {
					assert.Len(t, bid, 3)
				}
				for _, ask := range b.Asks {
					assert.Len(t, ask, 3)
				}
			}
		}
	}
}
func TestClient_RestGetCandlestick(t *testing.T) {
	method := publicGetCandlestick

	testCases := []struct {
		instrumentName        string
		depth                 int
		interval              Interval
		responseBody          []byte
		responseCode          int
		shouldErrorValidation bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", 0, Minute1, nil, 400, true, false},
		{"BTC_", 0, Minute1, nil, 400, true, false},
		{"_", 1, Minute1, nil, 400, true, false},
		{"", 1, Minute1, nil, 400, true, false},
		{"BTC_USDT", 151, 151, mockResponseBody(1, method, 10001, nil), 500, true, false},
		{"BTC_USDT", -100, -100, mockResponseBody(1, method, 10001, nil), 500, true, false},
		// invalid cases
		{"BTC_USDT", 1, 1, mockResponseBody(1, method, 10004, nil), 400, false, true},
		{"BTC_USDT", 1, 1, mockResponseBody(1, method, 10001, nil), 500, false, true},
		// valid cases
		{"BTC_USDT", 0, Month, mockResponseBody(1, method, 0, mockCandlestick("BTC_USDT", Month, 1000, []Candlestick{{time.Now().Unix(), 1, 1, 1, 1, 1}})), 200, false, false},
		{"BTC_USDT", 10, Week, mockResponseBody(1, method, 0, mockCandlestick("BTC_USDT", Week, 1000, []Candlestick{{time.Now().Unix(), 1, 1, 1, 1, 1}})), 200, false, false},
	}
	for _, r := range testCases {
		cli, mockClient := setupHttpMock(&mockBody{r.responseCode, r.responseBody})
		res, err := cli.RestGetCandlestick(r.instrumentName, r.interval, r.depth)
		if r.shouldErrorValidation {
			mockClient.AssertNotCalled(t, "Do")
		} else {
			req := mockClient.Calls[0].Arguments[0].(*http.Request)
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.Path, method)
			if r.depth > 0 {
				assert.Equal(t, strconv.Itoa(r.depth), req.URL.Query().Get("depth"))
			}
			assert.Equal(t, r.interval.Encode(), req.URL.Query().Get("interval"))
			assert.Equal(t, r.instrumentName, req.URL.Query().Get("instrument_name"))
			mockClient.AssertExpectations(t)
		}
		if r.shouldError {
			assert.NotNil(t, err, r)
		} else if !r.shouldErrorValidation && !r.shouldError {
			assert.Nil(t, err, r)
			assert.Equal(t, r.instrumentName, res.InstrumentName)
			assert.NotNil(t, res.Data)
		}
	}
}

func TestClient_GetTicker(t *testing.T) {
	method := publicGetTicker

	testCases := []struct {
		instrumentName        string
		responseBody          []byte
		responseCode          int
		shouldErrorValidation bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", nil, 400, true, false},
		{"BTC_", nil, 400, true, false},
		{"_", nil, 400, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		// invalid cases
		{"BTC_USDT", mockResponseBody(1, method, 10004, nil), 400, false, true},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, false, true},
		// valid cases
		{"", mockResponseBody(1, method, 0, mockTicker(Ticker{
			Instrument: "BTC_USDT",
			Bid:        0,
			Ask:        0,
			Trade:      0,
			Timestamp:  0,
			Volume:     0,
			Highest:    0,
			Lowest:     0,
			Change:     0,
		}, Ticker{
			Instrument: "ETC_BTC",
			Bid:        0,
			Ask:        0,
			Trade:      0,
			Timestamp:  0,
			Volume:     0,
			Highest:    0,
			Lowest:     0,
			Change:     0,
		})), 200, false, false},
		{"BTC_USDT", mockResponseBody(1, method, 0, mockTicker(Ticker{
			Instrument: "BTC_USDT",
			Bid:        1,
			Ask:        1,
			Trade:      1,
			Timestamp:  time.Now().Unix(),
			Volume:     1,
			Highest:    1,
			Lowest:     1,
			Change:     1,
		})), 200, false, false},
		{"BTC_USDT", mockResponseBody(1, method, 0, mockTicker(Ticker{
			Instrument: "BTC_USDT",
			Bid:        1,
			Ask:        1,
			Trade:      1,
			Timestamp:  time.Now().Unix(),
			Volume:     1,
			Highest:    1,
			Lowest:     1,
			Change:     1,
		})), 200, false, false},
	}
	for _, r := range testCases {
		cli, mockClient := setupHttpMock(&mockBody{r.responseCode, r.responseBody})
		res, err := cli.RestGetTicker(r.instrumentName)
		if r.shouldErrorValidation {
			mockClient.AssertNotCalled(t, "Do")
		} else {
			req := mockClient.Calls[0].Arguments[0].(*http.Request)
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.Path, method)
			assert.Equal(t, r.instrumentName, req.URL.Query().Get("instrument_name"))
			mockClient.AssertExpectations(t)
		}
		if r.shouldError {
			assert.NotNil(t, err, r)
		} else if !r.shouldErrorValidation && !r.shouldError {
			assert.Nil(t, err, r)
			assert.NotNil(t, res.Data)
		}
	}
}

func TestClient_GetPublicTrades(t *testing.T) {
	method := publicGetTrades

	testCases := []struct {
		instrumentName        string
		responseBody          []byte
		responseCode          int
		shouldErrorValidation bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", nil, 400, true, false},
		{"BTC_", nil, 400, true, false},
		{"_", nil, 400, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		// invalid cases
		{"BTC_USDT", mockResponseBody(1, method, 10004, nil), 400, false, true},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, false, true},
		// valid cases
		{"", mockResponseBody(1, method, 0, mockPublicTrades(PublicTrade{
			Instrument: "BTC_USDT",
			Quantity:   1,
			Price:      0.1,
			Side:       "BUY",
			Timestamp:  time.Now().Unix(),
			TradeID:    1,
		}, PublicTrade{
			Instrument: "ETC_USDT",
			Quantity:   1,
			Price:      0.2,
			Side:       "SELL",
			Timestamp:  time.Now().Unix(),
			TradeID:    2,
		})), 200, false, false},
		{"BTC_USDT", mockResponseBody(1, method, 0, mockPublicTrades(PublicTrade{
			Instrument: "BTC_USDT",
			Quantity:   1,
			Price:      0.2,
			Side:       "SELL",
			Timestamp:  time.Now().Unix(),
			TradeID:    2,
		})), 200, false, false},
		{"BTC_USDT", mockResponseBody(1, method, 0, mockPublicTrades(PublicTrade{
			Instrument: "BTC_USDT",
			Quantity:   1,
			Price:      0.2,
			Side:       "BUY",
			Timestamp:  time.Now().Unix(),
			TradeID:    2,
		})), 200, false, false},
	}
	for _, r := range testCases {
		mockClient := &httpClientMock{}
		mockResponse := &http.Response{
			StatusCode: r.responseCode,
			Body:       ioutil.NopCloser(bytes.NewReader(r.responseBody)),
		}
		mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
		cli := &Client{rest: newHttpClient(mockClient,
			fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
		)}
		res, err := cli.RestGetPublicTrades(r.instrumentName)
		if r.shouldErrorValidation {
			mockClient.AssertNotCalled(t, "Do")
		} else {
			req := mockClient.Calls[0].Arguments[0].(*http.Request)
			assert.Equal(t, "GET", req.Method)
			assert.Contains(t, req.URL.Path, method)
			assert.Equal(t, r.instrumentName, req.URL.Query().Get("instrument_name"))
			mockClient.AssertExpectations(t)
		}
		if r.shouldError {
			assert.NotNil(t, err, r)
		} else if !r.shouldErrorValidation && !r.shouldError {
			assert.Nil(t, err, r)
			assert.NotNil(t, res.Data)
		}
	}
}

func TestClient_RestCreateWithdrawal(t *testing.T)  {
	method := privateCreateWithdrawal
	testTable := []struct{
		reqID int
		in WithdrawParams
		body *mockBody
		expectedParams KVParams
		shouldValidationError bool
		shouldError bool
	}{
		{0, WithdrawParams{}, nil, nil, true, false},
		{0, WithdrawParams{Currency: "BTC_USDT"}, nil, nil, true, false},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress}, &mockBody{400, mockResponseBody(-1, method, 10004, Withdraw{})}, KVParams{"currency": "BTC", "amount": 0.1, "address": faker.BitcoinTestAddress}, false, true},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress, WithdrawID: "withdrawid"}, &mockBody{400, mockResponseBody(-1, method, 10004, Withdraw{})}, KVParams{"currency": "BTC", "amount": 0.1, "address": faker.BitcoinTestAddress, "client_wid": "withdrawid"}, false, true},
		{0, WithdrawParams{Currency: "BTC", Amount: 0.1, Address: faker.BitcoinTestAddress, WithdrawID: "withdrawid", AddressTag: "XRP"}, &mockBody{200, mockResponseBody(-1, method, 0, Withdraw{})}, KVParams{"currency": "BTC", "amount": 0.1, "address": faker.BitcoinTestAddress, "client_wid": "withdrawid", "address_tag": "XRP"}, false, false},
	}
	for _, c := range testTable {
		cli, mockClient := setupHttpMock(c.body)
		res, err := cli.RestCreateWithdrawal(c.reqID, c.in)
		if c.shouldValidationError {
			assert.Nil(t, res)
			assert.NotNil(t, err)
			mockClient.AssertNotCalled(t, "Do")
			continue
		}
		mockClient.AssertExpectations(t)
		req := mockClient.Calls[0].Arguments[0].(*http.Request)
		b, _ := ioutil.ReadAll(req.Body)
		var rq *Request
		_ = json.Unmarshal(b, &rq)
		if c.reqID > 0 {
			assert.Equal(t, c.reqID, rq.Id)
		}
		b1, _ := json.Marshal(rq.Params)
		b2, _ := json.Marshal(c.expectedParams)
		assert.Equal(t, method, rq.Method)
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, string(b1), string(b2))
		assert.NotEmpty(t, rq.Nonce)
		assert.NotEmpty(t, rq.Signature)
		assert.Equal(t, "something", rq.ApiKey)
		assert.Contains(t, req.URL.String(), method)
		if c.shouldError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, res, c)
		}
	}
}
func TestClient_RestGetWithdrawalHistory(t *testing.T)  {
	method := privateGetWithdrawalHistory
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testTable := []struct{
		reqID int
		in *WithdrawHistoryParam
		body *mockBody
		expectedParams KVParams
		shouldValidationError bool
		shouldError bool
	}{
		{0, &WithdrawHistoryParam{StartTS: 2, EndTS: 1}, nil, nil, true, false},
		{0, &WithdrawHistoryParam{Currency: "BTC_USDT"}, nil, nil, true, false},
		{0, nil, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{}, false, true},
		{0, &WithdrawHistoryParam{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", StartTS: timeAgo[0]}, &mockBody{400, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0]}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0]}, &mockBody{500, mockResponseBody(-1, method, 10005, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1]}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1}, &mockBody{500, mockResponseBody(-1, method, 10004, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1}, false, true},
		{0, &WithdrawHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1, Status: WithdrawCancelled}, &mockBody{200, mockResponseBody(-1, method, 0, mockWithdrawHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1, "status": "6"}, false, false},
	}
	for _, c := range testTable {
		cli, mockClient := setupHttpMock(c.body)
		res, err := cli.RestGetWithdrawalHistory(c.reqID, c.in)
		if c.shouldValidationError {
			assert.Nil(t, res)
			assert.NotNil(t, err)
			mockClient.AssertNotCalled(t, "Do")
			continue
		}
		mockClient.AssertExpectations(t)
		req := mockClient.Calls[0].Arguments[0].(*http.Request)
		b, _ := ioutil.ReadAll(req.Body)
		var rq *Request
		_ = json.Unmarshal(b, &rq)
		if c.reqID > 0 {
			assert.Equal(t, c.reqID, rq.Id)
		}
		b1, _ := json.Marshal(rq.Params)
		b2, _ := json.Marshal(c.expectedParams)
		assert.Equal(t, method, rq.Method)
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, string(b1), string(b2))
		assert.NotEmpty(t, rq.Nonce)
		assert.NotEmpty(t, rq.Signature)
		assert.Equal(t, "something", rq.ApiKey)
		assert.Contains(t, req.URL.String(), method)
		if c.shouldError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, res, c)
		}
	}
}
func TestClient_GetDepositAddress(t *testing.T) {
	method := privateGetDepositAddress

	testCases := []struct {
		instrumentName        string
		responseBody          []byte
		responseCode          int
		shouldErrorValidation bool
		shouldError           bool
	}{
		// invalid arguments
		{"_USDT", nil, 400, true, false},
		{"BTC_", nil, 400, true, false},
		{"_", nil, 400, true, false},
		{"", nil, 400, true, false},
		// invalid cases
		{"BTC_USDT", mockResponseBody(1, method, 10004, nil), 400, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		// valid cases
		{"BTC", mockResponseBody(1, method, 0, mockDepositAddress(DepositAddress{
			Currency: "BTC",
			Network:  "CRO",
		})), 200, false, false},
		{"USDT", mockResponseBody(1, method, 0, mockDepositAddress(DepositAddress{
			Currency: "BTC",
			Network:  "CRO",
		})), 200, false, false},
	}
	for _, r := range testCases {
		mockClient := &httpClientMock{}
		mockResponse := &http.Response{
			StatusCode: r.responseCode,
			Body:       ioutil.NopCloser(bytes.NewReader(r.responseBody)),
		}
		mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
		cli := &Client{
			key:    "something",
			secret: "something",
			rest: newHttpClient(mockClient,
				fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
			)}
		res, err := cli.RestGetDepositAddress(r.instrumentName)
		if r.shouldErrorValidation {
			mockClient.AssertNotCalled(t, "Do")
		} else {
			assert.Len(t, mockClient.Calls, 1, r)
			req := mockClient.Calls[0].Arguments[0].(*http.Request)
			b, _ := ioutil.ReadAll(req.Body)
			var body map[string]interface{}
			_ = json.Unmarshal(b, &body)
			params := body["params"].(map[string]interface{})
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.Path, method)
			if r.instrumentName != "" {
				assert.Equal(t, r.instrumentName, params["currency"])
			} else {
				assert.Equal(t, map[string]interface{}{}, params)
			}
			assert.NotEmpty(t, body["api_key"])
			assert.NotEmpty(t, body["sig"])
			mockClient.AssertExpectations(t)
		}
		if r.shouldError {
			assert.NotNil(t, err, r)
		} else if !r.shouldErrorValidation && !r.shouldError {
			assert.Nil(t, err, r)
			assert.NotNil(t, res.DepositAddressList)
		}
	}
}

func TestClient_RestGetDepositHistory(t *testing.T)  {
	method := privateGetDepositHistory
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testTable := []struct{
		reqID int
		in *DepositHistoryParam
		body *mockBody
		expectedParams KVParams
		shouldValidationError bool
		shouldError bool
	}{
		{0, &DepositHistoryParam{StartTS: 2, EndTS: 1}, nil, nil, true, false},
		{0, &DepositHistoryParam{Currency: "BTC_USDT"}, nil, nil, true, false},
		{0, nil, &mockBody{400, mockResponseBody(-1, method, 10004, mockDepositHistory())}, KVParams{}, false, true},
		{0, &DepositHistoryParam{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockDepositHistory())}, KVParams{}, false, true},
		{0, &DepositHistoryParam{Currency: "BTC", StartTS: timeAgo[0]}, &mockBody{400, mockResponseBody(-1, method, 10004, mockDepositHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0]}, false, true},
		{0, &DepositHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0]}, &mockBody{500, mockResponseBody(-1, method, 10005, mockDepositHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1]}, false, true},
		{0, &DepositHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1}, &mockBody{500, mockResponseBody(-1, method, 10004, mockDepositHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1}, false, true},
		{0, &DepositHistoryParam{Currency: "BTC", EndTS: timeAgo[1], StartTS: timeAgo[0], PageSize: 10, Page: 1, Status: DepositFailed}, &mockBody{200, mockResponseBody(-1, method, 0, mockDepositHistory())}, KVParams{"currency": "BTC", "start_ts": timeAgo[0], "end_ts": timeAgo[1], "page_size": 10, "page": 1, "status": "2"}, false, false},
	}
	for _, c := range testTable {
		cli, mockClient := setupHttpMock(c.body)
		res, err := cli.RestGetDepositHistory(c.reqID, c.in)
		if c.shouldValidationError {
			assert.Nil(t, res)
			assert.NotNil(t, err)
			mockClient.AssertNotCalled(t, "Do")
			continue
		}
		mockClient.AssertExpectations(t)
		req := mockClient.Calls[0].Arguments[0].(*http.Request)
		b, _ := ioutil.ReadAll(req.Body)
		var rq *Request
		_ = json.Unmarshal(b, &rq)
		if c.reqID > 0 {
			assert.Equal(t, c.reqID, rq.Id)
		}
		b1, _ := json.Marshal(rq.Params)
		b2, _ := json.Marshal(c.expectedParams)
		assert.Equal(t, method, rq.Method)
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, string(b1), string(b2))
		assert.NotEmpty(t, rq.Nonce)
		assert.NotEmpty(t, rq.Signature)
		assert.Equal(t, "something", rq.ApiKey)
		assert.Contains(t, req.URL.String(), method)
		if c.shouldError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, res, c)
		}
	}
}
