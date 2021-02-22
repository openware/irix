package cryptocom

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// Types
	AuthRequest = iota + 1
	SubscribeRequest
	HeartBeat
	OrderRequest
	RestOrderRequest
	RestBalanceRequest
	RestTradesRequest
	RestOpenOrdersRequest
)

type Request struct {
	Id        int
	Type      uint8
	Method    string
	ApiKey    string
	Signature string
	Nonce     string
	Params    map[string]interface{}
}

type Response struct {
	Id      int
	Method  string
	Code    int
	Message string
	Result  map[string]interface{}
}

func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().Unix()*1000)
}

func (r *Request) Encode() ([]byte, error) {
	switch r.Type {
	case AuthRequest:
		return json.Marshal(map[string]interface{}{
			"id":      r.Id,
			"method":  r.Method,
			"api_key": r.ApiKey,
			"sig":     r.Signature,
			"nonce":   r.Nonce,
		})
	case RestOrderRequest, RestBalanceRequest, RestTradesRequest, RestOpenOrdersRequest:
		return json.Marshal(map[string]interface{}{
			"id":      r.Id,
			"method":  r.Method,
			"params":  r.Params,
			"api_key": r.ApiKey,
			"sig":     r.Signature,
			"nonce":   r.Nonce,
		})
	case SubscribeRequest, OrderRequest:
		return json.Marshal(map[string]interface{}{
			"id":     r.Id,
			"method": r.Method,
			"params": r.Params,
			"nonce":  r.Nonce,
		})

	case HeartBeat:
		return json.Marshal(map[string]interface{}{
			"id":     r.Id,
			"method": r.Method,
		})
	default:
		return nil, errors.New("invalid type")
	}
}
