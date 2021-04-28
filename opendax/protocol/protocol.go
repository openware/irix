package protocol

const (
	EventBalanceUpdate      = "bu"
	EventOrderCreate        = "on"
	EventOrderCancel        = "oc"
	EventOrderUpdate        = "ou"
	EventOrderReject        = "or"
	EventTrade              = "tr"
	EventOrderBookIncrement = "obi"
	EventOrderBookSnapshot  = "obs"
	EventRawBookIncrement   = "rbi"
	EventRawBookSnapshot    = "rbs"

	MethodSubscribe       = "subscribe"
	MethodListOrders      = "list_orders"
	MethodGetOrders       = "get_orders"
	MethodGetOrderTrades  = "get_order_trades"
	MethodCreateOrder     = "create_order"
	MethodCreateOrderBulk = "create_bulk"
	MethodCancelOrder     = "cancel_order"
	MethodCancelOrderBulk = "cancel_bulk"

	OrderSideSell = "sell"
	OrderSideBuy  = "buy"

	OrderStatePending = "p"
	OrderStateWait    = "w"
	OrderStateDone    = "d"
	OrderStateReject  = "r"
	OrderStateCancel  = "c"

	OrderTypeLimit      = "l"
	OrderTypeMarket     = "m"
	OrderTypePostOnly   = "p"
	OrderTypeFillOrKill = "f"

	TopicBalances = "balances"
	TopicOrder    = "order"
	TopicTrade    = "trade"
)
