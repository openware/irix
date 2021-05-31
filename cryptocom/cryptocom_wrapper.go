package cryptocom

import (
	exchange "github.com/openware/irix"
	"github.com/openware/irix/config"
	"github.com/openware/irix/portfolio/withdraw"
	"github.com/openware/irix/protocol"
	"github.com/openware/irix/stream"
	"github.com/openware/irix/ticker"
	"github.com/openware/pkg/account"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/kline"
	"github.com/openware/pkg/log"
	"github.com/openware/pkg/order"
	"github.com/openware/pkg/orderbook"
	"github.com/openware/pkg/request"
	"github.com/openware/pkg/trade"
	"math/rand"
	"sync"
	"time"
)

func (c *Client) Setup(exch *config.ExchangeConfig) error {
	c.SetDefaults()
	return nil
}

func (c *Client) Start(wg *sync.WaitGroup) {
	c.Connect()
}

func (c *Client) SetDefaults() {
	c.Name = exchangeName
	c.Enabled = true
	c.Verbose = true
	c.API.CredentialsValidator.RequiresKey = true
	c.API.CredentialsValidator.RequiresClientID = true

	requestFmt := &currency.PairFormat{Uppercase: true}
	configFmt := &currency.PairFormat{Uppercase: true, Delimiter: currency.DashDelimiter}
	err := c.SetGlobalPairsManager(requestFmt, configFmt, asset.Spot)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	c.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:    true,
				TradeFetching:     true,
				OrderbookFetching: true,
				AutoPairUpdates:   true,
				AccountInfo:       true,
				GetOrders:         true,
				CancelOrders:      true,
				CancelOrder:       true,
				SubmitOrder:       true,
				SubmitOrders:      true,
				UserTradeHistory:  true,
				TradeFee:          true,
				FiatDepositFee:    true,
				FiatWithdrawalFee: true,
			},
			WebsocketCapabilities: protocol.Features{
				AccountBalance:         true,
				GetOrders:              true,
				CancelOrders:           true,
				CancelOrder:            true,
				SubmitOrder:            true,
				SubmitOrders:           true,
				UserTradeHistory:       true,
				TickerFetching:         true,
				TradeFetching:          true,
				OrderbookFetching:      true,
				AccountInfo:            true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				MessageCorrelation:     true,
			},
			WithdrawPermissions: exchange.WithdrawCryptoViaWebsiteOnly |
				exchange.WithdrawFiatViaWebsiteOnly,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	c.Requester = request.New(c.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	c.rest.(*httpClient).client = c.Requester.HTTPClient
	c.Websocket = stream.New()
	//c.publicConn.Transport = c.Websocket.AuthConn
	c.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	c.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	c.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
	rand.Seed(time.Now().UnixNano())
}

func (c *Client) GetName() string {
	return exchangeName
}

func (c *Client) IsEnabled() bool {
	return c.Enabled
}

func (c *Client) SetEnabled(b bool) {
	c.Enabled = b
}

func (c *Client) ValidateCredentials(a asset.Item) error {
	_, err := c.RestGetAccountSummary(a.String())
	return err
}

func (c *Client) FetchTicker(p currency.Pair, a asset.Item) (*ticker.Price, error) {
	if c.IsWebsocketEnabled() {
		c.Websocket.Conn.SendMessageReturnResponse(c.getTicker(p.String()))
	} else {
		res, err := c.RestGetTicker(p.String())
	}

}

func (c *Client) UpdateTicker(p currency.Pair, a asset.Item) (*ticker.Price, error) {
	panic("implement me")
}

func (c *Client) FetchOrderbook(p currency.Pair, a asset.Item) (*orderbook.Base, error) {
	panic("implement me")
}

func (c *Client) UpdateOrderbook(p currency.Pair, a asset.Item) (*orderbook.Base, error) {
	panic("implement me")
}

func (c *Client) FetchTradablePairs(a asset.Item) ([]string, error) {
	panic("implement me")
}

func (c *Client) UpdateTradablePairs(forceUpdate bool) error {
	panic("implement me")
}

func (c *Client) GetEnabledPairs(a asset.Item) (currency.Pairs, error) {
	panic("implement me")
}

func (c *Client) GetAvailablePairs(a asset.Item) (currency.Pairs, error) {
	panic("implement me")
}

func (c *Client) FetchAccountInfo(a asset.Item) (account.Holdings, error) {
	panic("implement me")
}

func (c *Client) UpdateAccountInfo(a asset.Item) (account.Holdings, error) {
	panic("implement me")
}

func (c *Client) GetAuthenticatedAPISupport(endpoint uint8) bool {
	panic("implement me")
}

func (c *Client) SetPairs(pairs currency.Pairs, a asset.Item, enabled bool) error {
	panic("implement me")
}

func (c *Client) GetAssetTypes() asset.Items {
	panic("implement me")
}

func (c *Client) GetRecentTrades(p currency.Pair, a asset.Item) ([]trade.Data, error) {
	panic("implement me")
}

func (c *Client) GetHistoricTrades(p currency.Pair, a asset.Item, startTime, endTime time.Time) ([]trade.Data, error) {
	panic("implement me")
}

func (c *Client) SupportsAutoPairUpdates() bool {
	panic("implement me")
}

func (c *Client) SupportsRESTTickerBatchUpdates() bool {
	panic("implement me")
}

func (c *Client) GetFeeByType(f *exchange.FeeBuilder) (float64, error) {
	panic("implement me")
}

func (c *Client) GetLastPairsUpdateTime() int64 {
	panic("implement me")
}

func (c *Client) GetWithdrawPermissions() uint32 {
	panic("implement me")
}

func (c *Client) FormatWithdrawPermissions() string {
	panic("implement me")
}

func (c *Client) SupportsWithdrawPermissions(permissions uint32) bool {
	panic("implement me")
}

func (c *Client) GetFundingHistory() ([]exchange.FundHistory, error) {
	panic("implement me")
}

func (c *Client) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	panic("implement me")
}

func (c *Client) ModifyOrder(action *order.Modify) (string, error) {
	panic("implement me")
}

func (c *Client) CancelOrder(o *order.Cancel) error {
	panic("implement me")
}

func (c *Client) CancelBatchOrders(o []order.Cancel) (order.CancelBatchResponse, error) {
	panic("implement me")
}

func (c *Client) CancelAllOrders(orders *order.Cancel) (order.CancelAllResponse, error) {
	panic("implement me")
}

func (c *Client) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	panic("implement me")
}

func (c *Client) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	panic("implement me")
}

func (c *Client) GetOrderHistory(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	panic("implement me")
}

func (c *Client) GetWithdrawalsHistory(code currency.Code) ([]exchange.WithdrawalHistory, error) {
	panic("implement me")
}

func (c *Client) GetActiveOrders(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	panic("implement me")
}

func (c *Client) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	panic("implement me")
}

func (c *Client) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	panic("implement me")
}

func (c *Client) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	panic("implement me")
}

func (c *Client) SetHTTPClientUserAgent(ua string) {
	panic("implement me")
}

func (c *Client) GetHTTPClientUserAgent() string {
	panic("implement me")
}

func (c *Client) SetClientProxyAddress(addr string) error {
	panic("implement me")
}

func (c *Client) SupportsREST() bool {
	panic("implement me")
}

func (c *Client) GetSubscriptions() ([]stream.ChannelSubscription, error) {
	panic("implement me")
}

func (c *Client) GetDefaultConfig() (*config.ExchangeConfig, error) {
	panic("implement me")
}

func (c *Client) GetHistoricCandles(p currency.Pair, a asset.Item, timeStart, timeEnd time.Time, interval kline.Interval) (kline.Item, error) {
	panic("implement me")
}

func (c *Client) GetHistoricCandlesExtended(p currency.Pair, a asset.Item, timeStart, timeEnd time.Time, interval kline.Interval) (kline.Item, error) {
	panic("implement me")
}

func (c *Client) AuthenticateWebsocket() error {
	return c.authenticate()
}
