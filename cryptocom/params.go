package cryptocom

import (
	"errors"
	"fmt"
	"github.com/openware/pkg/order"
	"strings"
	"time"
)

type KVParams map[string]interface{}

type TradeParams struct {
	Market   string `json:"instrument_name"`
	StartTS  int64  `json:"start_ts"`
	EndTS    int64  `json:"end_ts"`
	PageSize int    `json:"page_size"`
	Page     int    `json:"page"`
}

func (t *TradeParams) Validate() error {
	return tryOrError(func() error {
		if t.Market == "" {
			return nil
		}
		return validInstrument(t.Market)
	}, func() error {
		// timestamp validation
		if t.StartTS < 0 {
			return errors.New("start timestamp should be positive number")
		}
		if t.EndTS < 0 {
			return errors.New("end timestamp should be positive number")
		}
		if t.StartTS > 0 && t.EndTS > 0 {
			if t.StartTS > t.EndTS {
				return errors.New("start timestamp is ahead of end timestamp")
			}
			start := time.Unix(t.StartTS/1000, 0)
			end := time.Unix(t.EndTS/1000, 0)
			diff := end.Sub(start).Hours()
			if diff > 24 {
				return errors.New("max date range is 24 hours")
			}
		}
		return nil
	}, func() error {
		return validPagination(t.PageSize, t.Page)
	})
}

func (t *TradeParams) Encode() (KVParams, error) {
	pr := KVParams{}
	if t == nil {
		return pr, nil
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	if t.EndTS > 0 {
		pr["end_ts"] = t.EndTS
	}
	if t.Market != "" {
		pr["instrument_name"] = t.Market
	}
	if t.StartTS > 0 {
		pr["start_ts"] = t.StartTS
	}
	if t.Page > 0 {
		pr["page"] = t.Page
	}
	if t.PageSize > 0 {
		pr["page_size"] = t.PageSize
	}
	return pr, nil
}

type WithdrawParams struct {
	Amount     float64
	Address    string
	Currency   string
	WithdrawID string
	AddressTag string
}

func (w WithdrawParams) Validate() error {
	return tryOrError(func() error {
		return isValidCurrency(w.Currency)
	}, func() error {
		if w.Amount <= 0 {
			return errors.New("invalid withdraw amount")
		}
		return nil
	}, func() error {
		// TODO: maybe add crypto address validation? not really sure if it would be over-engineered
		if w.Address == "" {
			return errors.New("invalid withdraw address value")
		}
		return nil
	})
}
func (w WithdrawParams) Encode() (KVParams, error) {
	if err := w.Validate(); err != nil {
		return nil, err
	}
	pr := KVParams{
		"currency": w.Currency,
		"amount":   fmt.Sprintf("%v", w.Amount),
		"address":  w.Address,
	}
	if w.WithdrawID != "" {
		pr["client_wid"] = w.WithdrawID
	}
	if w.AddressTag != "" {
		pr["address_tag"] = w.AddressTag
	}
	return pr, nil
}

type WithdrawHistoryParam struct {
	Currency string
	StartTS  int64
	EndTS    int64
	PageSize int
	Page     int
	Status   WithdrawStatus
}

func (w *WithdrawHistoryParam) Validate() error {
	return tryOrError(func() error {
		if w.Currency == "" {
			return nil
		}
		return isValidCurrency(w.Currency)
	}, func() error {
		return validPagination(w.PageSize, w.Page)
	}, func() error {
		if w.StartTS < 0 {
			return errors.New("start timestamp should be positive number")
		}
		if w.EndTS < 0 {
			return errors.New("end timestamp should be positive number")
		}
		if w.StartTS > 0 && w.EndTS > 0 && w.StartTS > w.EndTS {
			return errors.New("start timestamp is ahead of end timestamp")
		}
		return nil
	}, func() error {
		if w.Status == 0 {
			w.Status = WithdrawNone
		}
		switch w.Status {
		case
			WithdrawNone,
			WithdrawPending,
			WithdrawProcessing,
			WithdrawRejected,
			WithdrawPaymentInProgress,
			WithdrawPaymentFailed,
			WithdrawCompleted,
			WithdrawCancelled:
			return nil
		default:
			return errors.New("invalid status value")
		}
	})
}
func (w *WithdrawHistoryParam) Encode() (KVParams, error) {
	pr := KVParams{}
	if w == nil {
		return pr, nil
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	if w.Currency != "" {
		pr["currency"] = w.Currency
	}
	if w.StartTS > 0 {
		pr["start_ts"] = w.StartTS
	}
	if w.EndTS > 0 {
		pr["end_ts"] = w.EndTS
	}
	if w.PageSize > 0 {
		pr["page_size"] = w.PageSize
	}
	if w.Page > 0 {
		pr["page"] = w.Page
	}
	if w.Status != WithdrawNone {
		pr["status"] = fmt.Sprintf("%d", w.Status)
	}

	return pr, nil
}

type DepositHistoryParam struct {
	Currency string
	StartTS  int64
	EndTS    int64
	PageSize int
	Page     int
	Status   DepositStatus
}

func (w *DepositHistoryParam) Validate() error {
	return tryOrError(func() error {
		if w.Currency == "" {
			return nil
		}
		return isValidCurrency(w.Currency)
	}, func() error {
		return validPagination(w.PageSize, w.Page)
	}, func() error {
		if w.StartTS < 0 {
			return errors.New("start timestamp should be positive number")
		}
		if w.EndTS < 0 {
			return errors.New("end timestamp should be positive number")
		}
		if w.StartTS > 0 && w.EndTS > 0 && w.StartTS > w.EndTS {
			return errors.New("start timestamp is ahead of end timestamp")
		}
		return nil
	}, func() error {
		if w.Status == 0 {
			w.Status = WithdrawNone
		}
		switch w.Status {
		case
			DepositNone,
			DepositArrived,
			DepositPending,
			DepositFailed,
			DepositNotArrived:
			return nil
		default:
			return errors.New("invalid status value")
		}
	})
}
func (w *DepositHistoryParam) Encode() (KVParams, error) {
	pr := KVParams{}
	if w == nil {
		return pr, nil
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	if w.Currency != "" {
		pr["currency"] = w.Currency
	}
	if w.StartTS > 0 {
		pr["start_ts"] = w.StartTS
	}
	if w.EndTS > 0 {
		pr["end_ts"] = w.EndTS
	}
	if w.PageSize > 0 {
		pr["page_size"] = w.PageSize
	}
	if w.Page > 0 {
		pr["page"] = w.Page
	}
	if w.Status != DepositNone {
		pr["status"] = fmt.Sprintf("%d", w.Status)
	}

	return pr, nil
}

type OpenOrderParam struct {
	Market   string
	PageSize int
	Page     int
}

func (o *OpenOrderParam) Validate() error {
	return tryOrError(func() error {
		if o.Market == "" {
			return nil
		}
		return validInstrument(o.Market)
	}, func() error {
		return validPagination(o.PageSize, o.Page)
	})
}
func (o *OpenOrderParam) Encode() (pr KVParams, err error) {
	pr = KVParams{}
	if o == nil {
		return
	}
	if err = o.Validate(); err != nil {
		return
	}
	if o.Market != "" {
		pr["instrument_name"] = o.Market
	}
	if o.Page > 0 {
		pr["page"] = o.Page
	}
	if o.PageSize > 0 {
		pr["page_size"] = o.PageSize
	}
	return
}

type CreateOrderParam struct {
	Market        string     `json:"instrument_name"`
	Side          order.Side `json:"side"`
	OrderType     order.Type `json:"order_type"`
	Price         float64    `json:"price"`
	Quantity      float64    `json:"quantity"`
	Notional      float64    `json:"notional"`
	ClientOrderID string     `json:"client_order_id"`
	TimeInForce   order.Type `json:"time_in_force"`
	ExecInst      order.Type `json:"exec_inst"`
	TriggerPrice  float64    `json:"trigger_price"`
}

func (c CreateOrderParam) Validate() error {
	return tryOrError(func() error {
		return validInstrument(c.Market)
	}, func() (err error) {
		if c.Side != order.Buy && c.Side != order.Sell {
			err = errors.New("invalid order side")
		}
		return
	}, func() error {
		switch c.OrderType {
		case order.Limit, order.StopLimit, order.Market, TakeProfitLimit, StopLoss, order.TakeProfit:
			return nil
		default:
			return errors.New("invalid order type")
		}
	}, func() (err error) {
		if c.OrderType == order.Limit {
			if c.Quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if c.Price <= 0 {
				err = errors.New("price required")
				return
			}
				if c.ExecInst != "" && c.ExecInst != PostOnly {
					err = fmt.Errorf("exec_inst value not allowed. either leave it empty or set it to %s", PostOnly)
					return
				}
				if c.TimeInForce != "" {
					switch c.TimeInForce {
					case
						GoodTillCancel,
						FillOrKill,
						order.ImmediateOrCancel:
						break
					default:
						err = fmt.Errorf("time_in_force value not allowed. either leave it empty or set it to %s, %s, or %s", GoodTillCancel, FillOrKill, order.ImmediateOrCancel)
						return
					}
				}
		}
		return
	}, func() (err error) {
		if c.OrderType == order.Market {
			if c.Side == order.Buy && (c.Notional <= 0) {
				err = errors.New("notional required")
				return
			}
			if c.Side == order.Sell && c.Quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
		}
		return
	}, func() (err error) {
		if c.OrderType == order.StopLimit || c.OrderType == TakeProfitLimit {
			if c.Price <= 0 {
				err = errors.New("price required")
				return
			}
			if c.Quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if c.TriggerPrice <= 0 {
				err = errors.New("trigger_price required")
				return
			}
		}
		return
	}, func() (err error) {
		if c.OrderType == StopLoss || c.OrderType == order.TakeProfit {
			if c.Side == order.Buy && c.Notional <= 0 {
				err = errors.New("notional required")
				return
			}
			if c.Side == order.Sell && c.Quantity <= 0 {
				err = errors.New("quantity required")
				return
			}
			if c.TriggerPrice <= 0 {
				err = errors.New("trigger_price required")
				return
			}
		}
		return
	})
}
func (c CreateOrderParam) Encode() (pr KVParams, err error) {
	if err = c.Validate(); err != nil {
		return
	}
	pr = KVParams{
		"instrument_name": c.Market,
		"side": c.Side.String(),
		"type": strings.ReplaceAll(c.OrderType.String(), " ", "-"),
	}
	if c.Price > 0 {
		pr["price"] = fmt.Sprintf("%v", c.Price)
	}
	if c.Quantity > 0 {
		pr["quantity"] = fmt.Sprintf("%v", c.Quantity)
	}
	if c.Notional > 0 {
		pr["notional"] = fmt.Sprintf("%v", c.Notional)
	}
	if c.TriggerPrice > 0 {
		pr["trigger_price"] = fmt.Sprintf("%v", c.TriggerPrice)
	}
	// set params only if order type is order.Limit
	if c.TimeInForce != "" && c.OrderType == order.Limit {
		pr["time_in_force"] = c.TimeInForce
	}
	if c.ExecInst != "" && c.OrderType == order.Limit {
		pr["exec_inst"] = c.ExecInst
	}
	if c.ClientOrderID != "" {
		pr["client_oid"] = c.ClientOrderID
	}
	return
}