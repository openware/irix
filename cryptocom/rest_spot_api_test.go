package cryptocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestClient_PrivateGetAccountSummary(t *testing.T) {
	method := privateGetAccountSummary

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
		{"BTC_USDT", mockResponseBody(1, method, 10004, nil), 400, true, false},
		{"BTC_USDT", mockResponseBody(1, method, 10001, nil), 500, true, false},
		{"BTC", mockResponseBody(1, method, 10004, nil), 500, false, true},
		{"USDT", mockResponseBody(1, method, 10001, nil), 500, false, true},
		// valid cases
		{"", mockResponseBody(1, method, 0, mockAccounts(AccountSummary{
			Balance:   0,
			Available: 0,
			Order:     0,
			Stake:     0,
			Currency:  "BTC",
		})), 200, false, false},
		{"USDT", mockResponseBody(1, method, 0, mockAccounts(AccountSummary{
			Balance:   0,
			Available: 0,
			Order:     0,
			Stake:     0,
			Currency:  "USDT",
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
		res, err := cli.RestGetAccountSummary(r.instrumentName)
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
			assert.NotNil(t, res.Accounts)
		}
	}
}
func TestClient_RestCreateOrder(t *testing.T)  {

}
func TestClient_RestCancelOrder(t *testing.T)  {

}
func TestClient_RestCancelAllOrder(t *testing.T) {

}
func TestClient_RestGetOrderHistory(t *testing.T) {

}
func TestClient_RestGetOpenOrders(t *testing.T) {
	t.Parallel()
	method := privateGetOpenOrders
	testTable := []struct{
		reqID int
		param *OpenOrderParam
		expectedParams KVParams
		body *mockBody
		shouldValidationError    bool
		shouldError bool
	}{
		{0, &OpenOrderParam{"-", 0, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC", 0, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC_USDT", -1, 0}, nil, nil, true, false},
		{0, &OpenOrderParam{"BTC_USDT", 0, -1}, nil, nil, true, false},
		// valid values
		{0, nil, KVParams{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{}, KVParams{}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{"BTC_USDT", 0, 0}, KVParams{"instrument_name": "BTC_USDT"}, &mockBody{400, mockResponseBody(-1, method, 10004, mockOpenOrders(0))}, false, true},
		{0, &OpenOrderParam{"BTC_USDT", 10, 0}, KVParams{"instrument_name": "BTC_USDT", "page_size": 10}, &mockBody{200, mockResponseBody(-1, method, 0, mockOpenOrders(1, OrderInfo{InstrumentName: "BTC_USDT", Status: "ACTIVE"}))}, false, false},
		{0, &OpenOrderParam{"BTC_USDT", 10, 1}, KVParams{"instrument_name": "BTC_USDT", "page_size": 10, "page": 1}, &mockBody{200, mockResponseBody(-1, method, 0, mockOpenOrders(1, OrderInfo{InstrumentName: "BTC_USDT", Status: "ACTIVE"}))}, false, false},
	}
	for _, c := range testTable {
		cli, mockClient := setupHttpMock(c.body)
		res, err := cli.RestOpenOrders(c.reqID, c.param)
		if c.shouldValidationError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
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
func TestClient_RestGetOrderDetails(t *testing.T) {
	t.Parallel()
	method := privateGetOrderDetail
	type input struct {
		reqID    int
		remoteID string
	}
	testTable := []struct {
		in                    input
		body                  *mockBody
		expectedBody          KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{input{0, "0"}, nil, nil, true, false},
		{input{0, "1212121212"}, &mockBody{400, mockResponseBody(-1, method, 10004, nil)}, KVParams{"order_id": "1212121212"}, false, true},
		{input{1213, "1212121212"}, &mockBody{200, mockResponseBody(1213, method, 0, mockOrderDetail(OrderInfo{}, Trade{}))}, KVParams{"order_id": "1212121212"}, false, false},
	}
	for _, c := range testTable {
		mockClient := &httpClientMock{}
		if c.body != nil {
			mockResponse := &http.Response{
				StatusCode: c.body.code,
				Body:       ioutil.NopCloser(bytes.NewReader(c.body.body)),
			}
			mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
		}
		cli := &Client{
			key:    "something",
			secret: "something",
			rest: newHttpClient(mockClient,
				fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
			)}
		res, err := cli.RestGetOrderDetails(c.in.reqID, c.in.remoteID)
		if c.shouldValidationError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
			continue
		}
		mockClient.AssertExpectations(t)
		req := mockClient.Calls[0].Arguments[0].(*http.Request)
		b, _ := ioutil.ReadAll(req.Body)
		var rq *Request
		_ = json.Unmarshal(b, &rq)
		if c.in.reqID > 0 {
			assert.Equal(t, c.in.reqID, rq.Id)
		}
		assert.Equal(t, method, rq.Method)
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, c.expectedBody, rq.Params)
		assert.NotEmpty(t, rq.Nonce)
		assert.NotEmpty(t, rq.Signature)
		assert.Equal(t, "something", rq.ApiKey)
		assert.Contains(t, req.URL.String(), privateGetOrderDetail)
		if c.shouldError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
		} else {
			assert.Nil(t, err, c)
			assert.NotNil(t, res, c)
		}
	}
}

func TestClient_RestGetTrades(t *testing.T) {
	t.Parallel()
	method := privateGetTrades
	type input struct {
		reqID int
		body  *TradeParams
	}
	timeAgo := []int64{timestampMs(time.Now().Add(time.Hour * -24)), timestampMs(time.Now().Add(time.Second * -5)), timestampMs(time.Now().Add(time.Hour * -23))}
	testTable := []struct {
		in                    input
		body                  *mockBody
		expectedParams        KVParams
		shouldValidationError bool
		shouldError           bool
	}{
		{input{0, &TradeParams{Market: "BTC"}}, nil, nil, true, false},
		{input{0, nil}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[0]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[0]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", StartTS: timeAgo[1]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "start_ts": timeAgo[1]}, false, true},
		{input{0, &TradeParams{Market: "BTC_USDT", EndTS: timeAgo[2]}}, &mockBody{400, mockResponseBody(0, method, 10004, nil)}, KVParams{"instrument_name": "BTC_USDT", "end_ts": timeAgo[2]}, false, true},
		{input{1212, &TradeParams{Page: 20}}, &mockBody{200, mockResponseBody(1212, method, 0, mockTrades())}, KVParams{"page": 20}, false, false},
		{input{121212, &TradeParams{PageSize: 1}}, &mockBody{200, mockResponseBody(1212, method, 0, mockTrades())}, KVParams{"page_size": 1}, false, false},
	}
	for _, c := range testTable {
		mockClient := &httpClientMock{}
		if c.body != nil {
			mockResponse := &http.Response{
				StatusCode: c.body.code,
				Body:       ioutil.NopCloser(bytes.NewReader(c.body.body)),
			}
			mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
		}
		cli := &Client{
			key:    "something",
			secret: "something",
			rest: newHttpClient(mockClient,
				fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
			)}
		res, err := cli.RestGetTrades(c.in.reqID, c.in.body)
		if c.shouldValidationError {
			assert.NotNil(t, err)
			assert.Nil(t, res)
			continue
		}
		mockClient.AssertExpectations(t)
		req := mockClient.Calls[0].Arguments[0].(*http.Request)
		b, _ := ioutil.ReadAll(req.Body)
		var rq *Request
		_ = json.Unmarshal(b, &rq)
		if c.in.reqID > 0 {
			assert.Equal(t, c.in.reqID, rq.Id)
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


