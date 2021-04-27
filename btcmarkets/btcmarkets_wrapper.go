package btcmarkets

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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
func (b *BTCMarkets) GetDefaultConfig() (*config.ExchangeConfig, error) {
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

// SetDefaults sets basic defaults
func (b *BTCMarkets) SetDefaults() {
	b.Name = "BTC Markets"
	b.Enabled = true
	b.Verbose = true
	b.API.CredentialsValidator.RequiresKey = true
	b.API.CredentialsValidator.RequiresSecret = true
	b.API.CredentialsValidator.RequiresBase64DecodeSecret = true
	requestFmt := &currency.PairFormat{Delimiter: currency.DashDelimiter, Uppercase: true}
	configFmt := &currency.PairFormat{Delimiter: currency.DashDelimiter, Uppercase: true}
	err := b.SetGlobalPairsManager(requestFmt, configFmt, asset.Spot)
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
				CancelOrder:         true,
				SubmitOrder:         true,
				UserTradeHistory:    true,
				CryptoWithdrawal:    true,
				FiatWithdraw:        true,
				TradeFee:            true,
				FiatWithdrawalFee:   true,
				CryptoWithdrawalFee: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:         true,
				TradeFetching:          true,
				OrderbookFetching:      true,
				AccountInfo:            true,
				Subscribe:              true,
				AuthenticatedEndpoints: true,
				GetOrders:              true,
				GetOrder:               true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
			Kline: kline.ExchangeCapabilitiesSupported{
				DateRanges: true,
				Intervals:  true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: map[string]bool{
					kline.OneMin.Word():  true,
					kline.OneHour.Word(): true,
					kline.OneDay.Word():  true,
				},
				ResultLimit: 1000,
			},
		},
	}

	b.Requester = request.New(b.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	b.API.Endpoints = b.NewEndpoints()
	err = b.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      btcMarketsAPIURL,
		exchange.WebsocketSpot: btcMarketsWSURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	b.Websocket = stream.New()
	b.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	b.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	b.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in an exchange configuration and sets all parameters
func (b *BTCMarkets) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	err := b.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsURL, err := b.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = b.Websocket.Setup(&stream.WebsocketSetup{
		Enabled:                          exch.Features.Enabled.Websocket,
		Verbose:                          exch.Verbose,
		AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
		WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
		DefaultURL:                       btcMarketsWSURL,
		ExchangeName:                     exch.Name,
		RunningURL:                       wsURL,
		Connector:                        b.WsConnect,
		Subscriber:                       b.Subscribe,
		GenerateSubscriptions:            b.generateDefaultSubscriptions,
		Features:                         &b.Features.Supports.WebsocketCapabilities,
		OrderbookBufferLimit:             exch.OrderbookConfig.WebsocketBufferLimit,
		BufferEnabled:                    exch.OrderbookConfig.WebsocketBufferEnabled,
		SortBuffer:                       true,
	})
	if err != nil {
		return err
	}

	return b.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the BTC Markets go routine
func (b *BTCMarkets) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the BTC Markets wrapper
func (b *BTCMarkets) Run() {
	if b.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s (url: %s).\n",
			b.Name,
			common.IsEnabled(b.Websocket.IsEnabled()),
			btcMarketsWSURL)
		b.PrintEnabledPairs()
	}

	forceUpdate := false
	pairs, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s Failed to update enabled currencies Err:%s\n",
			b.Name,
			err)
		return
	}
	format, err := b.GetPairFormat(asset.Spot, false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s Failed to update enabled currencies.\n",
			b.Name)
		return
	}

	avail, err := b.GetAvailablePairs(asset.Spot)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s Failed to update enabled currencies.\n",
			b.Name)
		return
	}

	if !common.StringDataContains(pairs.Strings(), format.Delimiter) ||
		!common.StringDataContains(avail.Strings(), format.Delimiter) {
		log.Warnln(log.ExchangeSys, "Available pairs for BTC Markets reset due to config upgrade, please enable the pairs you would like again.")
		forceUpdate = true
	}
	if forceUpdate {
		enabledPairs := currency.Pairs{currency.Pair{
			Base:      currency.BTC.Lower(),
			Quote:     currency.AUD.Lower(),
			Delimiter: format.Delimiter,
		},
		}
		err = b.UpdatePairs(enabledPairs, asset.Spot, true, true)
		if err != nil {
			log.Errorf(log.ExchangeSys,
				"%s Failed to update enabled currencies.\n",
				b.Name)
		}
	}

	if !b.GetEnabledFeatures().AutoPairUpdates && !forceUpdate {
		return
	}

	err = b.UpdateTradablePairs(forceUpdate)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			b.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *BTCMarkets) FetchTradablePairs(a asset.Item) ([]string, error) {
	if a != asset.Spot {
		return nil, fmt.Errorf("asset type of %s is not supported by %s", a, b.Name)
	}
	markets, err := b.GetMarkets()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for x := range markets {
		pairs = append(pairs, markets[x].MarketID)
	}
	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *BTCMarkets) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := b.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}
	p, err := currency.NewPairsFromStrings(pairs)
	if err != nil {
		return err
	}

	return b.UpdatePairs(p, asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *BTCMarkets) UpdateTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	allPairs, err := b.GetEnabledPairs(assetType)
	if err != nil {
		return nil, err
	}

	tickers, err := b.GetTickers(allPairs)
	if err != nil {
		return nil, err
	}

	if len(allPairs) != len(tickers) {
		return nil, errors.New("enabled pairs differ from returned tickers")
	}

	for x := range tickers {
		var newP currency.Pair
		newP, err = currency.NewPairFromString(tickers[x].MarketID)
		if err != nil {
			return nil, err
		}

		err = ticker.ProcessTicker(&ticker.Price{
			Pair:         newP,
			Last:         tickers[x].LastPrice,
			High:         tickers[x].High24h,
			Low:          tickers[x].Low24h,
			Bid:          tickers[x].BestBID,
			Ask:          tickers[x].BestAsk,
			Volume:       tickers[x].Volume,
			LastUpdated:  time.Now(),
			ExchangeName: b.Name,
			AssetType:    assetType,
		})
		if err != nil {
			return nil, err
		}
	}
	return ticker.GetTicker(b.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *BTCMarkets) FetchTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	tickerNew, err := ticker.GetTicker(b.Name, fPair, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (b *BTCMarkets) FetchOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	ob, err := orderbook.Get(b.Name, fPair, assetType)
	if err != nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *BTCMarkets) UpdateOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:         b.Name,
		Pair:             p,
		Asset:            assetType,
		PriceDuplication: true,
		VerifyOrderbook:  b.CanVerifyOrderbook,
	}

	fpair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return book, err
	}

	tempResp, err := b.GetOrderbook(fpair.String(), 2)
	if err != nil {
		return book, err
	}

	for x := range tempResp.Bids {
		book.Bids = append(book.Bids, orderbook.Item{
			Amount: tempResp.Bids[x].Volume,
			Price:  tempResp.Bids[x].Price})
	}
	for y := range tempResp.Asks {
		book.Asks = append(book.Asks, orderbook.Item{
			Amount: tempResp.Asks[y].Volume,
			Price:  tempResp.Asks[y].Price})
	}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(b.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (b *BTCMarkets) UpdateAccountInfo(assetType asset.Item) (account.Holdings, error) {
	var resp account.Holdings
	data, err := b.GetAccountBalance()
	if err != nil {
		return resp, err
	}
	var acc account.SubAccount
	for key := range data {
		c := currency.NewCode(data[key].AssetName)
		hold := data[key].Locked
		total := data[key].Balance
		acc.Currencies = append(acc.Currencies,
			account.Balance{CurrencyName: c,
				TotalValue: total,
				Hold:       hold})
	}
	resp.Accounts = append(resp.Accounts, acc)
	resp.Exchange = b.Name

	err = account.Process(&resp)
	if err != nil {
		return account.Holdings{}, err
	}

	return resp, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (b *BTCMarkets) FetchAccountInfo(assetType asset.Item) (account.Holdings, error) {
	acc, err := account.GetHoldings(b.Name, assetType)
	if err != nil {
		return b.UpdateAccountInfo(assetType)
	}

	return acc, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *BTCMarkets) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetWithdrawalsHistory returns previous withdrawals data
func (b *BTCMarkets) GetWithdrawalsHistory(c currency.Code) (resp []exchange.WithdrawalHistory, err error) {
	return nil, common.ErrNotYetImplemented
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (b *BTCMarkets) GetRecentTrades(p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	var err error
	p, err = b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	var resp []trade.Data
	var tradeData []Trade
	tradeData, err = b.GetTrades(p.String(), 0, 0, 200)
	if err != nil {
		return nil, err
	}
	for i := range tradeData {
		side := order.Side("")
		if tradeData[i].Side != "" {
			side, err = order.StringToOrderSide(tradeData[i].Side)
			if err != nil {
				return nil, err
			}
		}
		resp = append(resp, trade.Data{
			Exchange:     b.Name,
			TID:          tradeData[i].TradeID,
			CurrencyPair: p,
			AssetType:    assetType,
			Side:         side,
			Price:        tradeData[i].Price,
			Amount:       tradeData[i].Amount,
			Timestamp:    tradeData[i].Timestamp,
		})
	}

	err = b.AddTradesToBuffer(resp...)
	if err != nil {
		return nil, err
	}

	sort.Sort(trade.ByDate(resp))
	return resp, nil
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (b *BTCMarkets) GetHistoricTrades(_ currency.Pair, _ asset.Item, _, _ time.Time) ([]trade.Data, error) {
	return nil, common.ErrFunctionNotSupported
}

// SubmitOrder submits a new order
func (b *BTCMarkets) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var resp order.SubmitResponse
	if err := s.Validate(); err != nil {
		return resp, err
	}

	if s.Side == order.Sell {
		s.Side = order.Ask
	}
	if s.Side == order.Buy {
		s.Side = order.Bid
	}

	fpair, err := b.FormatExchangeCurrency(s.Pair, asset.Spot)
	if err != nil {
		return resp, err
	}

	tempResp, err := b.NewOrder(fpair.String(),
		s.Price,
		s.Amount,
		s.Type.String(),
		s.Side.String(),
		s.TriggerPrice,
		s.TargetAmount,
		"",
		false,
		"",
		s.ClientID)
	if err != nil {
		return resp, err
	}
	resp.IsOrderPlaced = true
	resp.OrderID = tempResp.OrderID
	return resp, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *BTCMarkets) ModifyOrder(action *order.Modify) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (b *BTCMarkets) CancelOrder(o *order.Cancel) error {
	err := o.Validate(o.StandardCancel())
	if err != nil {
		return err
	}
	_, err = b.RemoveOrder(o.ID)
	return err
}

// CancelBatchOrders cancels an orders by their corresponding ID numbers
func (b *BTCMarkets) CancelBatchOrders(o []order.Cancel) (order.CancelBatchResponse, error) {
	return order.CancelBatchResponse{}, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *BTCMarkets) CancelAllOrders(_ *order.Cancel) (order.CancelAllResponse, error) {
	var resp order.CancelAllResponse
	tempMap := make(map[string]string)
	var orderIDs []string
	orders, err := b.GetOrders("", -1, -1, -1, true)
	if err != nil {
		return resp, err
	}
	for x := range orders {
		orderIDs = append(orderIDs, orders[x].OrderID)
	}
	splitOrders := common.SplitStringSliceByLimit(orderIDs, 20)
	for z := range splitOrders {
		tempResp, err := b.CancelBatch(splitOrders[z])
		if err != nil {
			return resp, err
		}
		for y := range tempResp.CancelOrders {
			tempMap[tempResp.CancelOrders[y].OrderID] = "Success"
		}
		for z := range tempResp.UnprocessedRequests {
			tempMap[tempResp.UnprocessedRequests[z].RequestID] = "Cancellation Failed"
		}
	}
	resp.Status = tempMap
	return resp, nil
}

// GetOrderInfo returns order information based on order ID
func (b *BTCMarkets) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	var resp order.Detail
	o, err := b.FetchOrder(orderID)
	if err != nil {
		return resp, err
	}

	p, err := currency.NewPairFromString(o.MarketID)
	if err != nil {
		return order.Detail{}, err
	}

	resp.Exchange = b.Name
	resp.ID = orderID
	resp.Pair = p
	resp.Price = o.Price
	resp.Date = o.CreationTime
	resp.ExecutedAmount = o.Amount - o.OpenAmount
	resp.Side = order.Bid
	if o.Side == ask {
		resp.Side = order.Ask
	}
	switch o.Type {
	case limit:
		resp.Type = order.Limit
	case market:
		resp.Type = order.Market
	case stopLimit:
		resp.Type = order.Stop
	case stop:
		resp.Type = order.Stop
	case takeProfit:
		resp.Type = order.ImmediateOrCancel
	default:
		resp.Type = order.UnknownType
	}
	resp.RemainingAmount = o.OpenAmount
	switch o.Status {
	case orderAccepted:
		resp.Status = order.Active
	case orderPlaced:
		resp.Status = order.Active
	case orderPartiallyMatched:
		resp.Status = order.PartiallyFilled
	case orderFullyMatched:
		resp.Status = order.Filled
	case orderCancelled:
		resp.Status = order.Cancelled
	case orderPartiallyCancelled:
		resp.Status = order.PartiallyCancelled
	case orderFailed:
		resp.Status = order.Rejected
	default:
		resp.Status = order.UnknownStatus
	}
	return resp, nil
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *BTCMarkets) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	temp, err := b.FetchDepositAddress(strings.ToUpper(cryptocurrency.String()), -1, -1, -1)
	if err != nil {
		return "", err
	}
	return temp.Address, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is submitted
func (b *BTCMarkets) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}
	a, err := b.RequestWithdraw(withdrawRequest.Currency.String(),
		withdrawRequest.Amount,
		withdrawRequest.Crypto.Address,
		"",
		"",
		"",
		"")
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		ID:     a.ID,
		Status: a.Status,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (b *BTCMarkets) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}

	if withdrawRequest.Currency != currency.AUD {
		return nil, errors.New("only aud is supported for withdrawals")
	}
	a, err := b.RequestWithdraw(withdrawRequest.Currency.String(),
		withdrawRequest.Amount,
		"",
		withdrawRequest.Fiat.Bank.AccountName,
		withdrawRequest.Fiat.Bank.AccountNumber,
		withdrawRequest.Fiat.Bank.BSBNumber,
		withdrawRequest.Fiat.Bank.BankName)
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		ID:     a.ID,
		Status: a.Status,
	}, nil
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (b *BTCMarkets) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *BTCMarkets) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !b.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (b *BTCMarkets) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if len(req.Pairs) == 0 {
		allPairs, err := b.GetEnabledPairs(asset.Spot)
		if err != nil {
			return nil, err
		}
		for a := range allPairs {
			req.Pairs = append(req.Pairs,
				allPairs[a])
		}
	}

	var resp []order.Detail
	for x := range req.Pairs {
		fpair, err := b.FormatExchangeCurrency(req.Pairs[x], asset.Spot)
		if err != nil {
			return nil, err
		}
		tempData, err := b.GetOrders(fpair.String(), -1, -1, -1, true)
		if err != nil {
			return resp, err
		}
		for y := range tempData {
			var tempResp order.Detail
			tempResp.Exchange = b.Name
			tempResp.Pair = req.Pairs[x]
			tempResp.ID = tempData[y].OrderID
			tempResp.Side = order.Bid
			if tempData[y].Side == ask {
				tempResp.Side = order.Ask
			}
			tempResp.Date = tempData[y].CreationTime

			switch tempData[y].Type {
			case limit:
				tempResp.Type = order.Limit
			case market:
				tempResp.Type = order.Market
			default:
				log.Errorf(log.ExchangeSys,
					"%s unknown order type %s getting order",
					b.Name,
					tempData[y].Type)
				tempResp.Type = order.UnknownType
			}
			switch tempData[y].Status {
			case orderAccepted:
				tempResp.Status = order.Active
			case orderPlaced:
				tempResp.Status = order.Active
			case orderPartiallyMatched:
				tempResp.Status = order.PartiallyFilled
			default:
				log.Errorf(log.ExchangeSys,
					"%s unexpected status %s on order %v",
					b.Name,
					tempData[y].Status,
					tempData[y].OrderID)
				tempResp.Status = order.UnknownStatus
			}
			tempResp.Price = tempData[y].Price
			tempResp.Amount = tempData[y].Amount
			tempResp.ExecutedAmount = tempData[y].Amount - tempData[y].OpenAmount
			tempResp.RemainingAmount = tempData[y].OpenAmount
			resp = append(resp, tempResp)
		}
	}
	order.FilterOrdersByType(&resp, req.Type)
	order.FilterOrdersByTimeRange(&resp, req.StartTime, req.EndTime)
	order.FilterOrdersBySide(&resp, req.Side)
	return resp, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *BTCMarkets) GetOrderHistory(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var resp []order.Detail
	var tempResp order.Detail
	var tempArray []string
	if len(req.Pairs) == 0 {
		orders, err := b.GetOrders("", -1, -1, -1, false)
		if err != nil {
			return resp, err
		}
		for x := range orders {
			tempArray = append(tempArray, orders[x].OrderID)
		}
	}
	for y := range req.Pairs {
		fpair, err := b.FormatExchangeCurrency(req.Pairs[y], asset.Spot)
		if err != nil {
			return nil, err
		}

		orders, err := b.GetOrders(fpair.String(), -1, -1, -1, false)
		if err != nil {
			return resp, err
		}
		for z := range orders {
			tempArray = append(tempArray, orders[z].OrderID)
		}
	}
	splitOrders := common.SplitStringSliceByLimit(tempArray, 50)
	for x := range splitOrders {
		tempData, err := b.GetBatchTrades(splitOrders[x])
		if err != nil {
			return resp, err
		}
		for c := range tempData.Orders {
			switch tempData.Orders[c].Status {
			case orderFailed:
				tempResp.Status = order.Rejected
			case orderPartiallyCancelled:
				tempResp.Status = order.PartiallyCancelled
			case orderCancelled:
				tempResp.Status = order.Cancelled
			case orderFullyMatched:
				tempResp.Status = order.Filled
			case orderPartiallyMatched:
				continue
			case orderPlaced:
				continue
			case orderAccepted:
				continue
			}

			p, err := currency.NewPairFromString(tempData.Orders[c].MarketID)
			if err != nil {
				return nil, err
			}

			tempResp.Exchange = b.Name
			tempResp.Pair = p
			tempResp.Side = order.Bid
			if tempData.Orders[c].Side == ask {
				tempResp.Side = order.Ask
			}
			tempResp.ID = tempData.Orders[c].OrderID
			tempResp.Date = tempData.Orders[c].CreationTime
			tempResp.Price = tempData.Orders[c].Price
			tempResp.ExecutedAmount = tempData.Orders[c].Amount
			resp = append(resp, tempResp)
		}
	}
	return resp, nil
}

// ValidateCredentials validates current credentials used for wrapper
// functionality
func (b *BTCMarkets) ValidateCredentials(assetType asset.Item) error {
	_, err := b.UpdateAccountInfo(assetType)
	if err != nil {
		if b.CheckTransientError(err) == nil {
			return nil
		}
		// Check for specific auth errors; all other errors can be disregarded
		// as this does not affect authenticated requests.
		if strings.Contains(err.Error(), "InvalidAPIKey") ||
			strings.Contains(err.Error(), "InvalidAuthTimestamp") ||
			strings.Contains(err.Error(), "InvalidAuthSignature") ||
			strings.Contains(err.Error(), "InsufficientAPIPermission") {
			return err
		}
	}

	return nil
}

// FormatExchangeKlineInterval returns Interval to exchange formatted string
func (b *BTCMarkets) FormatExchangeKlineInterval(in kline.Interval) string {
	if in == kline.OneDay {
		return "1d"
	}
	return in.Short()
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (b *BTCMarkets) GetHistoricCandles(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := b.ValidateKline(pair, a, interval); err != nil {
		return kline.Item{}, err
	}

	if kline.TotalCandlesPerInterval(start, end, interval) > float64(b.Features.Enabled.Kline.ResultLimit) {
		return kline.Item{}, errors.New(kline.ErrRequestExceedsExchangeLimits)
	}

	formattedPair, err := b.FormatExchangeCurrency(pair, a)
	if err != nil {
		return kline.Item{}, err
	}

	candles, err := b.GetMarketCandles(formattedPair.String(),
		b.FormatExchangeKlineInterval(interval),
		start,
		end,
		-1,
		-1,
		-1)

	if err != nil {
		return kline.Item{}, err
	}
	ret := kline.Item{
		Exchange: b.Name,
		Pair:     formattedPair,
		Asset:    asset.Spot,
		Interval: interval,
	}

	for x := range candles {
		var tempTime time.Time
		var tempData kline.Candle
		tempTime, err = time.Parse(time.RFC3339, candles[x][0])
		if err != nil {
			return kline.Item{}, err
		}
		tempData.Time = tempTime
		tempData.Open, err = strconv.ParseFloat(candles[x][1], 64)
		if err != nil {
			return kline.Item{}, err
		}
		tempData.High, err = strconv.ParseFloat(candles[x][2], 64)
		if err != nil {
			return kline.Item{}, err
		}
		tempData.Low, err = strconv.ParseFloat(candles[x][3], 64)
		if err != nil {
			return kline.Item{}, err
		}
		tempData.Close, err = strconv.ParseFloat(candles[x][4], 64)
		if err != nil {
			return kline.Item{}, err
		}
		tempData.Volume, err = strconv.ParseFloat(candles[x][5], 64)
		if err != nil {
			return kline.Item{}, err
		}
		ret.Candles = append(ret.Candles, tempData)
	}

	ret.SortCandlesByTimestamp(false)
	return ret, nil
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (b *BTCMarkets) GetHistoricCandlesExtended(p currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := b.ValidateKline(p, a, interval); err != nil {
		return kline.Item{}, err
	}

	fPair, err := b.FormatExchangeCurrency(p, a)
	if err != nil {
		return kline.Item{}, err
	}

	ret := kline.Item{
		Exchange: b.Name,
		Pair:     fPair,
		Asset:    a,
		Interval: interval,
	}

	dates := kline.CalculateCandleDateRanges(start, end, interval, b.Features.Enabled.Kline.ResultLimit)
	for x := range dates.Ranges {
		var candles CandleResponse
		candles, err = b.GetMarketCandles(fPair.String(),
			b.FormatExchangeKlineInterval(interval),
			dates.Ranges[x].Start.Time, dates.Ranges[x].End.Time, -1, -1, -1)
		if err != nil {
			return kline.Item{}, err
		}

		for i := range candles {
			var tempTime time.Time
			var tempData kline.Candle
			tempTime, err = time.Parse(time.RFC3339, candles[i][0])
			if err != nil {
				return kline.Item{}, err
			}
			tempData.Time = tempTime
			tempData.Open, err = strconv.ParseFloat(candles[i][1], 64)
			if err != nil {
				return kline.Item{}, err
			}
			tempData.High, err = strconv.ParseFloat(candles[i][2], 64)
			if err != nil {
				return kline.Item{}, err
			}
			tempData.Low, err = strconv.ParseFloat(candles[i][3], 64)
			if err != nil {
				return kline.Item{}, err
			}
			tempData.Close, err = strconv.ParseFloat(candles[i][4], 64)
			if err != nil {
				return kline.Item{}, err
			}
			tempData.Volume, err = strconv.ParseFloat(candles[i][5], 64)
			if err != nil {
				return kline.Item{}, err
			}
			ret.Candles = append(ret.Candles, tempData)
		}
	}

	err = dates.VerifyResultsHaveData(ret.Candles)
	if err != nil {
		log.Warnf(log.ExchangeSys, "%s - %s", b.Name, err)
	}
	ret.RemoveDuplicates()
	ret.RemoveOutsideRange(start, end)
	ret.SortCandlesByTimestamp(false)
	return ret, nil
}
