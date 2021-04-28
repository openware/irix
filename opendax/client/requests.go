package client

import (
	"github.com/google/uuid"
	"github.com/openware/openfinex/pkg/model/order"
	"github.com/openware/openfinex/pkg/msg"
	"github.com/openware/openfinex/pkg/websocket/protocol"
)

func SubscribePrivate(reqID uint64, topics ...interface{}) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodSubscribe,
		Args: []interface{}{
			"private",
			topics,
		},
	}
}

func SubscribePublic(reqID uint64, topics ...interface{}) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodSubscribe,
		Args: []interface{}{
			"public",
			topics,
		},
	}
}

func ListOpenOrders(reqID uint64, market string) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodListOrders,
		Args:   []interface{}{market},
	}
}

func GetOrders(reqID uint64, list ...uuid.UUID) *msg.Msg {
	args := make([]interface{}, len(list))
	for i, uuid := range list {
		args[i] = uuid
	}

	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodGetOrders,
		Args:   args,
	}
}

func GetOrderTrades(reqID uint64, uuid uuid.UUID) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodGetOrderTrades,
		Args:   []interface{}{uuid},
	}
}

func orderType(t order.Type) string {
	switch t {
	case order.Limit:
		return protocol.OrderTypeLimit
	case order.Market:
		return protocol.OrderTypeMarket
	case order.PostOnly:
		return protocol.OrderTypePostOnly
	case order.FOK:
		return protocol.OrderTypeFillOrKill
	default:
		return "???"
	}
}

func OrderSide(s order.Side) string {
	switch s {
	case order.Buy:
		return protocol.OrderSideBuy
	case order.Sell:
		return protocol.OrderSideSell
	default:
		return "?"
	}
}

func orderState(s order.State) string {
	switch s {
	case order.StateCancel:
		return protocol.OrderStateCancel
	case order.StateDone:
		return protocol.OrderStateDone
	case order.StateWait:
		return protocol.OrderStateWait
	case order.StatePending:
		return protocol.OrderStatePending
	case order.StateReject:
		return protocol.OrderStateReject
	default:
		return "?"
	}
}

// [1,42,"createOrder",["btcusd", "M", "S", "0.250000", "9120.00"]]
func OrderCreateReq(reqID uint64, o *order.Model) *msg.Msg {
	return msg.NewRequest(
		reqID,
		protocol.MethodCreateOrder,
		o.Market,
		orderType(o.Type),
		OrderSide(o.Side),
		o.Volume.String(),
		o.Price.String(),
	)
}

func OrderCancelReq(reqID uint64, orderUUID string) *msg.Msg {
	return msg.NewRequest(reqID, protocol.MethodCancelOrder, "uuid", orderUUID)
}
