package client

import (
	"errors"

	"github.com/google/uuid"
	"github.com/openware/openfinex/pkg/incremental"
	"github.com/openware/openfinex/pkg/model/order"
	tr "github.com/openware/openfinex/pkg/model/trade"
	"github.com/openware/openfinex/pkg/msg"
	"github.com/openware/openfinex/pkg/websocket/protocol"
	"github.com/shopspring/decimal"
)

type OrderMessage struct {
	Market         string
	ID             uint64
	UUID           uuid.UUID
	Side           order.Side
	State          order.State
	Type           order.Type
	Price          decimal.Decimal
	AvgPrice       decimal.Decimal
	Volume         decimal.Decimal
	OriginVolume   decimal.Decimal
	ExecutedVolume decimal.Decimal
	TradesCount    int
	Timestamp      int64
}

type OrderUpdate OrderMessage
type OrderCreate OrderMessage
type OrderReject OrderMessage
type OrderCancel OrderMessage

type UserTrade = tr.UserTrade

type OrderSnapshot struct {
	Market string
	Orders []OrderMessage
}

type OrderInfo struct {
	Orders []OrderMessage
}

type OrderTrades struct {
	Trades []UserTrade
}

type PublicTrade struct {
	Market    string
	ID        uint64
	Price     decimal.Decimal
	Amount    decimal.Decimal
	CreatedAt int64
	TakerSide order.Side
}

type OrderbookIncrement incremental.OrderbookIncrement

type OrderbookSnapshot incremental.OrderbookSnapshot

type BalanceUpdate struct {
	Snapshot []Balance
}

type Balance struct {
	Currency  string
	Available decimal.Decimal
	Locked    decimal.Decimal
}

type Response msg.Msg

func Parse(m *msg.Msg) (interface{}, error) {
	switch m.Type {
	case msg.Response:
		return recognizeResponse(m)

	case msg.EventPrivate:
		return recognizeEventPrivate(m)

	case msg.EventPublic:
		return recognizeEventPublic(m)

	default:
		return m, nil
	}
}

func recognizeResponse(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case protocol.MethodListOrders:
		return orderSnapshot(m.Args)

	case protocol.MethodGetOrders:
		return orderInfo(m.Args)

	case protocol.MethodGetOrderTrades:
		return orderTrades(m.Args)

	default:
		return (*Response)(m), nil
	}
}

func orderTrades(args []interface{}) (*OrderTrades, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderTrades{}

	for {
		trArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		tr, err := trade(trArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Trades = append(snapshot.Trades, *tr)
	}

	return &snapshot, nil
}

func orderInfo(args []interface{}) (*OrderInfo, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderInfo{}

	for {
		ordArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		ord, err := orderMessage(ordArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Orders = append(snapshot.Orders, *ord)
	}

	return &snapshot, nil
}

func orderSnapshot(args []interface{}) (*OrderSnapshot, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderSnapshot{}

	market, err := it.NextString()
	if err != nil {
		return &snapshot, err
	}

	snapshot.Market = market

	for {
		ordArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		ord, err := orderMessage(ordArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Orders = append(snapshot.Orders, *ord)
	}

	return &snapshot, nil
}

func publicTrade(args []interface{}) (interface{}, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	amount, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	tSide, err := it.NextString()
	if err != nil {
		return nil, err
	}

	takerSide, err := recognizeSide(tSide)
	if err != nil {
		return nil, err
	}

	return &PublicTrade{
		Market:    market,
		ID:        id,
		Price:     price,
		Amount:    amount,
		CreatedAt: int64(ts),
		TakerSide: takerSide,
	}, nil
}

func priceLevel(args []interface{}) (incremental.PriceLevel, error) {
	it := msg.NewArgIterator(args)

	price, err := it.NextDecimal()
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	amountString, err := it.NextString()
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	if amountString == "" {
		return incremental.PriceLevel{Price: price, Amount: decimal.Zero}, nil
	}

	amount, err := decimal.NewFromString(amountString)
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	return incremental.PriceLevel{
		Price:  price,
		Amount: amount,
	}, nil
}

// ["btcusd",1,"asks",["9120","0.25"]]
func orderbookIncrement(args []interface{}) (*OrderbookIncrement, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	seq, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	s, err := it.NextString()
	if err != nil {
		return nil, err
	}

	var side order.Side
	switch s {
	case "asks":
		side = order.Sell
	case "bids":
		side = order.Buy
	default:
		return nil, errors.New("invalid increment side " + s)
	}

	slice, err := it.NextSlice()
	if err != nil {
		return nil, err
	}

	pl, err := priceLevel(slice)
	if err != nil {
		return nil, err
	}

	return &OrderbookIncrement{
		Market:     market,
		Sequence:   seq,
		Side:       side,
		PriceLevel: pl,
	}, nil
}

// ["btcusd",1,[["9120","0.25"]],[]]
func orderbookSnapshot(args []interface{}) (*OrderbookSnapshot, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	seq, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	slice, err := it.NextSlice()
	if err != nil {
		return nil, err
	}

	sells, err := getPriceLevels(slice)
	if err != nil {
		return nil, err
	}

	slice, err = it.NextSlice()
	if err != nil {
		return nil, err
	}

	buys, err := getPriceLevels(slice)
	if err != nil {
		return nil, err
	}

	return &OrderbookSnapshot{
		Market:   market,
		Sequence: seq,
		Buys:     buys,
		Sells:    sells,
	}, nil
}

func getPriceLevels(args []interface{}) ([]incremental.PriceLevel, error) {
	it := msg.NewArgIterator(args)
	res := make([]incremental.PriceLevel, 0)

	for {
		slice, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}

		if err != nil {
			return nil, err
		}

		pl, err := priceLevel(slice)
		if err != nil {
			return nil, err
		}

		res = append(res, pl)
	}

	return res, nil
}

func recognizeEventPublic(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case "trade":
		return publicTrade(m.Args)
	case "obi":
		return orderbookIncrement(m.Args)
	case "obs":
		return orderbookSnapshot(m.Args)
	default:
		return m, errors.New("unexpected message")
	}
}

func recognizeEventPrivate(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case protocol.EventOrderCreate:
		parsed, err := orderMessage(m.Args)
		return (*OrderCreate)(parsed), err

	case protocol.EventOrderCancel:
		parsed, err := orderMessage(m.Args)
		return (*OrderCancel)(parsed), err

	case protocol.EventOrderUpdate:
		parsed, err := orderMessage(m.Args)
		return (*OrderUpdate)(parsed), err

	case protocol.EventOrderReject:
		parsed, err := orderMessage(m.Args)
		return (*OrderReject)(parsed), err

	case protocol.EventTrade:
		return trade(m.Args)

	case protocol.EventBalanceUpdate:
		return balanceUpdate(m.Args)

	default:
		return m, errors.New("unexpected message")
	}
}

func balanceUpdate(args []interface{}) (*BalanceUpdate, error) {
	it := msg.NewArgIterator(args)

	balances := make([]Balance, 0, len(args))

loop:
	for {
		bu, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break loop
		}
		if err != nil {
			return nil, err
		}

		buIt := msg.NewArgIterator(bu)
		cur, err := buIt.NextString()
		if err != nil {
			return nil, err
		}

		avail, err := buIt.NextDecimal()
		if err != nil {
			return nil, err
		}

		locked, err := buIt.NextDecimal()
		if err != nil {
			return nil, err
		}

		balances = append(balances, Balance{Currency: cur, Available: avail, Locked: locked})
	}

	return &BalanceUpdate{balances}, nil
}

func trade(args []interface{}) (*UserTrade, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	amount, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	total, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	orderID, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	orderUUID, err := it.NextUUID()
	if err != nil {
		return nil, err
	}

	side, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ordSide, err := recognizeSide(side)
	if err != nil {
		return nil, err
	}

	tSide, err := it.NextString()
	if err != nil {
		return nil, err
	}

	takerSide, err := recognizeSide(tSide)
	if err != nil {
		return nil, err
	}

	fee, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	feeCur, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	return &UserTrade{
		Market:      market,
		ID:          id,
		Price:       price,
		Amount:      amount,
		Total:       total,
		OrderID:     orderID,
		OrderUUID:   orderUUID,
		OrderSide:   ordSide,
		TakerSide:   takerSide,
		Fee:         fee,
		FeeCurrency: feeCur,
		CreatedAt:   int64(ts),
	}, nil
}

func MessageFromUserTrade(tr *UserTrade) []interface{} {
	return []interface{}{
		tr.Market,
		tr.ID,
		tr.Price,
		tr.Amount,
		tr.Total,
		tr.OrderID,
		tr.OrderUUID,
		OrderSide(tr.OrderSide),
		OrderSide(tr.TakerSide),
		tr.Fee,
		tr.FeeCurrency,
		tr.CreatedAt,
	}
}

func MessageFromOrder(o *order.Model) []interface{} {
	return []interface{}{
		o.Market,
		o.ID,
		o.UUID,
		OrderSide(o.Side),
		orderState(o.State),
		orderType(o.Type),
		o.Price,
		o.AveragePrice(),
		o.Volume,
		o.OriginVolume,
		o.OriginVolume.Sub(o.Volume),
		o.TradesCount,
		o.CreatedAt,
	}
}

func recognizeSide(side string) (order.Side, error) {
	switch side {
	case protocol.OrderSideSell:
		return order.Sell, nil
	case protocol.OrderSideBuy:
		return order.Buy, nil
	default:
		return "?", errors.New("order side invalid: " + side)
	}
}

func orderMessage(args []interface{}) (*OrderMessage, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	uuidParsed, err := it.NextUUID()
	if err != nil {
		return nil, err
	}

	side, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ordSide, err := recognizeSide(side)
	if err != nil {
		return nil, err
	}

	var ordState order.State
	state, err := it.NextString()
	if err != nil {
		return nil, err
	}

	switch state {
	case protocol.OrderStatePending:
		ordState = order.StatePending
	case protocol.OrderStateWait:
		ordState = order.StateWait
	case protocol.OrderStateDone:
		ordState = order.StateDone
	case protocol.OrderStateReject:
		ordState = order.StateReject
	case protocol.OrderStateCancel:
		ordState = order.StateCancel
	default:
		return nil, errors.New("unexpected order state ")
	}

	var ordType order.Type
	typ, err := it.NextString()
	if err != nil {
		return nil, err
	}

	switch typ {
	case protocol.OrderTypeMarket:
		ordType = order.Market
	case protocol.OrderTypeLimit:
		ordType = order.Limit
	case protocol.OrderTypePostOnly:
		ordType = order.PostOnly
	default:
		return nil, errors.New("unexpected order type")
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	avgPrice, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	volume, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	originVolume, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	executed, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	trades, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	return &OrderMessage{
		Market:         market,
		ID:             id,
		UUID:           uuidParsed,
		Side:           ordSide,
		State:          ordState,
		Type:           ordType,
		Price:          price,
		AvgPrice:       avgPrice,
		Volume:         volume,
		OriginVolume:   originVolume,
		ExecutedVolume: executed,
		TradesCount:    int(trades),
		Timestamp:      int64(ts),
	}, nil
}
