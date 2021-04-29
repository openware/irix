package config

import (
	"time"

	"github.com/openware/irix/portfolio/banking"
	"github.com/openware/irix/protocol"
	"github.com/openware/pkg/currency"
)

type Config struct {
	Name      string           `json:"name"`
	Exchanges []ExchangeConfig `json:"exchanges"`
}

type ExchangeConfig struct {
	Name                          string                 `json:"name"`
	Enabled                       bool                   `json:"enabled"`
	Verbose                       bool                   `json:"verbose"`
	UseSandbox                    bool                   `json:"useSandbox,omitempty"`
	HTTPTimeout                   time.Duration          `json:"httpTimeout"`
	HTTPUserAgent                 string                 `json:"httpUserAgent,omitempty"`
	HTTPDebugging                 bool                   `json:"httpDebugging,omitempty"`
	WebsocketResponseCheckTimeout time.Duration          `json:"websocketResponseCheckTimeout"`
	WebsocketResponseMaxLimit     time.Duration          `json:"websocketResponseMaxLimit"`
	WebsocketTrafficTimeout       time.Duration          `json:"websocketTrafficTimeout"`
	ProxyAddress                  string                 `json:"proxyAddress,omitempty"`
	BaseCurrencies                currency.Currencies    `json:"baseCurrencies"`
	CurrencyPairs                 *currency.PairsManager `json:"currencyPairs"`
	API                           APIConfig              `json:"api"`
	Features                      *FeaturesConfig        `json:"features"`
	BankAccounts                  []banking.Account      `json:"bankAccounts,omitempty"`
	OrderbookConfig               `json:"orderbook"`

	// Deprecated settings which will be removed in a future update
	AvailablePairs                   *currency.Pairs      `json:"availablePairs,omitempty"`
	EnabledPairs                     *currency.Pairs      `json:"enabledPairs,omitempty"`
	AssetTypes                       *string              `json:"assetTypes,omitempty"`
	PairsLastUpdated                 *int64               `json:"pairsLastUpdated,omitempty"`
	ConfigCurrencyPairFormat         *currency.PairFormat `json:"configCurrencyPairFormat,omitempty"`
	RequestCurrencyPairFormat        *currency.PairFormat `json:"requestCurrencyPairFormat,omitempty"`
	AuthenticatedAPISupport          *bool                `json:"authenticatedApiSupport,omitempty"`
	AuthenticatedWebsocketAPISupport *bool                `json:"authenticatedWebsocketApiSupport,omitempty"`
	APIKey                           *string              `json:"apiKey,omitempty"`
	APISecret                        *string              `json:"apiSecret,omitempty"`
	APIAuthPEMKeySupport             *bool                `json:"apiAuthPemKeySupport,omitempty"`
	APIAuthPEMKey                    *string              `json:"apiAuthPemKey,omitempty"`
	APIURL                           *string              `json:"apiUrl,omitempty"`
	APIURLSecondary                  *string              `json:"apiUrlSecondary,omitempty"`
	ClientID                         *string              `json:"clientId,omitempty"`
	SupportsAutoPairUpdates          *bool                `json:"supportsAutoPairUpdates,omitempty"`
	Websocket                        *bool                `json:"websocket,omitempty"`
	WebsocketURL                     *string              `json:"websocketUrl,omitempty"`
}

// FeaturesSupportedConfig stores the exchanges supported features
type FeaturesSupportedConfig struct {
	REST                  bool              `json:"restAPI"`
	RESTCapabilities      protocol.Features `json:"restCapabilities,omitempty"`
	Websocket             bool              `json:"websocketAPI"`
	WebsocketCapabilities protocol.Features `json:"websocketCapabilities,omitempty"`
}

// FeaturesEnabledConfig stores the exchanges enabled features
type FeaturesEnabledConfig struct {
	AutoPairUpdates bool `json:"autoPairUpdates"`
	Websocket       bool `json:"websocketAPI"`
	SaveTradeData   bool `json:"saveTradeData"`
}

// FeaturesConfig stores the exchanges supported and enabled features
type FeaturesConfig struct {
	Supports FeaturesSupportedConfig `json:"supports"`
	Enabled  FeaturesEnabledConfig   `json:"enabled"`
}

// APIEndpointsConfig stores the API endpoint addresses
type APIEndpointsConfig struct {
	URL          string `json:"url"`
	URLSecondary string `json:"urlSecondary"`
	WebsocketURL string `json:"websocketURL"`
}

// APICredentialsConfig stores the API credentials
type APICredentialsConfig struct {
	Key       string `json:"key,omitempty"`
	Secret    string `json:"secret,omitempty"`
	ClientID  string `json:"clientID,omitempty"`
	PEMKey    string `json:"pemKey,omitempty"`
	OTPSecret string `json:"otpSecret,omitempty"`
}

// APICredentialsValidatorConfig stores the API credentials validator settings
type APICredentialsValidatorConfig struct {
	// For Huobi (optional)
	RequiresPEM bool `json:"requiresPEM,omitempty"`

	RequiresKey                bool `json:"requiresKey,omitempty"`
	RequiresSecret             bool `json:"requiresSecret,omitempty"`
	RequiresClientID           bool `json:"requiresClientID,omitempty"`
	RequiresBase64DecodeSecret bool `json:"requiresBase64DecodeSecret,omitempty"`
}

// APIConfig stores the exchange API config
type APIConfig struct {
	AuthenticatedSupport          bool `json:"authenticatedSupport"`
	AuthenticatedWebsocketSupport bool `json:"authenticatedWebsocketApiSupport"`
	PEMKeySupport                 bool `json:"pemKeySupport,omitempty"`

	Credentials          APICredentialsConfig           `json:"credentials"`
	CredentialsValidator *APICredentialsValidatorConfig `json:"credentialsValidator,omitempty"`
	OldEndPoints         *APIEndpointsConfig            `json:"endpoints,omitempty"`
	Endpoints            map[string]string              `json:"urlEndpoints"`
}

// OrderbookConfig stores the orderbook configuration variables
type OrderbookConfig struct {
	VerificationBypass     bool `json:"verificationBypass"`
	WebsocketBufferLimit   int  `json:"websocketBufferLimit"`
	WebsocketBufferEnabled bool `json:"websocketBufferEnabled"`
}
