package cryptocom

import "encoding/json"

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
func mockOpenOrders(count int, orders ...OrderInfo) OpenOrdersResult {
	return OpenOrdersResult{
		Count: count,
		OrderList: orders,
	}
}
