package irix

import (
	"sync"
	"time"

	"github.com/openware/irix/config"
	"github.com/openware/irix/protocol"
	"github.com/openware/irix/stream"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/kline"
	"github.com/openware/pkg/order"
	"github.com/openware/pkg/request"
)

// Endpoint authentication types
const (
	RestAuthentication      uint8 = 0
	WebsocketAuthentication uint8 = 1
	// Repeated exchange strings
	// FeeType custom type for calculating fees based on method
	WireTransfer InternationalBankTransactionType = iota
	PerfectMoney
	Neteller
	AdvCash
	Payeer
	Skrill
	Simplex
	SEPA
	Swift
	RapidTransfer
	MisterTangoSEPA
	Qiwi
	VisaMastercard
	WebMoney
	Capitalist
	WesternUnion
	MoneyGram
	Contact
	// Const declarations for fee types
	BankFee FeeType = iota
	InternationalBankDepositFee
	InternationalBankWithdrawalFee
	CryptocurrencyTradeFee
	CyptocurrencyDepositFee
	CryptocurrencyWithdrawalFee
	OfflineTradeFee
	// Definitions for each type of withdrawal method for a given exchange
	NoAPIWithdrawalMethods                  uint32 = 0
	NoAPIWithdrawalMethodsText              string = "NONE, WEBSITE ONLY"
	AutoWithdrawCrypto                      uint32 = (1 << 0)
	AutoWithdrawCryptoWithAPIPermission     uint32 = (1 << 1)
	AutoWithdrawCryptoWithSetup             uint32 = (1 << 2)
	AutoWithdrawCryptoText                  string = "AUTO WITHDRAW CRYPTO"
	AutoWithdrawCryptoWithAPIPermissionText string = "AUTO WITHDRAW CRYPTO WITH API PERMISSION"
	AutoWithdrawCryptoWithSetupText         string = "AUTO WITHDRAW CRYPTO WITH SETUP"
	WithdrawCryptoWith2FA                   uint32 = (1 << 3)
	WithdrawCryptoWithSMS                   uint32 = (1 << 4)
	WithdrawCryptoWithEmail                 uint32 = (1 << 5)
	WithdrawCryptoWithWebsiteApproval       uint32 = (1 << 6)
	WithdrawCryptoWithAPIPermission         uint32 = (1 << 7)
	WithdrawCryptoWith2FAText               string = "WITHDRAW CRYPTO WITH 2FA"
	WithdrawCryptoWithSMSText               string = "WITHDRAW CRYPTO WITH SMS"
	WithdrawCryptoWithEmailText             string = "WITHDRAW CRYPTO WITH EMAIL"
	WithdrawCryptoWithWebsiteApprovalText   string = "WITHDRAW CRYPTO WITH WEBSITE APPROVAL"
	WithdrawCryptoWithAPIPermissionText     string = "WITHDRAW CRYPTO WITH API PERMISSION"
	AutoWithdrawFiat                        uint32 = (1 << 8)
	AutoWithdrawFiatWithAPIPermission       uint32 = (1 << 9)
	AutoWithdrawFiatWithSetup               uint32 = (1 << 10)
	AutoWithdrawFiatText                    string = "AUTO WITHDRAW FIAT"
	AutoWithdrawFiatWithAPIPermissionText   string = "AUTO WITHDRAW FIAT WITH API PERMISSION"
	AutoWithdrawFiatWithSetupText           string = "AUTO WITHDRAW FIAT WITH SETUP"
	WithdrawFiatWith2FA                     uint32 = (1 << 11)
	WithdrawFiatWithSMS                     uint32 = (1 << 12)
	WithdrawFiatWithEmail                   uint32 = (1 << 13)
	WithdrawFiatWithWebsiteApproval         uint32 = (1 << 14)
	WithdrawFiatWithAPIPermission           uint32 = (1 << 15)
	WithdrawFiatWith2FAText                 string = "WITHDRAW FIAT WITH 2FA"
	WithdrawFiatWithSMSText                 string = "WITHDRAW FIAT WITH SMS"
	WithdrawFiatWithEmailText               string = "WITHDRAW FIAT WITH EMAIL"
	WithdrawFiatWithWebsiteApprovalText     string = "WITHDRAW FIAT WITH WEBSITE APPROVAL"
	WithdrawFiatWithAPIPermissionText       string = "WITHDRAW FIAT WITH API PERMISSION"
	WithdrawCryptoViaWebsiteOnly            uint32 = (1 << 16)
	WithdrawFiatViaWebsiteOnly              uint32 = (1 << 17)
	WithdrawCryptoViaWebsiteOnlyText        string = "WITHDRAW CRYPTO VIA WEBSITE ONLY"
	WithdrawFiatViaWebsiteOnlyText          string = "WITHDRAW FIAT VIA WEBSITE ONLY"
	NoFiatWithdrawals                       uint32 = (1 << 18)
	NoFiatWithdrawalsText                   string = "NO FIAT WITHDRAWAL"
	UnknownWithdrawalTypeText               string = "UNKNOWN"
)

// FeeType is the type for holding a custom fee type (International withdrawal fee)
type FeeType uint8

// InternationalBankTransactionType custom type for calculating fees based on fiat transaction types
type InternationalBankTransactionType uint8

// FeeBuilder is the type which holds all parameters required to calculate a fee
// for an exchange
type FeeBuilder struct {
	FeeType FeeType
	// Used for calculating crypto trading fees, deposits & withdrawals
	Pair    currency.Pair
	IsMaker bool
	// Fiat currency used for bank deposits & withdrawals
	FiatCurrency        currency.Code
	BankTransactionType InternationalBankTransactionType
	// Used to multiply for fee calculations
	PurchasePrice float64
	Amount        float64
}

// FundHistory holds exchange funding history data
type FundHistory struct {
	ExchangeName      string
	Status            string
	TransferID        string
	Description       string
	Timestamp         time.Time
	Currency          string
	Amount            float64
	Fee               float64
	TransferType      string
	CryptoToAddress   string
	CryptoFromAddress string
	CryptoTxID        string
	BankTo            string
	BankFrom          string
}

// WithdrawalHistory holds exchange Withdrawal history data
type WithdrawalHistory struct {
	Status          string
	TransferID      string
	Description     string
	Timestamp       time.Time
	Currency        string
	Amount          float64
	Fee             float64
	TransferType    string
	CryptoToAddress string
	CryptoTxID      string
	BankTo          string
}

// Features stores the supported and enabled features
// for the exchange
type Features struct {
	Supports FeaturesSupported
	Enabled  FeaturesEnabled
}

// FeaturesEnabled stores the exchange enabled features
type FeaturesEnabled struct {
	AutoPairUpdates bool
	Kline           kline.ExchangeCapabilitiesEnabled
	SaveTradeData   bool
}

// FeaturesSupported stores the exchanges supported features
type FeaturesSupported struct {
	REST                  bool
	RESTCapabilities      protocol.Features
	Websocket             bool
	WebsocketCapabilities protocol.Features
	WithdrawPermissions   uint32
	Kline                 kline.ExchangeCapabilitiesSupported
}

// Endpoints stores running url endpoints for exchanges
type Endpoints struct {
	Exchange string
	defaults map[string]string
	sync.RWMutex
}

// API stores the exchange API settings
type API struct {
	AuthenticatedSupport          bool
	AuthenticatedWebsocketSupport bool
	PEMKeySupport                 bool

	Endpoints *Endpoints

	Credentials struct {
		Key      string
		Secret   string
		ClientID string
		PEMKey   string
	}

	CredentialsValidator struct {
		// For Huobi (optional)
		RequiresPEM bool

		RequiresKey                bool
		RequiresSecret             bool
		RequiresClientID           bool
		RequiresBase64DecodeSecret bool
	}
}

// Base stores the individual exchange information
type Base struct {
	Name                          string
	Enabled                       bool
	Verbose                       bool
	LoadedByConfig                bool
	SkipAuthCheck                 bool
	API                           API
	BaseCurrencies                currency.Currencies
	CurrencyPairs                 currency.PairsManager
	Features                      Features
	HTTPTimeout                   time.Duration
	HTTPUserAgent                 string
	HTTPRecording                 bool
	HTTPDebugging                 bool
	WebsocketResponseCheckTimeout time.Duration
	WebsocketResponseMaxLimit     time.Duration
	WebsocketOrderbookBufferLimit int64
	Websocket                     *stream.Websocket
	*request.Requester
	Config        *config.ExchangeConfig
	settingsMutex sync.RWMutex
	// CanVerifyOrderbook determines if the orderbook verification can be bypassed,
	// increasing potential update speed but decreasing confidence in orderbook
	// integrity.
	CanVerifyOrderbook bool
	order.ExecutionLimits

	AssetWebsocketSupport
}

// url lookup consts
const (
	RestSpot URL = iota
	RestSpotSupplementary
	RestUSDTMargined
	RestCoinMargined
	RestFutures
	RestSwap
	RestSandbox
	WebsocketSpot
	WebsocketSpotSupplementary
	ChainAnalysis
	EdgeCase1
	EdgeCase2
	EdgeCase3
)

var keyURLs = []URL{RestSpot,
	RestSpotSupplementary,
	RestUSDTMargined,
	RestCoinMargined,
	RestFutures,
	RestSwap,
	RestSandbox,
	WebsocketSpot,
	WebsocketSpotSupplementary,
	ChainAnalysis,
	EdgeCase1,
	EdgeCase2,
	EdgeCase3}

// URL stores uint conversions
type URL uint16

// AssetWebsocketSupport defines the availability of websocket functionality to
// the specific asset type. TODO: Deprecate as this is a temp item to address
// certain limitations quickly.
type AssetWebsocketSupport struct {
	unsupported map[asset.Item]bool
	m           sync.RWMutex
}
