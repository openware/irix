package banking

import (
	"fmt"
	"strings"

	"github.com/openware/pkg/common"
	"github.com/openware/pkg/currency"
)

// GetBankAccountByID Returns a bank account based on its ID
func GetBankAccountByID(id string) (*Account, error) {
	m.Lock()
	defer m.Unlock()

	for x := range Accounts {
		if strings.EqualFold(Accounts[x].ID, id) {
			return &Accounts[x], nil
		}
	}
	return nil, fmt.Errorf(ErrBankAccountNotFound, id)
}

// ExchangeSupported Checks if exchange is supported by bank account
func (b *Account) ExchangeSupported(exchange string) bool {
	exchList := strings.Split(b.SupportedExchanges, ",")
	return common.StringDataCompareInsensitive(exchList, exchange)
}

// Validate validates bank account settings
func (b *Account) Validate() error {
	if b.BankName == "" ||
		b.BankAddress == "" ||
		b.BankPostalCode == "" ||
		b.BankPostalCity == "" ||
		b.BankCountry == "" ||
		b.AccountName == "" ||
		b.SupportedCurrencies == "" {
		return fmt.Errorf(
			"banking details for %s is enabled but variables not set correctly",
			b.BankName)
	}

	if b.SupportedExchanges == "" {
		b.SupportedExchanges = "ALL"
	}

	if strings.Contains(strings.ToUpper(
		b.SupportedCurrencies),
		currency.AUD.String()) {
		if b.BSBNumber == "" {
			return fmt.Errorf(
				"banking details for %s is enabled but BSB/SWIFT values not set",
				b.BankName)
		}
	} else {
		if b.IBAN == "" && b.SWIFTCode == "" {
			return fmt.Errorf(
				"banking details for %s is enabled but SWIFT/IBAN values not set",
				b.BankName)
		}
	}
	return nil
}

// ValidateForWithdrawal confirms bank account meets minimum requirements to submit
// a withdrawal request
func (b *Account) ValidateForWithdrawal(exchange string, cur currency.Code) (err []string) {
	if !b.Enabled {
		err = append(err, ErrBankAccountDisabled)
	}
	if !b.ExchangeSupported(exchange) {
		err = append(err, "Exchange "+exchange+" not supported by bank account")
	}

	if b.AccountNumber == "" {
		err = append(err, ErrAccountCannotBeEmpty)
	}

	if !common.StringDataCompareInsensitive(strings.Split(b.SupportedCurrencies, ","), cur.String()) {
		err = append(err, ErrCurrencyNotSupportedByAccount)
	}

	if cur.Upper() == currency.AUD {
		if b.BSBNumber == "" {
			err = append(err, ErrBSBRequiredforAUD)
		}
	} else {
		if b.IBAN == "" && b.SWIFTCode == "" {
			err = append(err, ErrIBANSwiftNotSet)
		}
	}
	return
}
