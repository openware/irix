package cryptocom

// basic configuration
const (
	host = "api.crypto.com"
	streamHost = "stream.crypto.com"
	sandboxHost = "uat-api.3ona.com"
	sandboxStreamHost = "uat-stream.3ona.com"
	apiVersion = "v2"
)

// available methods
const (
	// available in both ws and rest
	publicGetInstruments = "public/get-instruments"
	privateCreateWithdrawal = "private/create-withdrawal"
	privateGetWithdrawalHistory = "private/get-withdrawal-history"
	privateGetAccountSummary = "private/get-account-summary"
	privateCreateOrder = "private/create-order"
	privateCancelOrder = "private/cancel-order"
	privateCancelAllOrders = "private/cancel-all-orders"
	privateGetOrderHistory = "private/get-order-history"
	privateGetOpenOrders = "private/get-open-orders"
	privateGetOrderDetail = "private/get-order-detail"
	privateGetTrades = "private/get-trades"


	// only in rest
	publicGetBook = "public/get-book"
	publicGetCandlestick = "public/get-candlestick"
	publicGetTicker = "public/get-ticker"
	publicGetTrades = "public/get-trades"
	privateGetDepositHistory = "private/get-deposit-history"
	privateGetDepositAddress = "private/get-deposit-address"

	// only in ws
	publicAuth = "public/auth"
	publicRespondHeartbeat = "public/respond-heartbeat"
	privateSetCancelOnDisconnect = "private/set-cancel-on-disconnect"
	privateGetCancelOnDisconnect = "private/get-cancel-on-disconnect"
	subscribe = "subscribe"
	// ws endpoints
	userEndpoint   = "user"
	marketEndpoint = "market"
)
