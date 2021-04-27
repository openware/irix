package bitmex

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/openware/pkg/common"
	"github.com/openware/tradepoint/config"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
	"github.com/openware/tradepoint/portfolio/withdraw"
	exchange "github.com/openware/irix"
	"github.com/openware/irix/account"
	"github.com/openware/pkg/asset"
	"github.com/openware/irix/kline"
	"github.com/openware/irix/order"
	"github.com/openware/irix/orderbook"
	"github.com/openware/irix/protocol"
	"github.com/openware/pkg/request"
	"github.com/openware/irix/stream"
	"github.com/openware/irix/ticker"
	"github.com/openware/irix/trade"
)

// GetDefaultConfig returns a default exchange config
func (b *Bitmex) GetDefaultConfig() (*config.ExchangeConfig, error) {
	b.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = b.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = b.BaseCurrencies

	err := b.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if b.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = b.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets the basic defaults for Bitmex
func (b *Bitmex) SetDefaults() {
	b.Name = "Bitmex"
	b.Enabled = true
	b.Verbose = true
	b.API.CredentialsValidator.RequiresKey = true
	b.API.CredentialsValidator.RequiresSecret = true

	requestFmt := &currency.PairFormat{Uppercase: true}
	configFmt := &currency.PairFormat{Uppercase: true}
	err := b.SetGlobalPairsManager(requestFmt,
		configFmt,
		asset.PerpetualContract,
		asset.Futures,
		asset.Index)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	err = b.DisableAssetWebsocketSupport(asset.Index)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	b.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerBatching:      true,
				TickerFetching:      true,
				TradeFetching:       true,
				OrderbookFetching:   true,
				AutoPairUpdates:     true,
				AccountInfo:         true,
				GetOrder:            true,
				GetOrders:           true,
				CancelOrders:        true,
				CancelOrder:         true,
				SubmitOrder:         true,
				SubmitOrders:        true,
				ModifyOrder:         true,
				DepositHistory:      true,
				WithdrawalHistory:   true,
				UserTradeHistory:    true,
				CryptoDeposit:       true,
				CryptoWithdrawal:    true,
				TradeFee:            true,
				CryptoWithdrawalFee: true,
			},
			WebsocketCapabilities: protocol.Features{
				TradeFetching:          true,
				OrderbookFetching:      true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				AccountInfo:            true,
				DeadMansSwitch:         true,
				GetOrders:              true,
				GetOrder:               true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCryptoWithAPIPermission |
				exchange.WithdrawCryptoWithEmail |
				exchange.WithdrawCryptoWith2FA |
				exchange.NoFiatWithdrawals,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	b.Requester = request.New(b.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	b.API.Endpoints = b.NewEndpoints()
	err = b.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      bitmexAPIURL,
		exchange.WebsocketSpot: bitmexWSURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	b.Websocket = stream.New()
	b.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	b.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	b.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (b *Bitmex) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	err := b.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsEndpoint, err := b.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = b.Websocket.Setup(&stream.WebsocketSetup{
		Enabled:                          exch.Features.Enabled.Websocket,
		Verbose:                          exch.Verbose,
		AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
		WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
		DefaultURL:                       bitmexWSURL,
		ExchangeName:                     exch.Name,
		RunningURL:                       wsEndpoint,
		Connector:                        b.WsConnect,
		Subscriber:                       b.Subscribe,
		UnSubscriber:                     b.Unsubscribe,
		GenerateSubscriptions:            b.GenerateDefaultSubscriptions,
		Features:                         &b.Features.Supports.WebsocketCapabilities,
		OrderbookBufferLimit:             exch.OrderbookConfig.WebsocketBufferLimit,
		BufferEnabled:                    exch.OrderbookConfig.WebsocketBufferEnabled,
		UpdateEntriesByID:                true,
	})
	if err != nil {
		return err
	}
	return b.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the Bitmex go routine
func (b *Bitmex) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the Bitmex wrapper
func (b *Bitmex) Run() {
	if b.Verbose {
		wsEndpoint, err := b.API.Endpoints.GetURL(exchange.WebsocketSpot)
		if err != nil {
			log.Error(log.ExchangeSys, err)
		}
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s. (url: %s).\n",
			b.Name,
			common.IsEnabled(b.Websocket.IsEnabled()),
			wsEndpoint)
		b.PrintEnabledPairs()
	}

	if !b.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := b.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", b.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *Bitmex) FetchTradablePairs(asset asset.Item) ([]string, error) {
	marketInfo, err := b.GetActiveAndIndexInstruments()
	if err != nil {
		return nil, err
	}

	var products []string
	for x := range marketInfo {
		products = append(products, marketInfo[x].Symbol.String())
	}

	return products, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *Bitmex) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := b.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	// Zerovalue current list which will remove old asset pairs when contract
	// types expire or become obsolete
	var assetPairs = map[asset.Item][]string{
		asset.Index:             {},
		asset.PerpetualContract: {},
		asset.Futures:           {},
	}

	for x := range pairs {
		if strings.Contains(pairs[x], ".") {
			assetPairs[asset.Index] = append(assetPairs[asset.Index], pairs[x])
			continue
		}

		if strings.Contains(pairs[x], "USD") {
			assetPairs[asset.PerpetualContract] = append(assetPairs[asset.PerpetualContract],
				pairs[x])
			continue
		}

		assetPairs[asset.Futures] = append(assetPairs[asset.Futures], pairs[x])
	}

	for a, values := range assetPairs {
		p, err := currency.NewPairsFromStrings(values)
		if err != nil {
			return err
		}

		err = b.UpdatePairs(p, a, false, false)
		if err != nil {
			log.Warnf(log.ExchangeSys,
				"%s failed to update available pairs. Err: %v",
				b.Name,
				err)
		}
	}

	return nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bitmex) UpdateTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	tick, err := b.GetActiveAndIndexInstruments()
	if err != nil {
		return nil, err
	}

	pairs, err := b.GetEnabledPairs(assetType)
	if err != nil {
		return nil, err
	}

	for j := range tick {
		if !pairs.Contains(tick[j].Symbol, true) {
			continue
		}

		err = ticker.ProcessTicker(&ticker.Price{
			Last:         tick[j].LastPrice,
			High:         tick[j].HighPrice,
			Low:          tick[j].LowPrice,
			Bid:          tick[j].BidPrice,
			Ask:          tick[j].AskPrice,
			Volume:       tick[j].Volume24h,
			Close:        tick[j].PrevClosePrice,
			Pair:         tick[j].Symbol,
			LastUpdated:  tick[j].Timestamp,
			ExchangeName: b.Name,
			AssetType:    assetType})
		if err != nil {
			return nil, err
		}
	}
	return ticker.GetTicker(b.Name, fPair, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *Bitmex) FetchTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	tickerNew, err := ticker.GetTicker(b.Name, fPair, assetType)
	if err != nil {
		return b.UpdateTicker(fPair, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (b *Bitmex) FetchOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	ob, err := orderbook.Get(b.Name, fPair, assetType)
	if err != nil {
		return b.UpdateOrderbook(fPair, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bitmex) UpdateOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        b.Name,
		Pair:            p,
		Asset:           assetType,
		VerifyOrderbook: b.CanVerifyOrderbook,
	}

	if assetType == asset.Index {
		return book, common.ErrFunctionNotSupported
	}

	fpair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return book, err
	}

	orderbookNew, err := b.GetOrderbook(OrderBookGetL2Params{
		Symbol: fpair.String(),
		Depth:  500})
	if err != nil {
		return book, err
	}

	for i := range orderbookNew {
		switch {
		case strings.EqualFold(orderbookNew[i].Side, order.Sell.String()):
			book.Asks = append(book.Asks, orderbook.Item{
				Amount: float64(orderbookNew[i].Size),
				Price:  orderbookNew[i].Price})
		case strings.EqualFold(orderbookNew[i].Side, order.Buy.String()):
			book.Bids = append(book.Bids, orderbook.Item{
				Amount: float64(orderbookNew[i].Size),
				Price:  orderbookNew[i].Price})
		default:
			return book,
				fmt.Errorf("could not process orderbook, order side [%s] could not be matched",
					orderbookNew[i].Side)
		}
	}
	book.Asks.Reverse() // Reverse order of asks to ascending

	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(b.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies for the
// Bitmex exchange
func (b *Bitmex) UpdateAccountInfo(assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings

	bal, err := b.GetAllUserMargin()
	if err != nil {
		return info, err
	}

	// Need to update to add Margin/Liquidity availibilty
	var balances []account.Balance
	for i := range bal {
		balances = append(balances, account.Balance{
			CurrencyName: currency.NewCode(bal[i].Currency),
			TotalValue:   float64(bal[i].WalletBalance),
		})
	}

	info.Exchange = b.Name
	info.Accounts = append(info.Accounts, account.SubAccount{
		Currencies: balances,
	})

	err = account.Process(&info)
	if err != nil {
		return account.Holdings{}, err
	}

	return info, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (b *Bitmex) FetchAccountInfo(assetType asset.Item) (account.Holdings, error) {
	acc, err := account.GetHoldings(b.Name, assetType)
	if err != nil {
		return b.UpdateAccountInfo(assetType)
	}

	return acc, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bitmex) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// GetWithdrawalsHistory returns previous withdrawals data
func (b *Bitmex) GetWithdrawalsHistory(c currency.Code) (resp []exchange.WithdrawalHistory, err error) {
	return nil, common.ErrNotYetImplemented
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (b *Bitmex) GetRecentTrades(p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return b.GetHistoricTrades(p, assetType, time.Now().Add(-time.Hour), time.Now())
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (b *Bitmex) GetHistoricTrades(p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	if assetType == asset.Index {
		return nil, fmt.Errorf("asset type '%v' not supported", assetType)
	}
	if timestampEnd.After(time.Now()) || timestampEnd.Before(timestampStart) {
		return nil, fmt.Errorf("invalid time range supplied. Start: %v End %v", timestampStart, timestampEnd)
	}
	var err error
	p, err = b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	limit := 1000
	req := &GenericRequestParams{
		Symbol:  p.String(),
		Count:   int32(limit),
		EndTime: timestampEnd.UTC().Format("2006-01-02T15:04:05.000Z"),
	}
	ts := timestampStart
	var resp []trade.Data
allTrades:
	for {
		req.StartTime = ts.UTC().Format("2006-01-02T15:04:05.000Z")
		var tradeData []Trade
		tradeData, err = b.GetTrade(req)
		if err != nil {
			return nil, err
		}
		for i := range tradeData {
			if tradeData[i].Timestamp.Before(timestampStart) || tradeData[i].Timestamp.After(timestampEnd) {
				break allTrades
			}
			var side order.Side
			side, err = order.StringToOrderSide(tradeData[i].Side)
			if err != nil {
				return nil, err
			}
			if tradeData[i].Price == 0 {
				// Please note that indices (symbols starting with .) post trades at intervals to the trade feed.
				// These have a size of 0 and are used only to indicate a changing price.
				continue
			}
			resp = append(resp, trade.Data{
				Exchange:     b.Name,
				CurrencyPair: p,
				AssetType:    assetType,
				Side:         side,
				Price:        tradeData[i].Price,
				Amount:       float64(tradeData[i].Size),
				Timestamp:    tradeData[i].Timestamp,
				TID:          tradeData[i].TrdMatchID,
			})
			if i == len(tradeData)-1 {
				if ts.Equal(tradeData[i].Timestamp) {
					// reached end of trades to crawl
					break allTrades
				}
				ts = tradeData[i].Timestamp
			}
		}
		if len(tradeData) != limit {
			break allTrades
		}
	}
	err = b.AddTradesToBuffer(resp...)
	if err != nil {
		return nil, err
	}

	sort.Sort(trade.ByDate(resp))
	return trade.FilterTradesByTime(resp, timestampStart, timestampEnd), nil
}

// SubmitOrder submits a new order
func (b *Bitmex) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}

	if math.Mod(s.Amount, 1) != 0 {
		return submitOrderResponse,
			errors.New("order contract amount can not have decimals")
	}

	fPair, err := b.FormatExchangeCurrency(s.Pair, s.AssetType)
	if err != nil {
		return submitOrderResponse, err
	}

	var orderNewParams = OrderNewParams{
		OrderType:     s.Type.Title(),
		Symbol:        fPair.String(),
		OrderQuantity: s.Amount,
		Side:          s.Side.Title(),
	}

	if s.Type == order.Limit {
		orderNewParams.Price = s.Price
	}

	response, err := b.CreateOrder(&orderNewParams)
	if err != nil {
		return submitOrderResponse, err
	}
	if response.OrderID != "" {
		submitOrderResponse.OrderID = response.OrderID
	}
	if s.Type == order.Market {
		submitOrderResponse.FullyMatched = true
	}
	submitOrderResponse.IsOrderPlaced = true

	return submitOrderResponse, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bitmex) ModifyOrder(action *order.Modify) (string, error) {
	if err := action.Validate(); err != nil {
		return "", err
	}

	var params OrderAmendParams

	if math.Mod(action.Amount, 1) != 0 {
		return "", errors.New("contract amount can not have decimals")
	}

	params.OrderID = action.ID
	params.OrderQty = int32(action.Amount)
	params.Price = action.Price

	order, err := b.AmendOrder(&params)
	if err != nil {
		return "", err
	}

	return order.OrderID, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bitmex) CancelOrder(o *order.Cancel) error {
	if err := o.Validate(o.StandardCancel()); err != nil {
		return err
	}
	var params = OrderCancelParams{
		OrderID: o.ID,
	}
	_, err := b.CancelOrders(&params)
	return err
}

// CancelBatchOrders cancels an orders by their corresponding ID numbers
func (b *Bitmex) CancelBatchOrders(o []order.Cancel) (order.CancelBatchResponse, error) {
	return order.CancelBatchResponse{}, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bitmex) CancelAllOrders(_ *order.Cancel) (order.CancelAllResponse, error) {
	cancelAllOrdersResponse := order.CancelAllResponse{
		Status: make(map[string]string),
	}
	var emptyParams OrderCancelAllParams
	orders, err := b.CancelAllExistingOrders(emptyParams)
	if err != nil {
		return cancelAllOrdersResponse, err
	}

	for i := range orders {
		if orders[i].OrdRejReason != "" {
			cancelAllOrdersResponse.Status[orders[i].OrderID] = orders[i].OrdRejReason
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns order information based on order ID
func (b *Bitmex) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	var orderDetail order.Detail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *Bitmex) GetDepositAddress(cryptocurrency currency.Code, _ string) (string, error) {
	return b.GetCryptoDepositAddress(cryptocurrency.String())
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}

	var request = UserRequestWithdrawalParams{
		Address:  withdrawRequest.Crypto.Address,
		Amount:   withdrawRequest.Amount,
		Currency: withdrawRequest.Currency.String(),
		OtpToken: withdrawRequest.OneTimePassword,
	}
	if withdrawRequest.Crypto.FeeAmount > 0 {
		request.Fee = withdrawRequest.Crypto.FeeAmount
	}

	resp, err := b.UserRequestWithdrawal(request)
	if err != nil {
		return nil, err
	}

	return &withdraw.ExchangeResponse{
		Status: resp.Text,
		ID:     resp.Tx,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bitmex) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !b.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
// This function is not concurrency safe due to orderSide/orderType maps
func (b *Bitmex) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var orders []order.Detail
	params := OrdersRequest{}
	params.Filter = "{\"open\":true}"

	resp, err := b.GetOrders(&params)
	if err != nil {
		return nil, err
	}

	format, err := b.GetPairFormat(asset.PerpetualContract, false)
	if err != nil {
		return nil, err
	}

	for i := range resp {
		orderSide := orderSideMap[resp[i].Side]
		orderType := orderTypeMap[resp[i].OrdType]
		if orderType == "" {
			orderType = order.UnknownType
		}

		orderDetail := order.Detail{
			Date:     resp[i].Timestamp,
			Price:    resp[i].Price,
			Amount:   float64(resp[i].OrderQty),
			Exchange: b.Name,
			ID:       resp[i].OrderID,
			Side:     orderSide,
			Type:     orderType,
			Status:   order.Status(resp[i].OrdStatus),
			Pair: currency.NewPairWithDelimiter(resp[i].Symbol,
				resp[i].SettlCurrency,
				format.Delimiter),
		}

		orders = append(orders, orderDetail)
	}

	order.FilterOrdersBySide(&orders, req.Side)
	order.FilterOrdersByType(&orders, req.Type)
	order.FilterOrdersByTimeRange(&orders, req.StartTime, req.EndTime)
	order.FilterOrdersByCurrencies(&orders, req.Pairs)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
// This function is not concurrency safe due to orderSide/orderType maps
func (b *Bitmex) GetOrderHistory(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var orders []order.Detail
	params := OrdersRequest{}
	resp, err := b.GetOrders(&params)
	if err != nil {
		return nil, err
	}

	format, err := b.GetPairFormat(asset.PerpetualContract, false)
	if err != nil {
		return nil, err
	}

	for i := range resp {
		orderSide := orderSideMap[resp[i].Side]
		orderType := orderTypeMap[resp[i].OrdType]
		if orderType == "" {
			orderType = order.UnknownType
		}

		orderDetail := order.Detail{
			Price:    resp[i].Price,
			Amount:   float64(resp[i].OrderQty),
			Exchange: b.Name,
			ID:       resp[i].OrderID,
			Side:     orderSide,
			Type:     orderType,
			Status:   order.Status(resp[i].OrdStatus),
			Pair: currency.NewPairWithDelimiter(resp[i].Symbol,
				resp[i].SettlCurrency,
				format.Delimiter),
		}

		orders = append(orders, orderDetail)
	}

	order.FilterOrdersBySide(&orders, req.Side)
	order.FilterOrdersByType(&orders, req.Type)
	order.FilterOrdersByTimeRange(&orders, req.StartTime, req.EndTime)
	order.FilterOrdersByCurrencies(&orders, req.Pairs)
	return orders, nil
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (b *Bitmex) AuthenticateWebsocket() error {
	return b.websocketSendAuth()
}

// ValidateCredentials validates current credentials used for wrapper
// functionality
func (b *Bitmex) ValidateCredentials(assetType asset.Item) error {
	_, err := b.UpdateAccountInfo(assetType)
	return b.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (b *Bitmex) GetHistoricCandles(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return kline.Item{}, common.ErrFunctionNotSupported
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (b *Bitmex) GetHistoricCandlesExtended(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return kline.Item{}, common.ErrFunctionNotSupported
}
