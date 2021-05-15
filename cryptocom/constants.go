package cryptocom

import "github.com/openware/pkg/order"

// basic configuration
const (
	host              = "api.crypto.com"
	streamHost        = "stream.crypto.com"
	sandboxHost       = "uat-api.3ona.co"
	sandboxStreamHost = "uat-stream.3ona.co"
	apiVersion        = "v2"
)

// available methods
const (
	// available in both ws and rest
	publicGetInstruments        = "public/get-instruments"
	privateCreateWithdrawal     = "private/create-withdrawal"
	privateGetWithdrawalHistory = "private/get-withdrawal-history"
	privateGetAccountSummary    = "private/get-account-summary"
	privateCreateOrder          = "private/create-order"
	privateCancelOrder          = "private/cancel-order"
	privateCancelAllOrders      = "private/cancel-all-orders"
	privateGetOrderHistory      = "private/get-order-history"
	privateGetOpenOrders        = "private/get-open-orders"
	privateGetOrderDetail       = "private/get-order-detail"
	privateGetTrades            = "private/get-trades"

	// only in rest
	publicGetBook            = "public/get-book"
	publicGetCandlestick     = "public/get-candlestick"
	publicGetTicker          = "public/get-ticker"
	publicGetTrades          = "public/get-trades"
	privateGetDepositHistory = "private/get-deposit-history"
	privateGetDepositAddress = "private/get-deposit-address"

	// only in ws
	publicAuth                   = "public/auth"
	publicRespondHeartbeat       = "public/respond-heartbeat"
	privateSetCancelOnDisconnect = "private/set-cancel-on-disconnect"
	privateGetCancelOnDisconnect = "private/get-cancel-on-disconnect"
	subscribe                    = "subscribe"
	unsubscribe                  = "subscribe"
	// ws endpoints
	userEndpoint   = "user"
	marketEndpoint = "market"
)

type Interval int

const (
	Minute1 Interval = iota + 1
	Minute5
	Minute15
	Minute30
	Hour1
	Hour4
	Hour6
	Hour12
	Day
	Week
	Week2
	Month
)

func (c Interval) Encode() string {
	switch c {
	case Minute1:
		return "1m"
	case Minute5:
		return "5m"
	case Minute15:
		return "15m"
	case Minute30:
		return "30m"
	case Hour1:
		return "1h"
	case Hour4:
		return "4h"
	case Hour6:
		return "6h"
	case Hour12:
		return "12h"
	case Day:
		return "1d"
	case Week:
		return "1w"
	case Week2:
		return "2w"
	case Month:
		return "1M"
	default:
		return ""
	}
}

const (
	ScopeAccount    string = "ACCOUNT"
	ScopeConnection        = "CONNECTION"
)

// rate limit

// extend order types
const (
	TakeProfitLimit order.Type = "TAKE_PROFIT_LIMIT"
	StopLoss        order.Type = "STOP_LOSS"
	FillOrKill      order.Type = "FILL_OR_KILL"
	GoodTillCancel  order.Type = "GOOD_TILL_CANCEL"
	PostOnly        order.Type = "POST_ONLY"
)

type WithdrawStatus int

const (
	WithdrawPending WithdrawStatus = iota
	WithdrawProcessing
	WithdrawRejected
	WithdrawPaymentInProgress
	WithdrawPaymentFailed
	WithdrawCompleted
	WithdrawCancelled
	WithdrawNone = -1
)

type DepositStatus int

const (
	DepositNotArrived DepositStatus = iota
	DepositArrived
	DepositFailed
	DepositPending
	DepositNone = -1
)
