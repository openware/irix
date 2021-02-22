package cryptocom

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	id := 1
	method := "TestMethod"
	apiKey := "TestApiKey"
	signature := "TestSignature"
	params := map[string]interface{}{
		"key1": "test1",
		"key2": "test2",
	}
	jsonParams, _ := json.Marshal(params)
	nonce := generateNonce()

	t.Run("AuthRequest", func(t *testing.T) {
		expected := fmt.Sprintf(
			`{"api_key":"%s","id":%d,"method":"%s","nonce":"%s","sig":"%s"}`,
			apiKey,
			id,
			method,
			nonce,
			signature,
		)
		request := &Request{
			Id:        id,
			Type:      AuthRequest,
			Method:    method,
			ApiKey:    apiKey,
			Signature: signature,
			Nonce:     generateNonce(),
		}

		b, _ := request.Encode()
		assert.Equal(t, expected, string(b))
	})

	t.Run("RestOrderRequest, RestBalanceRequest, RestTradesRequest, RestOpenOrdersRequest", func(t *testing.T) {
		expected := fmt.Sprintf(
			`{"api_key":"%s","id":%d,"method":"%s","nonce":"%s","params":%s,"sig":"%s"}`,
			apiKey,
			id,
			method,
			nonce,
			string(jsonParams),
			signature,
		)

		types := []uint8{RestOrderRequest, RestBalanceRequest, RestTradesRequest, RestOpenOrdersRequest}

		for _, ty := range types {
			request := &Request{
				Id:        id,
				Type:      ty,
				Method:    method,
				ApiKey:    apiKey,
				Signature: signature,
				Nonce:     nonce,
				Params:    params,
			}

			b, _ := request.Encode()

			assert.Equal(t, expected, string(b))
		}
	})

	t.Run("SubscribeRequest, OrderRequest", func(t *testing.T) {
		expected := fmt.Sprintf(
			`{"id":%d,"method":"%s","nonce":"%s","params":%s}`,
			id,
			method,
			nonce,
			string(jsonParams),
		)

		types := []uint8{OrderRequest, SubscribeRequest}

		for _, ty := range types {
			request := &Request{
				Id:     id,
				Type:   ty,
				Method: method,
				Nonce:  nonce,
				Params: params,
			}

			b, _ := request.Encode()

			assert.Equal(t, expected, string(b))
		}
	})

	t.Run("HeartBeat", func(t *testing.T) {
		expected := fmt.Sprintf(
			`{"id":%d,"method":"%s"}`,
			id,
			method,
		)
		request := &Request{
			Id:     id,
			Type:   HeartBeat,
			Method: method,
		}

		b, _ := request.Encode()
		assert.Equal(t, expected, string(b))
	})

	t.Run("Invalid type", func(t *testing.T) {
		request := &Request{
			Id:     id,
			Type:   99,
			Method: method,
		}

		b, err := request.Encode()

		assert.Equal(t, []byte(nil), b)
		assert.Equal(t, err, errors.New("invalid type"))
	})
}
