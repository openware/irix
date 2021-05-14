package cryptocom

import (
	"errors"
	"time"
)

type KVParams map[string]interface{}
type TradeParams struct {
	Market   string
	StartTS  int64
	EndTS    int64
	PageSize int
	Page     int
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
	if t == nil {
		return KVParams{}, nil
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}

	return KVParams{
		"end_ts":          t.EndTS,
		"instrument_name": t.Market,
		"start_ts":        t.StartTS,
		"page":            t.Page,
		"page_size":       t.PageSize,
	}, nil
}

type WithdrawParams struct {
	Amount float64
	Address string
	Currency string
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
		"amount": w.Amount,
		"address": w.Address,
	}
	if w.WithdrawID != "" {
		pr["client_wid"] = w.WithdrawID
	}
	if w.AddressTag != "" {
		pr["address_tag"] = w.AddressTag
	}
	return pr, nil
}
