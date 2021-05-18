package cryptocom

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

type testingFunc func(client *Client)
type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) ReadMessage() (int, []byte, error) {
	args := m.Called()
	return args.Get(0).(int), args.Get(1).([]byte), args.Error(2)
}

func (m *mockTransport) WriteMessage(i int, i2 []byte) error {
	return m.Called(i, i2).Error(0)
}

func (m *mockTransport) Close() error {
	return m.Called().Error(0)
}

func mockWsClient() (cli *Client, public *mockTransport, private *mockTransport) {
	cli = New("test", "test", "test", "test")
	public = &mockTransport{}
	private = &mockTransport{}
	cli.publicConn = Connection{
		IsPrivate: false,
		Transport: public,
		Mutex:     &sync.Mutex{},
	}
	cli.privateConn = Connection{
		IsPrivate: true,
		Transport: private,
		Mutex:     &sync.Mutex{},
	}
	return
}

func TestFormat(t *testing.T) {
	markets := []string{"ETH_BTC", "ETH_COV", "XRP_BTC"}
	expected := []string{"trade.ETH_BTC", "trade.ETH_COV", "trade.XRP_BTC"}

	result := format(markets, func(s string) string {
		return fmt.Sprintf("trade.%s", s)
	})

	assert.Equal(t, result, expected)
}

func testSubscribe(t *testing.T, expected string, isPrivate bool, testFunc testingFunc) {
	// prepare expected
	var expectedResponse Request
	err := json.Unmarshal([]byte(expected), &expectedResponse)
	if err != nil {
		t.Fatal("error on parse expected", err)
	}

	// prepare mock
	client := New("test", "test", "test", "test")
	privateWritingMessage := bytes.NewBuffer(nil)
	publicWritingMessage := bytes.NewBuffer(nil)

	privateResponse := make(chan *bytes.Buffer)
	publicResponse := make(chan *bytes.Buffer)
	client.connectMock(privateResponse, publicResponse, privateWritingMessage, publicWritingMessage)

	// call test function
	testFunc(client)

	// get response
	var writingMessage Request
	if isPrivate {
		err = json.Unmarshal(privateWritingMessage.Bytes(), &writingMessage)
	} else {
		err = json.Unmarshal(publicWritingMessage.Bytes(), &writingMessage)
	}
	if err != nil {
		t.Fatal("error on parse writing message")
	}

	// assertion
	assert.NotEqual(t, Request{}, writingMessage)
	// doesn't assert on nonce
	assert.Equal(t, expectedResponse.Method, writingMessage.Method)
	assert.Equal(t, expectedResponse.Params, writingMessage.Params)
}

// for test parse json. this case expected and mock is the same thing
func testResponse(t *testing.T, expected string, isPrivate bool) {
	// prepare expected
	var expectedResponse Response
	err := json.Unmarshal([]byte(expected), &expectedResponse)
	if err != nil {
		t.Fatal("error on parse expected")
	}

	// prepare mock
	client := New("test", "test", "test", "test")
	privateResponse := make(chan *bytes.Buffer, 1)
	publicResponse := make(chan *bytes.Buffer, 1)
	client.connectMock(privateResponse, publicResponse, bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	if isPrivate {
		privateResponse <- bytes.NewBufferString(expected)
	} else {
		publicResponse <- bytes.NewBufferString(expected)
	}

	msgs := client.Listen()
	resp := <-msgs

	// assertion
	assert.NotEqual(t, Response{}, resp)
	assert.Equal(t, expectedResponse, resp)
}


func TestCreateLimitOrder(t *testing.T) {
	t.Run("Subscribe BUY", func(t *testing.T) {
		// prepare expected
		u := uuid.New()
		price := decimal.NewFromFloat(0.01)
		volume := decimal.NewFromFloat(0.0001)

		expected := fmt.Sprintf(
			`{"id":1,"method":"private/create-order","nonce":0,"params":{"client_oid":"%s","instrument_name":"ETH_CRO","price":"%s","quantity":"%s","side":"%s","type":"LIMIT"}}`,
			u, price.String(), volume.String(), "BUY",
		)
		testSubscribe(t, expected, true, func(client *Client) {
			client.CreateLimitOrder(
				1,
				"ETH",
				"CRO",
				"buy",
				price,
				volume,
				u,
			)
		})
	})

	t.Run("Subscribe Sell", func(t *testing.T) {
		// prepare expected
		u := uuid.New()
		price := decimal.NewFromFloat(0.01)
		volume := decimal.NewFromFloat(0.0001)

		expected := fmt.Sprintf(
			`{"id":1,"method":"private/create-order","nonce":0,"params":{"client_oid":"%s","instrument_name":"ETH_CRO","price":"%s","quantity":"%s","side":"%s","type":"LIMIT"}}`,
			u, price.String(), volume.String(), "SELL",
		)
		testSubscribe(t, expected, true, func(client *Client) {
			client.CreateLimitOrder(
				1,
				"ETH",
				"CRO",
				"sell",
				price,
				volume,
				u,
			)
		})
	})

	t.Run("Read response", func(t *testing.T) {
		jsonExpected := `{
			"id": 11,
			"method": "private/create-order",
			"result": {
				"order_id": "337843775021233500",
				"client_oid": "my_order_0002"
			}
		}`
		testResponse(t, jsonExpected, true)
	})
}

func TestCreateMarketOrder(t *testing.T) {
	t.Run("Subscribe BUY", func(t *testing.T) {
		// prepare expected
		uuid := uuid.New()
		volume := decimal.NewFromFloat(0.0001)

		expected := fmt.Sprintf(
			`{"id":1,"method":"private/create-order","nonce":0,"params":{"client_oid":"%s","instrument_name":"ETH_CRO","notional":"%s","side":"%s","type":"MARKET"}}`,
			uuid, volume.String(), "BUY",
		)
		testSubscribe(t, expected, true, func(client *Client) {
			client.CreateMarketOrder(
				1,
				"ETH",
				"CRO",
				"buy",
				volume,
				uuid,
			)
		})
	})

	t.Run("Subscribe Sell", func(t *testing.T) {
		// prepare expected
		uuid := uuid.New()
		volume := decimal.NewFromFloat(0.0001)

		expected := fmt.Sprintf(
			`{"id":1,"method":"private/create-order","nonce":0,"params":{"client_oid":"%s","instrument_name":"ETH_CRO","quantity":"%s","side":"%s","type":"MARKET"}}`,
			uuid, volume.String(), "SELL",
		)
		testSubscribe(t, expected, true, func(client *Client) {
			client.CreateMarketOrder(
				1,
				"ETH",
				"CRO",
				"sell",
				volume,
				uuid,
			)
		})
	})

	t.Run("Read response", func(t *testing.T) {
		jsonExpected := `{
      "id": 11,
      "method": "private/create-order",
      "result": {
        "order_id": "337843775021233500",
        "client_oid": "my_order_0002"
      }
    }`
		testResponse(t, jsonExpected, true)
	})
}

func TestCancelOrder(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		remoteID := sql.NullString{String: "1138210129647637539", Valid: true}

		// prepare expected
		expected := fmt.Sprintf(
			`{"id":1,"method":"private/cancel-order","nonce":0,"params":{"instrument_name":"ETH_CRO","order_id":"%s"}}`,
			remoteID.String,
		)
		testSubscribe(t, expected, true, func(client *Client) {
			client.CancelOrder(
				1,
				remoteID.String,
				"ETH_CRO",
			)
		})
	})

	t.Run("Read response", func(t *testing.T) {
		jsonExpected := `{
      "id": 11,
      "method": "private/cancel-order",
      "code":0
    }`
		testResponse(t, jsonExpected, true)
	})
}

func TestCancelAllOrders(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		// prepare expected
		expected := `{"id":1,"method":"private/cancel-all-orders","nonce":0,"params":{"instrument_name":"ETH_CRO"}}`
		testSubscribe(t, expected, true, func(client *Client) { client.CancelAllOrders(1, "ETH_CRO") })
	})

	t.Run("Read response", func(t *testing.T) {
		jsonExpected := `{
      "id": 12,
      "method": "private/cancel-all-order",
      "code": 0
    }`
		testResponse(t, jsonExpected, true)
	})
}

func TestGetOrderDetails(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		// prepare expected
		remoteID := "1138210129647637539"
		expected := fmt.Sprintf(`{"id":1,"method":"private/get-order-detail","nonce":%d,"params":{"order_id":"1138210129647637539"}}`, generateNonce())
		testSubscribe(t, expected, true, func(client *Client) { client.GetOrderDetails(1, remoteID) })
	})

	t.Run("Read response", func(t *testing.T) {
		jsonExpected := `{
      "id": 11,
      "method": "private/get-order-detail",
      "code": 0,
      "result": {
        "trade_list": [
          {
            "side": "BUY",
            "instrument_name": "ETH_CRO",
            "fee": 0.007,
            "trade_id": "371303044218155296",
            "create_time": 1588902493045,
            "traded_price": 7,
            "traded_quantity": 7,
            "fee_currency": "CRO",
            "order_id": "371302913889488619"
          }
        ],
        "order_info": {
          "status": "FILLED",
          "side": "BUY",
          "order_id": "371302913889488619",
          "client_oid": "9_yMYJDNEeqHxLqtD_2j3g",
          "create_time": 1588902489144,
          "update_time": 1588902493024,
          "type": "LIMIT",
          "instrument_name": "ETH_CRO",
          "cumulative_quantity": 7,
          "cumulative_value": 7,
          "avg_price": 7,
          "fee_currency": "CRO",
          "time_in_force": "GOOD_TILL_CANCEL",
          "exec_inst": "POST_ONLY"
        }
      }
    }`
		testResponse(t, jsonExpected, true)
	})
}
