package cryptocom

import (
	exchange "github.com/openware/irix"
	"github.com/openware/irix/config"
	"github.com/openware/irix/portfolio/withdraw"
	"github.com/openware/irix/stream"
	"github.com/openware/irix/ticker"
	"github.com/openware/pkg/account"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/kline"
	"github.com/openware/pkg/order"
	"github.com/openware/pkg/orderbook"
	"github.com/openware/pkg/trade"
	"sync"
	"time"
)

func (c *Client) Setup(exch *config.ExchangeConfig) error {
	panic("implement me")
}

func (c *Client) Start(wg *sync.WaitGroup) {
	panic("implement me")
}

func (c *Client) SetDefaults() {
	panic("implement me")
}

func (c *Client) GetName() string {
	panic("implement me")
}

func (c *Client) IsEnabled() bool {
	panic("implement me")
}

func (c *Client) SetEnabled(b bool) {
	panic("implement me")
}

func (c *Client) ValidateCredentials(a asset.Item) error {
	panic("implement me")
}

func (c *Client) FetchTicker(p currency.Pair, a asset.Item) (*ticker.Price, error) {
	panic("implement me")
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

func (c *Client) GetBase() *exchange.Base {
	panic("implement me")
}

func (c *Client) SupportsAsset(assetType asset.Item) bool {
	panic("implement me")
}

func (c *Client) GetHistoricCandles(p currency.Pair, a asset.Item, timeStart, timeEnd time.Time, interval kline.Interval) (kline.Item, error) {
	panic("implement me")
}

func (c *Client) GetHistoricCandlesExtended(p currency.Pair, a asset.Item, timeStart, timeEnd time.Time, interval kline.Interval) (kline.Item, error) {
	panic("implement me")
}

func (c *Client) DisableRateLimiter() error {
	panic("implement me")
}

func (c *Client) EnableRateLimiter() error {
	panic("implement me")
}

func (c *Client) GetWebsocket() (*stream.Websocket, error) {
	panic("implement me")
}

func (c *Client) IsWebsocketEnabled() bool {
	panic("implement me")
}

func (c *Client) SupportsWebsocket() bool {
	panic("implement me")
}

func (c *Client) SubscribeToWebsocketChannels(channels []stream.ChannelSubscription) error {
	panic("implement me")
}

func (c *Client) UnsubscribeToWebsocketChannels(channels []stream.ChannelSubscription) error {
	panic("implement me")
}

func (c *Client) IsAssetWebsocketSupported(aType asset.Item) bool {
	panic("implement me")
}

func (c *Client) FlushWebsocketChannels() error {
	panic("implement me")
}

func (c *Client) AuthenticateWebsocket() error {
	panic("implement me")
}

func (c *Client) GetOrderExecutionLimits(a asset.Item, cp currency.Pair) (*order.Limits, error) {
	panic("implement me")
}

func (c *Client) CheckOrderExecutionLimits(a asset.Item, cp currency.Pair, price, amount float64, orderType order.Type) error {
	panic("implement me")
}

func (c *Client) UpdateOrderExecutionLimits(a asset.Item) error {
	panic("implement me")
}
