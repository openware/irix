package config

import (
	"strings"
	"testing"

	"github.com/openware/irix/portfolio/banking"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/common/convert"
	"github.com/openware/pkg/currency"
)

const (
	// Default number of enabled exchanges. Modify this whenever an exchange is
	// added or removed
	testFakeExchangeName = "Stampbit"
	testPair             = "BTC-USD"
	testString           = "test"
)

func TestLoadConfig(t *testing.T) {
	loadConfig := GetConfig()
	err := loadConfig.LoadConfig(TestFile, true)
	if err != nil {
		t.Error("TestLoadConfig " + err.Error())
	}

	err = loadConfig.LoadConfig("testy", true)
	if err == nil {
		t.Error("TestLoadConfig Expected error")
	}
}

// TestCheckExchangeConfigValues logic test
func TestCheckExchangeConfigValues(t *testing.T) {
	var cfg Config
	if err := cfg.CheckExchangeConfigValues(); err == nil {
		t.Error("nil exchanges should throw an err")
	}

	err := cfg.LoadConfig(TestFile, true)
	if err != nil {
		t.Fatal(err)
	}

	// Test our default test config and report any errors
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Fatal(err)
	}

	cfg.Exchanges[0].Name = "GDAX"
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}
	if cfg.Exchanges[0].Name != "CoinbasePro" {
		t.Error("exchange name should have been updated from GDAX to CoinbasePRo")
	}

	// Test API settings migration
	sptr := func(s string) *string { return &s }
	int64ptr := func(i int64) *int64 { return &i }

	cfg.Exchanges[0].APIKey = sptr("awesomeKey")
	cfg.Exchanges[0].APISecret = sptr("meowSecret")
	cfg.Exchanges[0].ClientID = sptr("clientIDerino")
	cfg.Exchanges[0].APIAuthPEMKey = sptr("-----BEGIN EC PRIVATE KEY-----\nASDF\n-----END EC PRIVATE KEY-----\n")
	cfg.Exchanges[0].APIAuthPEMKeySupport = convert.BoolPtr(true)
	cfg.Exchanges[0].AuthenticatedAPISupport = convert.BoolPtr(true)
	cfg.Exchanges[0].AuthenticatedWebsocketAPISupport = convert.BoolPtr(true)
	cfg.Exchanges[0].WebsocketURL = sptr("wss://1337")
	cfg.Exchanges[0].APIURL = sptr(APIURLNonDefaultMessage)
	cfg.Exchanges[0].APIURLSecondary = sptr(APIURLNonDefaultMessage)
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	// Ensure that all of our previous settings are migrated
	if cfg.Exchanges[0].API.Credentials.Key != "awesomeKey" ||
		cfg.Exchanges[0].API.Credentials.Secret != "meowSecret" ||
		cfg.Exchanges[0].API.Credentials.ClientID != "clientIDerino" ||
		!strings.Contains(cfg.Exchanges[0].API.Credentials.PEMKey, "ASDF") ||
		!cfg.Exchanges[0].API.PEMKeySupport ||
		!cfg.Exchanges[0].API.AuthenticatedSupport ||
		!cfg.Exchanges[0].API.AuthenticatedWebsocketSupport {
		t.Error("unexpected values")
	}

	if cfg.Exchanges[0].APIKey != nil ||
		cfg.Exchanges[0].APISecret != nil ||
		cfg.Exchanges[0].ClientID != nil ||
		cfg.Exchanges[0].APIAuthPEMKey != nil ||
		cfg.Exchanges[0].APIAuthPEMKeySupport != nil ||
		cfg.Exchanges[0].AuthenticatedAPISupport != nil ||
		cfg.Exchanges[0].AuthenticatedWebsocketAPISupport != nil ||
		cfg.Exchanges[0].WebsocketURL != nil ||
		cfg.Exchanges[0].APIURL != nil ||
		cfg.Exchanges[0].APIURLSecondary != nil {
		t.Error("unexpected values")
	}

	// Test feature and endpoint migrations migrations
	cfg.Exchanges[0].Features = nil
	cfg.Exchanges[0].SupportsAutoPairUpdates = convert.BoolPtr(true)
	cfg.Exchanges[0].Websocket = convert.BoolPtr(true)

	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if !cfg.Exchanges[0].Features.Enabled.AutoPairUpdates ||
		!cfg.Exchanges[0].Features.Enabled.Websocket ||
		!cfg.Exchanges[0].Features.Supports.RESTCapabilities.AutoPairUpdates {
		t.Error("unexpected values")
	}

	p1, err := currency.NewPairDelimiter(testPair, "-")
	if err != nil {
		t.Fatal(err)
	}

	// Test currency pair migration
	setupPairs := func(emptyAssets bool) {
		cfg.Exchanges[0].CurrencyPairs = nil
		p := currency.Pairs{
			p1,
		}
		cfg.Exchanges[0].PairsLastUpdated = int64ptr(1234567)

		if !emptyAssets {
			cfg.Exchanges[0].AssetTypes = sptr("spot")
		}

		cfg.Exchanges[0].AvailablePairs = &p
		cfg.Exchanges[0].EnabledPairs = &p
		cfg.Exchanges[0].ConfigCurrencyPairFormat = &currency.PairFormat{
			Uppercase: true,
			Delimiter: "-",
		}
		cfg.Exchanges[0].RequestCurrencyPairFormat = &currency.PairFormat{
			Uppercase: false,
			Delimiter: "~",
		}
	}

	setupPairs(false)
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	setupPairs(true)
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if cfg.Exchanges[0].CurrencyPairs.LastUpdated != 1234567 {
		t.Error("last updated has wrong value")
	}

	pFmt := cfg.Exchanges[0].CurrencyPairs.ConfigFormat
	if pFmt.Delimiter != "-" ||
		!pFmt.Uppercase {
		t.Error("unexpected config format values")
	}

	pFmt = cfg.Exchanges[0].CurrencyPairs.RequestFormat
	if pFmt.Delimiter != "~" ||
		pFmt.Uppercase {
		t.Error("unexpected request format values")
	}

	if !cfg.Exchanges[0].CurrencyPairs.GetAssetTypes().Contains(asset.Spot) ||
		!cfg.Exchanges[0].CurrencyPairs.UseGlobalFormat {
		t.Error("unexpected results")
	}

	pairs, err := cfg.Exchanges[0].CurrencyPairs.GetPairs(asset.Spot, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(pairs) == 0 || pairs.Join() != testPair {
		t.Error("pairs not set properly")
	}

	pairs, err = cfg.Exchanges[0].CurrencyPairs.GetPairs(asset.Spot, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(pairs) == 0 || pairs.Join() != testPair {
		t.Error("pairs not set properly")
	}

	// Ensure that all old settings are flushed
	if cfg.Exchanges[0].PairsLastUpdated != nil ||
		cfg.Exchanges[0].ConfigCurrencyPairFormat != nil ||
		cfg.Exchanges[0].RequestCurrencyPairFormat != nil ||
		cfg.Exchanges[0].AssetTypes != nil ||
		cfg.Exchanges[0].AvailablePairs != nil ||
		cfg.Exchanges[0].EnabledPairs != nil {
		t.Error("unexpected results")
	}

	// Test AutoPairUpdates
	cfg.Exchanges[0].Features.Supports.RESTCapabilities.AutoPairUpdates = false
	cfg.Exchanges[0].Features.Supports.WebsocketCapabilities.AutoPairUpdates = false
	cfg.Exchanges[0].CurrencyPairs.LastUpdated = 0
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	// Test websocket and HTTP timeout values
	cfg.Exchanges[0].WebsocketResponseMaxLimit = 0
	cfg.Exchanges[0].WebsocketResponseCheckTimeout = 0
	cfg.Exchanges[0].OrderbookConfig.WebsocketBufferLimit = 0
	cfg.Exchanges[0].WebsocketTrafficTimeout = 0
	cfg.Exchanges[0].HTTPTimeout = 0
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if cfg.Exchanges[0].WebsocketResponseMaxLimit == 0 {
		t.Errorf("expected exchange %s to have updated WebsocketResponseMaxLimit value",
			cfg.Exchanges[0].Name)
	}
	if cfg.Exchanges[0].OrderbookConfig.WebsocketBufferLimit == 0 {
		t.Errorf("expected exchange %s to have updated WebsocketOrderbookBufferLimit value",
			cfg.Exchanges[0].Name)
	}
	if cfg.Exchanges[0].WebsocketTrafficTimeout == 0 {
		t.Errorf("expected exchange %s to have updated WebsocketTrafficTimeout value",
			cfg.Exchanges[0].Name)
	}
	if cfg.Exchanges[0].HTTPTimeout == 0 {
		t.Errorf("expected exchange %s to have updated HTTPTimeout value",
			cfg.Exchanges[0].Name)
	}

	v := &APICredentialsValidatorConfig{
		RequiresKey:    true,
		RequiresSecret: true,
	}
	cfg.Exchanges[0].API.CredentialsValidator = v
	cfg.Exchanges[0].API.Credentials.Key = "Key"
	cfg.Exchanges[0].API.Credentials.Secret = "Secret"
	cfg.Exchanges[0].API.AuthenticatedSupport = true
	cfg.Exchanges[0].API.AuthenticatedWebsocketSupport = true
	cfg.CheckExchangeConfigValues()
	if cfg.Exchanges[0].API.AuthenticatedSupport ||
		cfg.Exchanges[0].API.AuthenticatedWebsocketSupport {
		t.Error("Expected authenticated endpoints to be false from invalid API keys")
	}

	v.RequiresKey = false
	v.RequiresClientID = true
	cfg.Exchanges[0].API.AuthenticatedSupport = true
	cfg.Exchanges[0].API.AuthenticatedWebsocketSupport = true
	cfg.Exchanges[0].API.Credentials.ClientID = DefaultAPIClientID
	cfg.Exchanges[0].API.Credentials.Secret = "TESTYTEST"
	cfg.CheckExchangeConfigValues()
	if cfg.Exchanges[0].API.AuthenticatedSupport ||
		cfg.Exchanges[0].API.AuthenticatedWebsocketSupport {
		t.Error("Expected AuthenticatedAPISupport to be false from invalid API keys")
	}

	v.RequiresKey = true
	cfg.Exchanges[0].API.AuthenticatedSupport = true
	cfg.Exchanges[0].API.AuthenticatedWebsocketSupport = true
	cfg.Exchanges[0].API.Credentials.Key = "meow"
	cfg.Exchanges[0].API.Credentials.Secret = "test123"
	cfg.Exchanges[0].API.Credentials.ClientID = "clientIDerino"
	cfg.CheckExchangeConfigValues()
	if !cfg.Exchanges[0].API.AuthenticatedSupport ||
		!cfg.Exchanges[0].API.AuthenticatedWebsocketSupport {
		t.Error("Expected AuthenticatedAPISupport and AuthenticatedWebsocketAPISupport to be false from invalid API keys")
	}

	// Make a sneaky copy for bank account testing
	cpy := append(cfg.Exchanges[:0:0], cfg.Exchanges...)

	// Test empty exchange name for an enabled exchange
	cfg.Exchanges[0].Enabled = true
	cfg.Exchanges[0].Name = ""
	cfg.CheckExchangeConfigValues()
	if cfg.Exchanges[0].Enabled {
		t.Errorf(
			"Exchange with no name should be empty",
		)
	}

	// Test no enabled exchanges
	cfg.Exchanges = cfg.Exchanges[:1]
	cfg.Exchanges[0].Enabled = false
	err = cfg.CheckExchangeConfigValues()
	if err == nil {
		t.Error("Expected error from no enabled exchanges")
	}

	cfg.Exchanges = cpy
	// Check bank account validation for exchange
	cfg.Exchanges[0].BankAccounts = []banking.Account{
		{
			Enabled: true,
		},
	}

	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if cfg.Exchanges[0].BankAccounts[0].Enabled {
		t.Fatal("bank aaccount details not provided this should disable")
	}

	// Test international bank
	cfg.Exchanges[0].BankAccounts[0].Enabled = true
	cfg.Exchanges[0].BankAccounts[0].BankName = testString
	cfg.Exchanges[0].BankAccounts[0].BankAddress = testString
	cfg.Exchanges[0].BankAccounts[0].BankPostalCode = testString
	cfg.Exchanges[0].BankAccounts[0].BankPostalCity = testString
	cfg.Exchanges[0].BankAccounts[0].BankCountry = testString
	cfg.Exchanges[0].BankAccounts[0].AccountName = testString
	cfg.Exchanges[0].BankAccounts[0].SupportedCurrencies = "monopoly moneys"
	cfg.Exchanges[0].BankAccounts[0].IBAN = "some iban"
	cfg.Exchanges[0].BankAccounts[0].SWIFTCode = "some swifty"

	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if !cfg.Exchanges[0].BankAccounts[0].Enabled {
		t.Fatal("bank aaccount details provided this should not disable")
	}

	// Test aussie bank
	cfg.Exchanges[0].BankAccounts[0].Enabled = true
	cfg.Exchanges[0].BankAccounts[0].BankName = testString
	cfg.Exchanges[0].BankAccounts[0].BankAddress = testString
	cfg.Exchanges[0].BankAccounts[0].BankPostalCode = testString
	cfg.Exchanges[0].BankAccounts[0].BankPostalCity = testString
	cfg.Exchanges[0].BankAccounts[0].BankCountry = testString
	cfg.Exchanges[0].BankAccounts[0].AccountName = testString
	cfg.Exchanges[0].BankAccounts[0].SupportedCurrencies = "AUD"
	cfg.Exchanges[0].BankAccounts[0].BSBNumber = "some BSB nonsense"
	cfg.Exchanges[0].BankAccounts[0].IBAN = ""
	cfg.Exchanges[0].BankAccounts[0].SWIFTCode = ""

	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	if !cfg.Exchanges[0].BankAccounts[0].Enabled {
		t.Fatal("bank account details provided this should not disable")
	}

	cfg.Exchanges = nil
	cfg.Exchanges = append(cfg.Exchanges, cpy[0])

	cfg.Exchanges[0].CurrencyPairs.Pairs[asset.Spot].Enabled = nil
	cfg.Exchanges[0].CurrencyPairs.Pairs[asset.Spot].AssetEnabled = convert.BoolPtr(false)
	err = cfg.CheckExchangeConfigValues()
	if err != nil {
		t.Error(err)
	}

	cfg.Exchanges[0].CurrencyPairs.Pairs = make(map[asset.Item]*currency.PairStore)
	err = cfg.CheckExchangeConfigValues()
	if err == nil {
		t.Error("err cannot be nil")
	}
}
func TestGetExchangeAssetTypes(t *testing.T) {
	t.Parallel()
	var c Config
	_, err := c.GetExchangeAssetTypes("void")
	if err == nil {
		t.Error("err should have been thrown on a non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name: testFakeExchangeName,
			CurrencyPairs: &currency.PairsManager{
				Pairs: map[asset.Item]*currency.PairStore{
					asset.Spot:    new(currency.PairStore),
					asset.Futures: new(currency.PairStore),
				},
			},
		},
	)

	var assets asset.Items
	assets, err = c.GetExchangeAssetTypes(testFakeExchangeName)
	if err != nil {
		t.Error(err)
	}

	if !assets.Contains(asset.Spot) || !assets.Contains(asset.Futures) {
		t.Error("unexpected results")
	}

	c.Exchanges[0].CurrencyPairs = nil
	_, err = c.GetExchangeAssetTypes(testFakeExchangeName)
	if err == nil {
		t.Error("Expected error from nil currency pair")
	}
}

func TestGetEnabledPairs(t *testing.T) {
	t.Parallel()

	var c Config
	_, err := c.GetEnabledPairs("asdf", asset.Spot)
	if err == nil {
		t.Error("Expected error from non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name:          testFakeExchangeName,
			CurrencyPairs: &currency.PairsManager{},
		},
	)

	_, err = c.GetEnabledPairs(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error from nil pair manager")
	}

	c.Exchanges[0].CurrencyPairs.Pairs = map[asset.Item]*currency.PairStore{
		asset.Spot: {
			ConfigFormat: &currency.PairFormat{
				Delimiter: "-",
				Uppercase: true,
			},
		},
	}
	_, err = c.GetEnabledPairs(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error("nil pairs should return a nil error")
	}

	c.Exchanges[0].CurrencyPairs.Pairs[asset.Spot].Enabled = currency.Pairs{
		currency.NewPair(currency.BTC, currency.USD),
	}

	c.Exchanges[0].CurrencyPairs.Pairs[asset.Spot].Available = currency.Pairs{
		currency.NewPair(currency.BTC, currency.USD),
	}

	_, err = c.GetEnabledPairs(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPairFormat(t *testing.T) {
	t.Parallel()

	var c Config
	_, err := c.GetPairFormat("meow", asset.Spot)
	if err == nil {
		t.Error("Expected error from non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name: testFakeExchangeName,
		},
	)
	_, err = c.GetPairFormat(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error from nil pair manager")
	}

	c.Exchanges[0].CurrencyPairs = &currency.PairsManager{
		UseGlobalFormat: false,
		RequestFormat: &currency.PairFormat{
			Uppercase: false,
			Delimiter: "_",
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: "_",
		},
		Pairs: map[asset.Item]*currency.PairStore{
			asset.Spot: nil,
		},
	}

	_, err = c.GetPairFormat(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error from nil pair manager")
	}

	c.Exchanges[0].CurrencyPairs = &currency.PairsManager{
		UseGlobalFormat: true,
		RequestFormat: &currency.PairFormat{
			Uppercase: false,
			Delimiter: "_",
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: "_",
		},
		Pairs: map[asset.Item]*currency.PairStore{
			asset.Spot: new(currency.PairStore),
		},
	}
	_, err = c.GetPairFormat(testFakeExchangeName, asset.Item("invalid"))
	if err == nil {
		t.Error("Expected error from non-existent asset item")
	}

	_, err = c.GetPairFormat(testFakeExchangeName, asset.Futures)
	if err == nil {
		t.Error("Expected error from valid but non supported asset type")
	}

	var p currency.PairFormat
	p, err = c.GetPairFormat(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error(err)
	}

	if !p.Uppercase && p.Delimiter != "_" {
		t.Error("unexpected results")
	}

	// Test nil pair store
	c.Exchanges[0].CurrencyPairs.UseGlobalFormat = false
	_, err = c.GetPairFormat(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error")
	}

	c.Exchanges[0].CurrencyPairs.Pairs = map[asset.Item]*currency.PairStore{
		asset.Spot: {
			ConfigFormat: &currency.PairFormat{
				Uppercase: true,
				Delimiter: "~",
			},
		},
	}
	p, err = c.GetPairFormat(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error(err)
	}

	if p.Delimiter != "~" && !p.Uppercase {
		t.Error("unexpected results")
	}
}

func TestSupportsExchangeAssetType(t *testing.T) {
	t.Parallel()
	var c Config
	err := c.SupportsExchangeAssetType("void", asset.Spot)
	if err == nil {
		t.Error("Expected error for non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name: testFakeExchangeName,
			CurrencyPairs: &currency.PairsManager{
				Pairs: map[asset.Item]*currency.PairStore{
					asset.Spot: new(currency.PairStore),
				},
			},
		},
	)

	err = c.SupportsExchangeAssetType(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error(err)
	}

	err = c.SupportsExchangeAssetType(testFakeExchangeName, "asdf")
	if err == nil {
		t.Error("Expected error from invalid asset item")
	}

	c.Exchanges[0].CurrencyPairs = nil
	err = c.SupportsExchangeAssetType(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error from nil pair manager")
	}
}

func TestGetAvailablePairs(t *testing.T) {
	t.Parallel()

	var c Config
	_, err := c.GetAvailablePairs("asdf", asset.Spot)
	if err == nil {
		t.Error("Expected error from non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name:          testFakeExchangeName,
			CurrencyPairs: &currency.PairsManager{},
		},
	)

	_, err = c.GetAvailablePairs(testFakeExchangeName, asset.Spot)
	if err == nil {
		t.Error("Expected error from nil pair manager")
	}

	c.Exchanges[0].CurrencyPairs.Pairs = map[asset.Item]*currency.PairStore{
		asset.Spot: {
			ConfigFormat: &currency.PairFormat{
				Delimiter: "-",
				Uppercase: true,
			},
		},
	}
	_, err = c.GetAvailablePairs(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error("Expected error from nil pairs")
	}

	c.Exchanges[0].CurrencyPairs.Pairs[asset.Spot].Available = currency.Pairs{
		currency.NewPair(currency.BTC, currency.USD),
	}
	_, err = c.GetAvailablePairs(testFakeExchangeName, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestSetPairs(t *testing.T) {
	t.Parallel()

	var c Config
	pairs := currency.Pairs{
		currency.NewPair(currency.BTC, currency.USD),
		currency.NewPair(currency.BTC, currency.EUR),
	}

	err := c.SetPairs("asdf", asset.Spot, true, nil)
	if err == nil {
		t.Error("Expected error from nil pairs")
	}

	err = c.SetPairs("asdf", asset.Spot, true, pairs)
	if err == nil {
		t.Error("Expected error from non-existent exchange")
	}

	c.Exchanges = append(c.Exchanges,
		ExchangeConfig{
			Name: testFakeExchangeName,
		},
	)

	err = c.SetPairs(testFakeExchangeName, asset.Index, true, pairs)
	if err == nil {
		t.Error("Expected error from non initialised pair manager")
	}

	c.Exchanges[0].CurrencyPairs = &currency.PairsManager{
		Pairs: map[asset.Item]*currency.PairStore{
			asset.Spot: new(currency.PairStore),
		},
	}

	err = c.SetPairs(testFakeExchangeName, asset.Index, true, pairs)
	if err == nil {
		t.Error("Expected error from non supported asset type")
	}

	err = c.SetPairs(testFakeExchangeName, asset.Spot, true, pairs)
	if err != nil {
		t.Error(err)
	}
}
