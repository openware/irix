package cryptocom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockHTTPClientError struct {
	response *http.Response
	endpoint string
}

func (m *mockHTTPClientError) Post(endpoint, contentType string, body io.Reader) (resp *http.Response, err error) {
	return nil, errors.New("")
}

func (m *mockHTTPClientError) Get(endpoint string) (resp *http.Response, err error) {
	return nil, errors.New("")
}

type testRestFunc func(client *Client) (Response, error)

func testRest(t *testing.T, expectEndPoint string, jsonExpected string, fn testRestFunc) {
	// prepare mock
	var endpoint string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonExpected))
		endpoint = r.URL.String()
	}))

	defer ts.Close()

	client := New(ts.URL, ts.URL, "test", "test")

	privateResponse := make(chan *bytes.Buffer)
	publicResponse := make(chan *bytes.Buffer)
	client.connectMock(privateResponse, publicResponse, bytes.NewBuffer(nil), bytes.NewBuffer(nil))

	// test function
	resp, _ := fn(client)

	// prepare expect
	var expectedResponse Response
	_ = json.NewDecoder(bytes.NewBufferString(jsonExpected)).Decode(&expectedResponse)

	// assert response
	assert.NotEqual(t, Response{}, resp)
	assert.Equal(t, expectedResponse.Method, resp.Method)
	assert.Equal(t, expectedResponse.Result, resp.Result)
	assert.Equal(t, expectedResponse.Code, resp.Code)
	assert.Equal(t, expectedResponse.Message, resp.Message)

	// assert endpoint
	assert.Equal(t, expectEndPoint, endpoint)
}

func TestRestGetOrderDetails(t *testing.T) {
	remoteID := "1138210129647637539"

	t.Run("Success", func(t *testing.T) {
		// mock response
		jsonStr := `{"id": 1,
      "method": "private/get-order-detail",
      "code": 0,
      "result": { 
        "trade_list": [],
        "order_info": {
          "avg_price": 0.01,
          "client_oid": "2238094a-ec65-4ba6-8c9f-49597723c7fe",
          "create_time": 1611750064904,
          "cumulative_quantity": 0.0001,
          "cumulative_value": 0.000001,
          "exec_inst": "",
          "fee_currency": "ETH",
          "instrument_name": "ETH_CRO",
          "order_id": "1137940341134421889",
          "price": 0.01,
          "quantity": 0.0001,
          "side": "sel",
          "status": "FILLED",
          "time_in_force": "GOOD_TILL_CANCEL",
          "type": "LIMIT",
          "update_time": 1611750065006
        }
      }}`
		expectedEndpoint := `/v2/private/get-order-detail`
		testRest(t,
			expectedEndpoint,
			jsonStr,
			func(client *Client) (Response, error) {
				return client.RestGetOrderDetails(1, remoteID)
			},
		)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		client := &Client{}
		client.httpClient = &mockHTTPClientError{}
		response, err := client.RestGetOrderDetails(1, remoteID)
		assert.Equal(t, Response{}, response)
		assert.Equal(t, errors.New(""), err)
	})
}

func TestRestGetBalance(t *testing.T) {
	reqID := 1

	t.Run("Success", func(t *testing.T) {
		// mock response
		jsonStr := fmt.Sprintf(`{"id": %d,
      "method": "private/get-account-summary",
      "code": 0,
      "result": {
        "accounts": [
          {
            "balance": 99999999.905000000000000000,
            "available": 99999996.905000000000000000,
            "order": 3.000000000000000000,
            "stake": 0,
            "currency": "CRO"
          },
          {
            "available": 1000000000,
            "balance": 1000000000,
            "currency": "BTC",
            "order": 0,
            "stake": 0
          }
        ]
      }}`,
			reqID,
		)
		expectedEndpoint := `/v2/private/get-account-summary`
		testRest(t,
			expectedEndpoint,
			jsonStr,
			func(client *Client) (Response, error) { return client.RestGetBalance(reqID) },
		)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		client := &Client{}
		client.httpClient = &mockHTTPClientError{}
		response, err := client.RestGetBalance(reqID)
		assert.Equal(t, Response{}, response)
		assert.Equal(t, errors.New(""), err)
	})
}

func TestRestGetTrades(t *testing.T) {
	// mock response
	reqID := 1
	market := "ETH_CRO"

	t.Run("Success", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`
      {
        "id": %d,
        "method": "private/get-trades",
        "code": 0,
        "result": {
          "trade_list": [
            {
              "side": "SELL",
              "instrument_name": "%s",
              "fee": 0.014,
              "trade_id": "367107655537806900",
              "create_time": 1588777459755,
              "traded_price": 7,
              "traded_quantity": 1,
              "fee_currency": "CRO",
              "order_id": "367107623521528450"
            }
          ]
        }
      }`,
			reqID,
			market,
		)
		expectedEndpoint := `/v2/private/get-trades`
		testRest(t,
			expectedEndpoint,
			jsonStr,
			func(client *Client) (Response, error) { return client.RestGetTrades(reqID, market) },
		)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		client := &Client{}
		client.httpClient = &mockHTTPClientError{}
		response, err := client.RestGetTrades(reqID, market)
		assert.Equal(t, Response{}, response)
		assert.Equal(t, errors.New(""), err)
	})
}

func TestRestOpenOrders(t *testing.T) {
	// mock response
	reqID := 1
	market := "ETH_CRO"
	pageNumber := 1
	pageSize := 200

	t.Run("Success", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`
    {
      "id": %d,
      "method": "private/get-open-orders",
      "code": 0,
      "result": {
        "count": 28,
        "order_list": [
          {
            "avg_price": 0,
            "client_oid": "8146de38-4514-414c-9f41-db2f339f7202",
            "create_time": 1611880079268,
            "cumulative_quantity": 0,
            "cumulative_value": 0,
            "exec_inst": "",
            "fee_currency": "ETH",
            "instrument_name": "ETH_CRO",
            "order_id": "1142302899251991010",
            "price": 0.4,
            "quantity": 0.3,
            "side": "BUY",
            "status": "ACTIVE",
            "time_in_force": "GOOD_TILL_CANCEL",
            "type": "LIMIT",
            "update_time": 1611880079298
          }
        ]
      }
    }`, reqID)
		expectedEndpoint := `/v2/private/get-open-orders`
		testRest(t,
			expectedEndpoint,
			jsonStr,
			func(client *Client) (Response, error) {
				return client.RestOpenOrders(reqID, market, pageNumber, pageSize)
			},
		)
	})

	t.Run("HTTP client error", func(t *testing.T) {
		client := &Client{}
		client.httpClient = &mockHTTPClientError{}
		response, err := client.RestOpenOrders(reqID, market, pageNumber, pageSize)
		assert.Equal(t, Response{}, response)
		assert.Equal(t, errors.New(""), err)
	})
}

func mockResponseBody(id int, method string, code int, result interface{}) []byte {
	b, _ := json.Marshal(map[string]interface {
	}{
		"id":     id,
		"method": method,
		"code":   code,
		"result": result,
	})
	return b
}
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
		mockClient := &httpClientMock{}
		mockResponse := &http.Response{
			StatusCode: r.responseCode,
			Body:       ioutil.NopCloser(bytes.NewReader(r.responseBody)),
		}
		mockClient.On("Do", mock.Anything).Once().Return(mockResponse, nil)
		cli := &Client{rest: newHttpClient(mockClient,
			fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion),
		)}
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
