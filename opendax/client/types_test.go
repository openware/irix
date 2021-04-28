package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/openware/openfinex/pkg/model/order"
	"github.com/openware/openfinex/pkg/msg"
	"github.com/openware/openfinex/pkg/websocket/protocol"
)

func TestTypes_ParseReponse(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		rawResponse := []byte(`[2,1,"order_cancel",["12d9ff0a-9b7b-11ea-ac47-0242ac170004"]]`)
		m, _ := msg.Parse(rawResponse)

		resp, err := Parse(m)

		assert.NoError(t, err)
		assert.IsType(t, &Response{}, resp)
	})
}

func TestTypes_MsgEncodeDecode(t *testing.T) {
	t.Run("order", func(t *testing.T) {
		ord := order.Model{
			ID:           33,
			UID:          "UID",
			UUID:         uuid.MustParse("4dc187c6-3c52-11ea-8b74-5795e76d4efe"),
			User:         79,
			Side:         order.Buy,
			Type:         order.Limit,
			Volume:       decimal.RequireFromString("30.0"),
			OriginVolume: decimal.RequireFromString("130.0"),
			Price:        decimal.RequireFromString("100.0"),
			Locked:       decimal.RequireFromString("100.0"),
			OriginLocked: decimal.RequireFromString("200.0"),
			TradesCount:  77,
			Received:     decimal.RequireFromString("77.7"),
			Market:       "ethusd",
			Ask:          "eth",
			Bid:          "usd",
			MakerFee:     decimal.RequireFromString("0.001"),
			TakerFee:     decimal.RequireFromString("0.001"),
			State:        order.StateWait,
			CreatedAt:    time.Now().Unix(),
		}
		expected := OrderCreate{
			Market:         ord.Market,
			ID:             ord.ID,
			UUID:           ord.UUID,
			Side:           ord.Side,
			State:          ord.State,
			Type:           ord.Type,
			Price:          ord.Price,
			AvgPrice:       ord.AveragePrice(),
			Volume:         ord.Volume,
			OriginVolume:   ord.OriginVolume,
			ExecutedVolume: decimal.RequireFromString("100.0"),
			TradesCount:    int(ord.TradesCount),
			Timestamp:      ord.CreatedAt,
		}
		m := msg.NewEvent(msg.EventPrivate, protocol.EventOrderCreate, MessageFromOrder(&ord))
		enc, err := m.Encode()
		fmt.Println("encoded", string(enc))

		require.NoError(t, err)

		parsed, err := msg.Parse(enc)
		require.NoError(t, err)

		dec, err := Parse(parsed)
		require.NoError(t, err)
		require.IsType(t, &OrderCreate{}, dec)
		assert.Equal(t, fmt.Sprintf("%v", expected), fmt.Sprintf("%v", *dec.(*OrderCreate)))
	})

	t.Run("trade", func(t *testing.T) {
		tr := UserTrade{
			Market:      "btcusd",
			ID:          45,
			Price:       decimal.RequireFromString("123.5"),
			Amount:      decimal.RequireFromString("33.5"),
			OrderID:     333,
			OrderSide:   order.Buy,
			TakerSide:   order.Sell,
			OrderUUID:   uuid.MustParse("4dc187c6-3c52-11ea-8b74-5795e76d4efe"),
			Fee:         decimal.RequireFromString("0.009"),
			FeeCurrency: "usd",
			Total:       decimal.RequireFromString("4445.33"),
			CreatedAt:   time.Now().Unix(),
		}
		expected := tr

		m := msg.NewEvent(msg.EventPrivate, protocol.EventTrade, MessageFromUserTrade(&tr))
		enc, err := m.Encode()
		fmt.Println("encoded", string(enc))
		require.NoError(t, err)

		parsed, err := msg.Parse(enc)
		require.NoError(t, err)

		dec, err := Parse(parsed)
		require.NoError(t, err)
		require.IsType(t, &UserTrade{}, dec)
		assert.Equal(t, expected, *dec.(*UserTrade))
	})
}

func TestTypes_ParseEvent(t *testing.T) {
	t.Run("order create message", func(t *testing.T) {
		rawResponse := []byte(`[4,"on",["btcusd",60,"12d9ff0a-9b7b-11ea-ac47-0242ac170004","sell","w","l","1","0","1","1","0",0,1590076348]]`)
		m, _ := msg.Parse(rawResponse)

		resp, err := Parse(m)

		require.NoError(t, err)
		require.IsType(t, &OrderCreate{}, resp)

		oc := resp.(*OrderCreate)
		assert.Equal(t, "btcusd", oc.Market)
		assert.Equal(t, uuid.MustParse("12d9ff0a-9b7b-11ea-ac47-0242ac170004"), oc.UUID)
		assert.Equal(t, order.Sell, oc.Side)
		assert.Equal(t, uint64(60), oc.ID)
		assert.Equal(t, order.Limit, oc.Type)
		assert.Equal(t, order.StateWait, oc.State)
	})

	t.Run("trade", func(t *testing.T) {
		rawResponse := []byte(`[4,"tr",["btcusd",8,"1","1","1",14,"cf15d277-9f42-11ea-9c56-0242ac190005","sell","buy","0.001", "usd",1590492195]]`)
		m, err := msg.Parse(rawResponse)
		assert.NoError(t, err)

		resp, err := Parse(m)

		assert.NoError(t, err)
		require.IsType(t, &UserTrade{}, resp)

		tr := resp.(*UserTrade)

		assert.Equal(t, int64(1590492195), tr.CreatedAt)
		assert.Equal(t, decimal.RequireFromString("1"), tr.Total)
		assert.Equal(t, uint64(8), tr.ID)
		assert.Equal(t, uint64(14), tr.OrderID)
		assert.Equal(t, "cf15d277-9f42-11ea-9c56-0242ac190005", tr.OrderUUID.String())
	})
}
