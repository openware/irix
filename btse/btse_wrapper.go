package btse

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openware/pkg/common"
	"github.com/openware/irix/config"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
	"github.com/openware/irix/portfolio/withdraw"
	exchange "github.com/openware/irix"
	"github.com/openware/pkg/account"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/kline"
	"github.com/openware/pkg/order"
	"github.com/openware/pkg/orderbook"
	"github.com/openware/irix/protocol"
	"github.com/openware/pkg/request"
	"github.com/openware/irix/stream"
	"github.com/openware/irix/ticker"
	"github.com/openware/pkg/trade"
)

const (
	spotURL   = "spotURL"
	spotWSURL = "websocketURL"
)

// GetDefaultConfig returns a default exchange config
func (b *BTSE) GetDefaultConfig() (*config.ExchangeConfig, error) {
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

// SetDefaults sets the basic defaults for BTSE
func (b *BTSE) SetDefaults() {
	b.Name = "BTSE"
	b.Enabled = true
	b.Verbose = true
	b.API.CredentialsValidator.RequiresKey = true
	b.API.CredentialsValidator.RequiresSecret = true

	fmt1 := currency.PairStore{
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: "-",
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: "-",
		},
	}
	err := b.StoreAssetPairFormat(asset.Spot, fmt1)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	fmt2 := currency.PairStore{
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
		},
	}
	err = b.StoreAssetPairFormat(asset.Futures, fmt2)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	b.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:      true,
				TickerBatching:      true,
				KlineFetching:       true,
				TradeFetching:       true,
				OrderbookFetching:   true,
				AutoPairUpdates:     true,
				AccountInfo:         true,
				GetOrder:            true,
				GetOrders:           true,
				CancelOrders:        true,
				CancelOrder:         true,
				SubmitOrder:         true,
				TradeFee:            true,
				FiatDepositFee:      true,
				FiatWithdrawalFee:   true,
				CryptoWithdrawalFee: true,
			},
			WebsocketCapabilities: protocol.Features{
				OrderbookFetching: true,
				TradeFetching:     true,
				Subscribe:         true,
				Unsubscribe:       true,
				GetOrders:         true,
				GetOrder:          true,
			},
			WithdrawPermissions: exchange.NoAPIWithdrawalMethods,
			Kline: kline.ExchangeCapabilitiesSupported{
				DateRanges: true,
				Intervals:  true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: map[string]bool{
					kline.OneMin.Word():     true,
					kline.ThreeMin.Word():   true,
					kline.FiveMin.Word():    true,
					kline.FifteenMin.Word(): true,
					kline.ThirtyMin.Word():  true,
					kline.OneHour.Word():    true,
					kline.TwoHour.Word():    true,
					kline.FourHour.Word():   true,
					kline.SixHour.Word():    true,
					kline.TwelveHour.Word(): true,
					kline.OneDay.Word():     true,
					kline.ThreeDay.Word():   true,
					kline.OneWeek.Word():    true,
					kline.OneMonth.Word():   true,
				},
				ResultLimit: 300,
			},
		},
	}

	b.Requester = request.New(b.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	b.API.Endpoints = b.NewEndpoints()
	err = b.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      btseAPIURL,
		exchange.RestFutures:   btseAPIURL,
		exchange.WebsocketSpot: btseWebsocket,
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
func (b *BTSE) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	err := b.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsRunningURL, err := b.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = b.Websocket.Setup(&stream.WebsocketSetup{
		Enabled:                          exch.Features.Enabled.Websocket,
		Verbose:                          exch.Verbose,
		AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
		WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
		DefaultURL:                       btseWebsocket,
		ExchangeName:                     exch.Name,
		RunningURL:                       wsRunningURL,
		Connector:                        b.WsConnect,
		Subscriber:                       b.Subscribe,
		UnSubscriber:                     b.Unsubscribe,
		GenerateSubscriptions:            b.GenerateDefaultSubscriptions,
		Features:                         &b.Features.Supports.WebsocketCapabilities,
		OrderbookBufferLimit:             exch.OrderbookConfig.WebsocketBufferLimit,
		BufferEnabled:                    exch.OrderbookConfig.WebsocketBufferEnabled,
	})
	if err != nil {
		return err
	}

	err = b.seedOrderSizeLimits()
	if err != nil {
		return err
	}

	return b.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the BTSE go routine
func (b *BTSE) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the BTSE wrapper
func (b *BTSE) Run() {
	if b.Verbose {
		b.PrintEnabledPairs()
	}

	if !b.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := b.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s Failed to update tradable pairs. Error: %s", b.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *BTSE) FetchTradablePairs(a asset.Item) ([]string, error) {
	var currencies []string
	m, err := b.GetMarketSummary("", a == asset.Spot)
	if err != nil {
		return nil, err
	}

	for x := range m {
		if !m[x].Active {
			continue
		}
		currencies = append(currencies, m[x].Symbol)
	}
	return currencies, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *BTSE) UpdateTradablePairs(forceUpdate bool) error {
	a := b.GetAssetTypes()
	for i := range a {
		pairs, err := b.FetchTradablePairs(a[i])
		if err != nil {
			return err
		}

		p, err := currency.NewPairsFromStrings(pairs)
		if err != nil {
			return err
		}

		err = b.UpdatePairs(p, a[i], false, forceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *BTSE) UpdateTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickers, err := b.GetMarketSummary("", assetType == asset.Spot)
	if err != nil {
		return nil, err
	}
	for x := range tickers {
		var pair currency.Pair
		pair, err = currency.NewPairFromString(tickers[x].Symbol)
		if err != nil {
			return nil, err
		}

		err = ticker.ProcessTicker(&ticker.Price{
			Pair:         pair,
			Ask:          tickers[x].LowestAsk,
			Bid:          tickers[x].HighestBid,
			Low:          tickers[x].Low24Hr,
			Last:         tickers[x].Last,
			Volume:       tickers[x].Volume,
			High:         tickers[x].High24Hr,
			ExchangeName: b.Name,
			AssetType:    assetType})
		if err != nil {
			return nil, err
		}
	}

	return ticker.GetTicker(b.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *BTSE) FetchTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(b.Name, p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (b *BTSE) FetchOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(b.Name, p, assetType)
	if err != nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *BTSE) UpdateOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        b.Name,
		Pair:            p,
		Asset:           assetType,
		VerifyOrderbook: b.CanVerifyOrderbook,
	}
	fPair, err := b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return book, err
	}
	a, err := b.FetchOrderBook(fPair.String(), 0, 0, 0, assetType == asset.Spot)
	if err != nil {
		return book, err
	}

	for x := range a.BuyQuote {
		if b.orderbookFilter(a.BuyQuote[x].Price, a.BuyQuote[x].Size) {
			continue
		}
		book.Bids = append(book.Bids, orderbook.Item{
			Price:  a.BuyQuote[x].Price,
			Amount: a.BuyQuote[x].Size})
	}
	for x := range a.SellQuote {
		if b.orderbookFilter(a.SellQuote[x].Price, a.SellQuote[x].Size) {
			continue
		}
		book.Asks = append(book.Asks, orderbook.Item{
			Price:  a.SellQuote[x].Price,
			Amount: a.SellQuote[x].Size})
	}
	book.Asks.Reverse() // Reverse asks for correct alignment
	book.Pair = p
	book.Exchange = b.Name
	book.Asset = assetType
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(b.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies for the
// BTSE exchange
func (b *BTSE) UpdateAccountInfo(assetType asset.Item) (account.Holdings, error) {
	var a account.Holdings
	balance, err := b.GetWalletInformation()
	if err != nil {
		return a, err
	}

	var currencies []account.Balance
	for b := range balance {
		currencies = append(currencies,
			account.Balance{
				CurrencyName: currency.NewCode(balance[b].Currency),
				TotalValue:   balance[b].Total,
				Hold:         balance[b].Available,
			},
		)
	}
	a.Exchange = b.Name
	a.Accounts = []account.SubAccount{
		{
			Currencies: currencies,
		},
	}

	err = account.Process(&a)
	if err != nil {
		return account.Holdings{}, err
	}

	return a, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (b *BTSE) FetchAccountInfo(assetType asset.Item) (account.Holdings, error) {
	acc, err := account.GetHoldings(b.Name, assetType)
	if err != nil {
		return b.UpdateAccountInfo(assetType)
	}

	return acc, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *BTSE) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

func (b *BTSE) withinLimits(pair currency.Pair, amount float64) bool {
	val, found := OrderSizeLimits(pair.String())
	if !found {
		return false
	}
	return (math.Mod(amount, val.MinSizeIncrement) == 0) ||
		amount < val.MinOrderSize ||
		amount > val.MaxOrderSize
}

// GetWithdrawalsHistory returns previous withdrawals data
func (b *BTSE) GetWithdrawalsHistory(c currency.Code) (resp []exchange.WithdrawalHistory, err error) {
	return nil, common.ErrNotYetImplemented
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (b *BTSE) GetRecentTrades(p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	var err error
	p, err = b.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	var resp []trade.Data
	limit := 500

	var tradeData []Trade
	tradeData, err = b.GetTrades(p.String(),
		time.Time{}, time.Time{},
		0, 0, limit,
		false,
		assetType == asset.Spot)
	if err != nil {
		return nil, err
	}
	for i := range tradeData {
		tradeTimestamp := time.Unix(tradeData[i].Time/1000, 0)
		var side order.Side
		side, err = order.StringToOrderSide(tradeData[i].Side)
		if err != nil {
			return nil, err
		}
		resp = append(resp, trade.Data{
			Exchange:     b.Name,
			TID:          strconv.FormatInt(tradeData[i].SerialID, 10),
			CurrencyPair: p,
			AssetType:    assetType,
			Side:         side,
			Price:        tradeData[i].Price,
			Amount:       tradeData[i].Amount,
			Timestamp:    tradeTimestamp,
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
func (b *BTSE) GetHistoricTrades(_ currency.Pair, _ asset.Item, _, _ time.Time) ([]trade.Data, error) {
	return nil, common.ErrFunctionNotSupported
}

// SubmitOrder submits a new order
func (b *BTSE) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var resp order.SubmitResponse
	if err := s.Validate(); err != nil {
		return resp, err
	}

	fPair, err := b.FormatExchangeCurrency(s.Pair, s.AssetType)
	if err != nil {
		return resp, err
	}
	inLimits := b.withinLimits(fPair, s.Amount)
	if !inLimits {
		return resp, errors.New("order outside of limits")
	}

	r, err := b.CreateOrder(s.ClientID, 0.0,
		false,
		s.Price, s.Side.String(), s.Amount, 0, 0,
		fPair.String(), goodTillCancel,
		0.0, s.TriggerPrice,
		"", s.Type.String())
	if err != nil {
		return resp, err
	}

	resp.IsOrderPlaced = true
	resp.OrderID = r[0].OrderID

	if s.Type == order.Market {
		resp.FullyMatched = true
	}
	return resp, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *BTSE) ModifyOrder(action *order.Modify) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (b *BTSE) CancelOrder(o *order.Cancel) error {
	if err := o.Validate(o.StandardCancel()); err != nil {
		return err
	}

	fPair, err := b.FormatExchangeCurrency(o.Pair,
		o.AssetType)
	if err != nil {
		return err
	}

	_, err = b.CancelExistingOrder(o.ID, fPair.String(), o.ClientOrderID)
	if err != nil {
		return err
	}

	return nil
}

// CancelBatchOrders cancels an orders by their corresponding ID numbers
func (b *BTSE) CancelBatchOrders(o []order.Cancel) (order.CancelBatchResponse, error) {
	return order.CancelBatchResponse{}, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
// If product ID is sent, all orders of that specified market will be cancelled
// If not specified, all orders of all markets will be cancelled
func (b *BTSE) CancelAllOrders(orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	if err := orderCancellation.Validate(); err != nil {
		return order.CancelAllResponse{}, err
	}

	var resp order.CancelAllResponse

	fPair, err := b.FormatExchangeCurrency(orderCancellation.Pair,
		orderCancellation.AssetType)
	if err != nil {
		return resp, err
	}

	allOrders, err := b.CancelExistingOrder("", fPair.String(), "")
	if err != nil {
		return resp, nil
	}

	resp.Status = make(map[string]string)
	for x := range allOrders {
		if allOrders[x].Status == orderCancelled {
			resp.Status[allOrders[x].OrderID] = order.Cancelled.String()
		}
	}
	return resp, nil
}

func orderIntToType(i int) order.Type {
	if i == 77 {
		return order.Market
	} else if i == 76 {
		return order.Limit
	}
	return order.UnknownType
}

// GetOrderInfo returns order information based on order ID
func (b *BTSE) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	o, err := b.GetOrders("", orderID, "")
	if err != nil {
		return order.Detail{}, err
	}

	var od order.Detail
	if len(o) == 0 {
		return od, errors.New("no orders found")
	}

	format, err := b.GetPairFormat(asset.Spot, false)
	if err != nil {
		return order.Detail{}, err
	}

	for i := range o {
		if o[i].OrderID != orderID {
			continue
		}

		var side = order.Buy
		if strings.EqualFold(o[i].Side, order.Ask.String()) {
			side = order.Sell
		}

		od.Pair, err = currency.NewPairDelimiter(o[i].Symbol,
			format.Delimiter)
		if err != nil {
			log.Errorf(log.ExchangeSys,
				"%s GetOrderInfo unable to parse currency pair: %s\n",
				b.Name,
				err)
		}
		od.Exchange = b.Name
		od.Amount = o[i].Size
		od.ID = o[i].OrderID
		od.Date = time.Unix(o[i].Timestamp, 0)
		od.Side = side

		od.Type = orderIntToType(o[i].OrderType)

		od.Price = o[i].Price
		od.Status = order.Status(o[i].OrderState)

		th, err := b.TradeHistory("",
			time.Time{}, time.Time{},
			0, 0, 0,
			false,
			"", orderID)
		if err != nil {
			return od,
				fmt.Errorf("unable to get order fills for orderID %s", orderID)
		}

		for i := range th {
			createdAt, err := parseOrderTime(th[i].TradeID)
			if err != nil {
				log.Errorf(log.ExchangeSys,
					"%s GetOrderInfo unable to parse time: %s\n", b.Name, err)
			}
			od.Trades = append(od.Trades, order.TradeHistory{
				Timestamp: createdAt,
				TID:       th[i].TradeID,
				Price:     th[i].Price,
				Amount:    th[i].Size,
				Exchange:  b.Name,
				Side:      order.Side(th[i].Side),
				Fee:       th[i].FeeAmount,
			})
		}
	}
	return od, nil
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *BTSE) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	address, err := b.GetWalletAddress(cryptocurrency.String())
	if err != nil {
		return "", err
	}
	if len(address) == 0 {
		addressCreate, err := b.CreateWalletAddress(cryptocurrency.String())
		if err != nil {
			return "", err
		}
		if len(addressCreate) != 0 {
			return addressCreate[0].Address, nil
		}
		return "", errors.New("address not found")
	}
	return address[0].Address, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *BTSE) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}

	amountToString := strconv.FormatFloat(withdrawRequest.Amount, 'f', 8, 64)
	resp, err := b.WalletWithdrawal(withdrawRequest.Currency.String(),
		withdrawRequest.Crypto.Address,
		withdrawRequest.Crypto.AddressTag,
		amountToString)
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		Name: b.Name,
		ID:   resp.WithdrawID,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *BTSE) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (b *BTSE) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetActiveOrders retrieves any orders that are active/open
func (b *BTSE) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if len(req.Pairs) == 0 {
		return nil, errors.New("no pair provided")
	}

	var orders []order.Detail
	for x := range req.Pairs {
		formattedPair, err := b.FormatExchangeCurrency(req.Pairs[x], asset.Spot)
		if err != nil {
			return nil, err
		}
		resp, err := b.GetOrders(formattedPair.String(), "", "")
		if err != nil {
			return nil, err
		}

		format, err := b.GetPairFormat(asset.Spot, false)
		if err != nil {
			return nil, err
		}

		for i := range resp {
			var side = order.Buy
			if strings.EqualFold(resp[i].Side, order.Ask.String()) {
				side = order.Sell
			}

			p, err := currency.NewPairDelimiter(resp[i].Symbol,
				format.Delimiter)
			if err != nil {
				log.Errorf(log.ExchangeSys,
					"%s GetActiveOrders unable to parse currency pair: %s\n",
					b.Name,
					err)
			}

			openOrder := order.Detail{
				Pair:     p,
				Exchange: b.Name,
				Amount:   resp[i].Size,
				ID:       resp[i].OrderID,
				Date:     time.Unix(resp[i].Timestamp, 0),
				Side:     side,
				Price:    resp[i].Price,
				Status:   order.Status(resp[i].OrderState),
			}

			if resp[i].OrderType == 77 {
				openOrder.Type = order.Market
			} else if resp[i].OrderType == 76 {
				openOrder.Type = order.Limit
			}

			fills, err := b.TradeHistory(
				"",
				time.Time{}, time.Time{},
				0, 0, 0,
				false,
				"", resp[i].OrderID)
			if err != nil {
				log.Errorf(log.ExchangeSys,
					"%s: Unable to get order fills for orderID %s",
					b.Name,
					resp[i].OrderID)
				continue
			}

			for i := range fills {
				createdAt, err := parseOrderTime(fills[i].Timestamp)
				if err != nil {
					log.Errorf(log.ExchangeSys,
						"%s GetActiveOrders unable to parse time: %s\n",
						b.Name,
						err)
				}
				openOrder.Trades = append(openOrder.Trades, order.TradeHistory{
					Timestamp: createdAt,
					TID:       fills[i].TradeID,
					Price:     fills[i].Price,
					Amount:    fills[i].Size,
					Exchange:  b.Name,
					Side:      order.Side(fills[i].Side),
					Fee:       fills[i].FeeAmount,
				})
			}
			orders = append(orders, openOrder)
		}
	}

	order.FilterOrdersByType(&orders, req.Type)
	order.FilterOrdersByTimeRange(&orders, req.StartTime, req.EndTime)
	order.FilterOrdersBySide(&orders, req.Side)
	return orders, nil
}

func matchType(input int, required order.Type) bool {
	if (required == order.AnyType) || (input == 76 && required == order.Limit) || input == 77 && required == order.Market {
		return true
	}
	return false
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *BTSE) GetOrderHistory(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}

	var resp []order.Detail
	if len(getOrdersRequest.Pairs) == 0 {
		var err error
		getOrdersRequest.Pairs, err = b.GetEnabledPairs(asset.Spot)
		if err != nil {
			return nil, err
		}
	}
	orderDeref := *getOrdersRequest
	for x := range orderDeref.Pairs {
		fPair, err := b.FormatExchangeCurrency(orderDeref.Pairs[x], asset.Spot)
		if err != nil {
			return nil, err
		}
		currentOrder, err := b.GetOrders(fPair.String(), "", "")
		if err != nil {
			return nil, err
		}
		for y := range currentOrder {
			if !matchType(currentOrder[y].OrderType, orderDeref.Type) {
				continue
			}
			tempOrder := order.Detail{
				Price:  currentOrder[y].Price,
				Amount: currentOrder[y].Size,
				Side:   order.Side(currentOrder[y].Side),
				Pair:   orderDeref.Pairs[x],
			}
			switch currentOrder[x].OrderState {
			case "STATUS_ACTIVE":
				tempOrder.Status = order.Active
			case "ORDER_CANCELLED":
				tempOrder.Status = order.Cancelled
			case "ORDER_FULLY_TRANSACTED":
				tempOrder.Status = order.Filled
			case "ORDER_PARTIALLY_TRANSACTED":
				tempOrder.Status = order.PartiallyFilled
			default:
				tempOrder.Status = order.UnknownStatus
			}
			resp = append(resp, tempOrder)
		}
	}
	return resp, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *BTSE) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !b.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)
}

// ValidateCredentials validates current credentials used for wrapper
// functionality
func (b *BTSE) ValidateCredentials(assetType asset.Item) error {
	_, err := b.UpdateAccountInfo(assetType)
	return b.CheckTransientError(err)
}

// FormatExchangeKlineInterval formats kline interval to exchange requested type
func (b *BTSE) FormatExchangeKlineInterval(in kline.Interval) string {
	return strconv.FormatFloat(in.Duration().Minutes(), 'f', 0, 64)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (b *BTSE) GetHistoricCandles(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := b.ValidateKline(pair, a, interval); err != nil {
		return kline.Item{}, err
	}

	fPair, err := b.FormatExchangeCurrency(pair, a)
	if err != nil {
		return kline.Item{}, err
	}
	intervalInt, err := strconv.Atoi(b.FormatExchangeKlineInterval(interval))
	if err != nil {
		return kline.Item{}, err
	}

	klineRet := kline.Item{
		Exchange: b.Name,
		Pair:     fPair,
		Asset:    a,
		Interval: interval,
	}

	switch a {
	case asset.Spot:
		req, err := b.OHLCV(fPair.String(),
			start,
			end,
			intervalInt)
		if err != nil {
			return kline.Item{}, err
		}
		for x := range req {
			klineRet.Candles = append(klineRet.Candles, kline.Candle{
				Time:   time.Unix(int64(req[x][0]), 0),
				Open:   req[x][1],
				High:   req[x][2],
				Low:    req[x][3],
				Close:  req[x][4],
				Volume: req[x][5],
			})
		}
	case asset.Futures:
		return kline.Item{}, common.ErrNotYetImplemented
	default:
		return kline.Item{}, fmt.Errorf("asset %v not supported", a.String())
	}

	klineRet.SortCandlesByTimestamp(false)
	return klineRet, nil
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (b *BTSE) GetHistoricCandlesExtended(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := b.ValidateKline(pair, a, interval); err != nil {
		return kline.Item{}, err
	}

	if kline.TotalCandlesPerInterval(start, end, interval) > float64(b.Features.Enabled.Kline.ResultLimit) {
		return kline.Item{}, errors.New(kline.ErrRequestExceedsExchangeLimits)
	}

	fPair, err := b.FormatExchangeCurrency(pair, a)
	if err != nil {
		return kline.Item{}, err
	}
	intervalInt, err := strconv.Atoi(b.FormatExchangeKlineInterval(interval))
	if err != nil {
		return kline.Item{}, err
	}

	klineRet := kline.Item{
		Exchange: b.Name,
		Pair:     fPair,
		Asset:    a,
		Interval: interval,
	}

	switch a {
	case asset.Spot:
		req, err := b.OHLCV(fPair.String(),
			start,
			end,
			intervalInt)
		if err != nil {
			return kline.Item{}, err
		}
		for x := range req {
			klineRet.Candles = append(klineRet.Candles, kline.Candle{
				Time:   time.Unix(int64(req[x][0]), 0),
				Open:   req[x][1],
				High:   req[x][2],
				Low:    req[x][3],
				Close:  req[x][4],
				Volume: req[x][5],
			})
		}
	case asset.Futures:
		return kline.Item{}, common.ErrNotYetImplemented
	default:
		return kline.Item{}, fmt.Errorf("asset %v not supported", a.String())
	}

	klineRet.SortCandlesByTimestamp(false)
	return klineRet, nil
}

func (b *BTSE) seedOrderSizeLimits() error {
	pairs, err := b.GetMarketSummary("", true)
	if err != nil {
		return err
	}
	for x := range pairs {
		tempValues := OrderSizeLimit{
			MinOrderSize:     pairs[x].MinOrderSize,
			MaxOrderSize:     pairs[x].MaxOrderSize,
			MinSizeIncrement: pairs[x].MinSizeIncrement,
		}
		orderSizeLimitMap.Store(pairs[x].Symbol, tempValues)
	}

	pairs, err = b.GetMarketSummary("", false)
	if err != nil {
		return err
	}
	for x := range pairs {
		tempValues := OrderSizeLimit{
			MinOrderSize:     pairs[x].MinOrderSize,
			MaxOrderSize:     pairs[x].MaxOrderSize,
			MinSizeIncrement: pairs[x].MinSizeIncrement,
		}
		orderSizeLimitMap.Store(pairs[x].Symbol, tempValues)
	}
	return nil
}

// OrderSizeLimits looks up currency pair in orderSizeLimitMap and returns OrderSizeLimit
func OrderSizeLimits(pair string) (limits OrderSizeLimit, found bool) {
	resp, ok := orderSizeLimitMap.Load(pair)
	if !ok {
		return
	}
	val, ok := resp.(OrderSizeLimit)
	return val, ok
}
