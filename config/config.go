package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/common/convert"
	"github.com/openware/pkg/common/file"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
)

// Constants declared here are filename strings and test strings
const (
	FXProviderFixer                      = "fixer"
	EncryptedFile                        = "config.dat"
	File                                 = "config.json"
	TestFile                             = "../configtest.json"
	fileEncryptionPrompt                 = 0
	fileEncryptionEnabled                = 1
	fileEncryptionDisabled               = -1
	pairsLastUpdatedWarningThreshold     = 30 // 30 days
	defaultHTTPTimeout                   = time.Second * 15
	defaultWebsocketResponseCheckTimeout = time.Millisecond * 30
	defaultWebsocketResponseMaxLimit     = time.Second * 7
	defaultWebsocketOrderbookBufferLimit = 5
	defaultWebsocketTrafficTimeout       = time.Second * 30
	maxAuthFailures                      = 3
	defaultNTPAllowedDifference          = 50000000
	defaultNTPAllowedNegativeDifference  = 50000000
	DefaultAPIKey                        = "Key"
	DefaultAPISecret                     = "Secret"
	DefaultAPIClientID                   = "ClientID"
)

// Constants here hold some messages
const (
	ErrExchangeNameEmpty                       = "exchange #%d name is empty"
	ErrExchangeAvailablePairsEmpty             = "exchange %s available pairs is empty"
	ErrExchangeEnabledPairsEmpty               = "exchange %s enabled pairs is empty"
	ErrExchangeBaseCurrenciesEmpty             = "exchange %s base currencies is empty"
	ErrExchangeNotFound                        = "exchange %s not found"
	ErrNoEnabledExchanges                      = "no exchanges enabled"
	ErrCryptocurrenciesEmpty                   = "cryptocurrencies variable is empty"
	ErrFailureOpeningConfig                    = "fatal error opening %s file. Error: %s"
	ErrCheckingConfigValues                    = "fatal error checking config values. Error: %s"
	ErrSavingConfigBytesMismatch               = "config file %q bytes comparison doesn't match, read %s expected %s"
	WarningWebserverCredentialValuesEmpty      = "webserver support disabled due to empty Username/Password values"
	WarningWebserverListenAddressInvalid       = "webserver support disabled due to invalid listen address"
	WarningExchangeAuthAPIDefaultOrEmptyValues = "exchange %s authenticated API support disabled due to default/empty APIKey/Secret/ClientID values"
	WarningPairsLastUpdatedThresholdExceeded   = "exchange %s last manual update of available currency pairs has exceeded %d days. Manual update required!"
)

// Constants here define unset default values displayed in the config.json
// file
const (
	APIURLNonDefaultMessage              = "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API"
	WebsocketURLNonDefaultMessage        = "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
	DefaultUnsetAPIKey                   = "Key"
	DefaultUnsetAPISecret                = "Secret"
	DefaultUnsetAccountPlan              = "accountPlan"
	DefaultForexProviderExchangeRatesAPI = "ExchangeRates"
)

var Cfg Config

// GetConfig returns a pointer to a configuration object
func GetConfig() *Config {
	return &Cfg
}

// LoadConfig loads your configuration file into your configuration object
func (c *Config) LoadConfig(configPath string, dryrun bool) error {
	err := c.ReadConfigFromFile(configPath, dryrun)
	if err != nil {
		return fmt.Errorf(ErrFailureOpeningConfig, configPath, err)
	}

	return c.CheckConfig()
}

// GetFilePath returns the desired config file or the default config file name
// and whether it was loaded from a default location (rather than explicitly specified)
func GetFilePath(configfile string) (configPath string, isImplicitDefaultPath bool, err error) {
	if configfile != "" {
		return configfile, false, nil
	}

	exePath, err := common.GetExecutablePath()
	if err != nil {
		return "", false, err
	}
	newDir := common.GetDefaultDataDir(runtime.GOOS)
	defaultPaths := []string{
		filepath.Join(exePath, File),
		filepath.Join(exePath, EncryptedFile),
		filepath.Join(newDir, File),
		filepath.Join(newDir, EncryptedFile),
	}

	for _, p := range defaultPaths {
		if file.Exists(p) {
			configfile = p
			break
		}
	}
	if configfile == "" {
		return "", false, fmt.Errorf("config.json file not found in %s, please follow README.md in root dir for config generation",
			newDir)
	}

	return configfile, true, nil
}

// ReadConfigFromFile reads the configuration from the given file
// if target file is encrypted, prompts for encryption key
// Also - if not in dryrun mode - it checks if the configuration needs to be encrypted
// and stores the file as encrypted, if necessary (prompting for enryption key)
func (c *Config) ReadConfigFromFile(configPath string, dryrun bool) error {
	defaultPath, _, err := GetFilePath(configPath)
	if err != nil {
		return err
	}
	confFile, err := os.Open(defaultPath)
	if err != nil {
		return err
	}
	defer confFile.Close()
	result, err := ReadConfig(confFile)
	if err != nil {
		return fmt.Errorf("error reading config %w", err)
	}
	// Override values in the current config
	*c = *result

	if dryrun {
		return nil
	}

	return nil
}

// ReadConfig verifies and checks for encryption and loads the config from a JSON object.
// Prompts for decryption key, if target data is encrypted.
// Returns the loaded configuration and whether it was encrypted.
func ReadConfig(configReader io.Reader) (*Config, error) {
	reader := bufio.NewReader(configReader)

	// Read unencrypted configuration
	decoder := json.NewDecoder(reader)
	c := &Config{}
	err := decoder.Decode(c)
	return c, err
}

// GetExchangeConfig returns exchange configurations by its indivdual name
func (c *Config) GetExchangeConfig(name string) (*ExchangeConfig, error) {
	for i := range c.Exchanges {
		if strings.EqualFold(c.Exchanges[i].Name, name) {
			return &c.Exchanges[i], nil
		}
	}
	return nil, fmt.Errorf(ErrExchangeNotFound, name)
}

// CheckConfig checks all config settings
func (c *Config) CheckConfig() error {
	if err := c.CheckExchangeConfigValues(); err != nil {
		return fmt.Errorf(ErrCheckingConfigValues, err)
	}

	return nil
}

// CheckExchangeConfigValues returns configuation values for all enabled
// exchanges
func (c *Config) CheckExchangeConfigValues() error {
	if len(c.Exchanges) == 0 {
		return errors.New("no exchange configs found")
	}

	exchanges := 0
	for i := range c.Exchanges {
		if strings.EqualFold(c.Exchanges[i].Name, "GDAX") {
			c.Exchanges[i].Name = "CoinbasePro"
		}

		// Check to see if the old API storage format is used
		if c.Exchanges[i].APIKey != nil {
			// It is, migrate settings to new format
			c.Exchanges[i].API.AuthenticatedSupport = *c.Exchanges[i].AuthenticatedAPISupport
			if c.Exchanges[i].AuthenticatedWebsocketAPISupport != nil {
				c.Exchanges[i].API.AuthenticatedWebsocketSupport = *c.Exchanges[i].AuthenticatedWebsocketAPISupport
			}
			c.Exchanges[i].API.Credentials.Key = *c.Exchanges[i].APIKey
			c.Exchanges[i].API.Credentials.Secret = *c.Exchanges[i].APISecret

			if c.Exchanges[i].APIAuthPEMKey != nil {
				c.Exchanges[i].API.Credentials.PEMKey = *c.Exchanges[i].APIAuthPEMKey
			}

			if c.Exchanges[i].APIAuthPEMKeySupport != nil {
				c.Exchanges[i].API.PEMKeySupport = *c.Exchanges[i].APIAuthPEMKeySupport
			}

			if c.Exchanges[i].ClientID != nil {
				c.Exchanges[i].API.Credentials.ClientID = *c.Exchanges[i].ClientID
			}

			// Flush settings
			c.Exchanges[i].AuthenticatedAPISupport = nil
			c.Exchanges[i].AuthenticatedWebsocketAPISupport = nil
			c.Exchanges[i].APIKey = nil
			c.Exchanges[i].APISecret = nil
			c.Exchanges[i].ClientID = nil
			c.Exchanges[i].APIAuthPEMKeySupport = nil
			c.Exchanges[i].APIAuthPEMKey = nil
			c.Exchanges[i].APIURL = nil
			c.Exchanges[i].APIURLSecondary = nil
			c.Exchanges[i].WebsocketURL = nil
		}

		if c.Exchanges[i].Features == nil {
			c.Exchanges[i].Features = &FeaturesConfig{}
		}

		if c.Exchanges[i].SupportsAutoPairUpdates != nil {
			c.Exchanges[i].Features.Supports.RESTCapabilities.AutoPairUpdates = *c.Exchanges[i].SupportsAutoPairUpdates
			c.Exchanges[i].Features.Enabled.AutoPairUpdates = *c.Exchanges[i].SupportsAutoPairUpdates
			c.Exchanges[i].SupportsAutoPairUpdates = nil
		}

		if c.Exchanges[i].Websocket != nil {
			c.Exchanges[i].Features.Enabled.Websocket = *c.Exchanges[i].Websocket
			c.Exchanges[i].Websocket = nil
		}

		// Check if see if the new currency pairs format is empty and flesh it out if so
		if c.Exchanges[i].CurrencyPairs == nil {
			c.Exchanges[i].CurrencyPairs = new(currency.PairsManager)
			c.Exchanges[i].CurrencyPairs.Pairs = make(map[asset.Item]*currency.PairStore)

			if c.Exchanges[i].PairsLastUpdated != nil {
				c.Exchanges[i].CurrencyPairs.LastUpdated = *c.Exchanges[i].PairsLastUpdated
			}

			c.Exchanges[i].CurrencyPairs.ConfigFormat = c.Exchanges[i].ConfigCurrencyPairFormat
			c.Exchanges[i].CurrencyPairs.RequestFormat = c.Exchanges[i].RequestCurrencyPairFormat

			var availPairs, enabledPairs currency.Pairs
			if c.Exchanges[i].AvailablePairs != nil {
				availPairs = *c.Exchanges[i].AvailablePairs
			}

			if c.Exchanges[i].EnabledPairs != nil {
				enabledPairs = *c.Exchanges[i].EnabledPairs
			}

			c.Exchanges[i].CurrencyPairs.UseGlobalFormat = true
			c.Exchanges[i].CurrencyPairs.Store(asset.Spot,
				currency.PairStore{
					AssetEnabled: convert.BoolPtr(true),
					Available:    availPairs,
					Enabled:      enabledPairs,
				},
			)

			// flush old values
			c.Exchanges[i].PairsLastUpdated = nil
			c.Exchanges[i].ConfigCurrencyPairFormat = nil
			c.Exchanges[i].RequestCurrencyPairFormat = nil
			c.Exchanges[i].AssetTypes = nil
			c.Exchanges[i].AvailablePairs = nil
			c.Exchanges[i].EnabledPairs = nil
		} else {
			assets := c.Exchanges[i].CurrencyPairs.GetAssetTypes()
			var atLeastOne bool
			for index := range assets {
				err := c.Exchanges[i].CurrencyPairs.IsAssetEnabled(assets[index])
				if err != nil {
					// Checks if we have an old config without the ability to
					// enable disable the entire asset
					if err.Error() == "cannot ascertain if asset is enabled, variable is nil" {
						log.Warnf(log.ConfigMgr,
							"Exchange %s: upgrading config for asset type %s and setting enabled.\n",
							c.Exchanges[i].Name,
							assets[index])
						err = c.Exchanges[i].CurrencyPairs.SetAssetEnabled(assets[index], true)
						if err != nil {
							return err
						}
						atLeastOne = true
					}
					continue
				}
				atLeastOne = true
			}

			if !atLeastOne {
				if len(assets) == 0 {
					c.Exchanges[i].Enabled = false
					log.Warnf(log.ConfigMgr,
						"%s no assets found, disabling...",
						c.Exchanges[i].Name)
					continue
				}

				// turn on an asset if all disabled
				log.Warnf(log.ConfigMgr,
					"%s assets disabled, turning on asset %s",
					c.Exchanges[i].Name,
					assets[0])

				err := c.Exchanges[i].CurrencyPairs.SetAssetEnabled(assets[0], true)
				if err != nil {
					return err
				}
			}
		}

		if c.Exchanges[i].Enabled {
			if c.Exchanges[i].Name == "" {
				log.Errorf(log.ConfigMgr, ErrExchangeNameEmpty, i)
				c.Exchanges[i].Enabled = false
				continue
			}
			if (c.Exchanges[i].API.AuthenticatedSupport || c.Exchanges[i].API.AuthenticatedWebsocketSupport) &&
				c.Exchanges[i].API.CredentialsValidator != nil {
				var failed bool
				if c.Exchanges[i].API.CredentialsValidator.RequiresKey &&
					(c.Exchanges[i].API.Credentials.Key == "" || c.Exchanges[i].API.Credentials.Key == DefaultAPIKey) {
					failed = true
				}

				if c.Exchanges[i].API.CredentialsValidator.RequiresSecret &&
					(c.Exchanges[i].API.Credentials.Secret == "" || c.Exchanges[i].API.Credentials.Secret == DefaultAPISecret) {
					failed = true
				}

				if c.Exchanges[i].API.CredentialsValidator.RequiresClientID &&
					(c.Exchanges[i].API.Credentials.ClientID == DefaultAPIClientID || c.Exchanges[i].API.Credentials.ClientID == "") {
					failed = true
				}

				if failed {
					c.Exchanges[i].API.AuthenticatedSupport = false
					c.Exchanges[i].API.AuthenticatedWebsocketSupport = false
					log.Warnf(log.ConfigMgr, WarningExchangeAuthAPIDefaultOrEmptyValues, c.Exchanges[i].Name)
				}
			}
			if !c.Exchanges[i].Features.Supports.RESTCapabilities.AutoPairUpdates &&
				!c.Exchanges[i].Features.Supports.WebsocketCapabilities.AutoPairUpdates {
				lastUpdated := convert.UnixTimestampToTime(c.Exchanges[i].CurrencyPairs.LastUpdated)
				lastUpdated = lastUpdated.AddDate(0, 0, pairsLastUpdatedWarningThreshold)
				if lastUpdated.Unix() <= time.Now().Unix() {
					log.Warnf(log.ConfigMgr,
						WarningPairsLastUpdatedThresholdExceeded,
						c.Exchanges[i].Name,
						pairsLastUpdatedWarningThreshold)
				}
			}
			if c.Exchanges[i].HTTPTimeout <= 0 {
				log.Warnf(log.ConfigMgr,
					"Exchange %s HTTP Timeout value not set, defaulting to %v.\n",
					c.Exchanges[i].Name,
					defaultHTTPTimeout)
				c.Exchanges[i].HTTPTimeout = defaultHTTPTimeout
			}

			if c.Exchanges[i].WebsocketResponseCheckTimeout <= 0 {
				log.Warnf(log.ConfigMgr,
					"Exchange %s Websocket response check timeout value not set, defaulting to %v.",
					c.Exchanges[i].Name,
					defaultWebsocketResponseCheckTimeout)
				c.Exchanges[i].WebsocketResponseCheckTimeout = defaultWebsocketResponseCheckTimeout
			}

			if c.Exchanges[i].WebsocketResponseMaxLimit <= 0 {
				log.Warnf(log.ConfigMgr,
					"Exchange %s Websocket response max limit value not set, defaulting to %v.",
					c.Exchanges[i].Name,
					defaultWebsocketResponseMaxLimit)
				c.Exchanges[i].WebsocketResponseMaxLimit = defaultWebsocketResponseMaxLimit
			}
			if c.Exchanges[i].WebsocketTrafficTimeout <= 0 {
				log.Warnf(log.ConfigMgr,
					"Exchange %s Websocket response traffic timeout value not set, defaulting to %v.",
					c.Exchanges[i].Name,
					defaultWebsocketTrafficTimeout)
				c.Exchanges[i].WebsocketTrafficTimeout = defaultWebsocketTrafficTimeout
			}
			if c.Exchanges[i].OrderbookConfig.WebsocketBufferLimit <= 0 {
				log.Warnf(log.ConfigMgr,
					"Exchange %s Websocket orderbook buffer limit value not set, defaulting to %v.",
					c.Exchanges[i].Name,
					defaultWebsocketOrderbookBufferLimit)
				c.Exchanges[i].OrderbookConfig.WebsocketBufferLimit = defaultWebsocketOrderbookBufferLimit
			}
			err := c.CheckPairConsistency(c.Exchanges[i].Name)
			if err != nil {
				log.Errorf(log.ConfigMgr,
					"Exchange %s: CheckPairConsistency error: %s\n",
					c.Exchanges[i].Name,
					err)
				c.Exchanges[i].Enabled = false
				continue
			}
			for x := range c.Exchanges[i].BankAccounts {
				if !c.Exchanges[i].BankAccounts[x].Enabled {
					continue
				}
				err := c.Exchanges[i].BankAccounts[x].Validate()
				if err != nil {
					c.Exchanges[i].BankAccounts[x].Enabled = false
					log.Warnln(log.ConfigMgr, err.Error())
				}
			}
			exchanges++
		}
	}

	if exchanges == 0 {
		return errors.New(ErrNoEnabledExchanges)
	}
	return nil
}

// CheckPairConsistency checks to see if the enabled pair exists in the
// available pairs list
func (c *Config) CheckPairConsistency(exchName string) error {
	assetTypes, err := c.GetExchangeAssetTypes(exchName)
	if err != nil {
		return err
	}

	var atLeastOneEnabled bool
	for x := range assetTypes {
		enabledPairs, err := c.GetEnabledPairs(exchName, assetTypes[x])
		if err == nil {
			if len(enabledPairs) != 0 {
				atLeastOneEnabled = true
				continue
			}
			var enabled bool
			enabled, err = c.AssetTypeEnabled(assetTypes[x], exchName)
			if err != nil {
				return err
			}

			if !enabled {
				continue
			}

			var availPairs currency.Pairs
			availPairs, err = c.GetAvailablePairs(exchName, assetTypes[x])
			if err != nil {
				return err
			}

			err = c.SetPairs(exchName,
				assetTypes[x],
				true,
				currency.Pairs{availPairs.GetRandomPair()})
			if err != nil {
				return err
			}
			atLeastOneEnabled = true
			continue
		}

		// On error an enabled pair is not found in the available pairs list
		// so remove and report
		availPairs, err := c.GetAvailablePairs(exchName, assetTypes[x])
		if err != nil {
			return err
		}

		var pairs, pairsRemoved currency.Pairs
		for x := range enabledPairs {
			if !availPairs.Contains(enabledPairs[x], true) {
				pairsRemoved = append(pairsRemoved, enabledPairs[x])
				continue
			}
			pairs = append(pairs, enabledPairs[x])
		}

		if len(pairsRemoved) == 0 {
			return fmt.Errorf("check pair consistency fault for asset %s, conflict found but no pairs removed",
				assetTypes[x])
		}

		// Flush corrupted/misspelled enabled pairs in config
		err = c.SetPairs(exchName, assetTypes[x], true, pairs)
		if err != nil {
			return err
		}

		log.Warnf(log.ConfigMgr,
			"Exchange %s: [%v] Removing enabled pair(s) %v from enabled pairs list, as it isn't located in the available pairs list.\n",
			exchName,
			assetTypes[x],
			pairsRemoved.Strings())

		if len(pairs) != 0 {
			atLeastOneEnabled = true
			continue
		}

		enabled, err := c.AssetTypeEnabled(assetTypes[x], exchName)
		if err != nil {
			return err
		}

		if !enabled {
			continue
		}

		err = c.SetPairs(exchName,
			assetTypes[x],
			true,
			currency.Pairs{availPairs.GetRandomPair()})
		if err != nil {
			return err
		}
		atLeastOneEnabled = true
	}

	// If no pair is enabled across the entire range of assets, then atleast
	// enable one and turn on the asset type
	if !atLeastOneEnabled {
		avail, err := c.GetAvailablePairs(exchName, assetTypes[0])
		if err != nil {
			return err
		}

		newPair := avail.GetRandomPair()
		err = c.SetPairs(exchName, assetTypes[0], true, currency.Pairs{newPair})
		if err != nil {
			return err
		}
		log.Warnf(log.ConfigMgr,
			"Exchange %s: [%v] No enabled pairs found in available pairs list, randomly added %v pair.\n",
			exchName,
			assetTypes[0],
			newPair)
	}
	return nil
}

// GetExchangeAssetTypes returns the exchanges supported asset types
func (c *Config) GetExchangeAssetTypes(exchName string) (asset.Items, error) {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return nil, err
	}

	if exchCfg.CurrencyPairs == nil {
		return nil, fmt.Errorf("exchange %s currency pairs is nil", exchName)
	}

	return exchCfg.CurrencyPairs.GetAssetTypes(), nil
}

// GetEnabledPairs returns a list of currency pairs for a specifc exchange
func (c *Config) GetEnabledPairs(exchName string, assetType asset.Item) (currency.Pairs, error) {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return nil, err
	}

	pairFormat, err := c.GetPairFormat(exchName, assetType)
	if err != nil {
		return nil, err
	}

	pairs, err := exchCfg.CurrencyPairs.GetPairs(assetType, true)
	if err != nil {
		return pairs, err
	}

	if pairs == nil {
		return nil, nil
	}

	return pairs.Format(pairFormat.Delimiter,
			pairFormat.Index,
			pairFormat.Uppercase),
		nil
}

// GetPairFormat returns the exchanges pair config storage format
func (c *Config) GetPairFormat(exchName string, assetType asset.Item) (currency.PairFormat, error) {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return currency.PairFormat{}, err
	}

	err = c.SupportsExchangeAssetType(exchName, assetType)
	if err != nil {
		return currency.PairFormat{}, err
	}

	if exchCfg.CurrencyPairs.UseGlobalFormat {
		return *exchCfg.CurrencyPairs.ConfigFormat, nil
	}

	p, err := exchCfg.CurrencyPairs.Get(assetType)
	if err != nil {
		return currency.PairFormat{}, err
	}

	if p == nil {
		return currency.PairFormat{},
			fmt.Errorf("exchange %s pair store for asset type %s is nil",
				exchName,
				assetType)
	}

	if p.ConfigFormat == nil {
		return currency.PairFormat{},
			fmt.Errorf("exchange %s pair config format for asset type %s is nil",
				exchName,
				assetType)
	}

	return *p.ConfigFormat, nil
}

// SupportsExchangeAssetType returns whether or not the exchange supports the supplied asset type
func (c *Config) SupportsExchangeAssetType(exchName string, assetType asset.Item) error {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return err
	}

	if exchCfg.CurrencyPairs == nil {
		return fmt.Errorf("exchange %s currency pairs is nil", exchName)
	}

	if !assetType.IsValid() {
		return fmt.Errorf("exchange %s invalid asset type %s",
			exchName,
			assetType)
	}

	if !exchCfg.CurrencyPairs.GetAssetTypes().Contains(assetType) {
		return fmt.Errorf("exchange %s unsupported asset type %s",
			exchName,
			assetType)
	}
	return nil
}

// AssetTypeEnabled checks to see if the asset type is enabled in configuration
func (c *Config) AssetTypeEnabled(a asset.Item, exch string) (bool, error) {
	cfg, err := c.GetExchangeConfig(exch)
	if err != nil {
		return false, err
	}

	err = cfg.CurrencyPairs.IsAssetEnabled(a)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetAvailablePairs returns a list of currency pairs for a specifc exchange
func (c *Config) GetAvailablePairs(exchName string, assetType asset.Item) (currency.Pairs, error) {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return nil, err
	}

	pairFormat, err := c.GetPairFormat(exchName, assetType)
	if err != nil {
		return nil, err
	}

	pairs, err := exchCfg.CurrencyPairs.GetPairs(assetType, false)
	if err != nil {
		return nil, err
	}

	if pairs == nil {
		return nil, nil
	}

	return pairs.Format(pairFormat.Delimiter, pairFormat.Index,
		pairFormat.Uppercase), nil
}

// SetPairs sets the exchanges currency pairs
func (c *Config) SetPairs(exchName string, assetType asset.Item, enabled bool, pairs currency.Pairs) error {
	exchCfg, err := c.GetExchangeConfig(exchName)
	if err != nil {
		return err
	}

	err = c.SupportsExchangeAssetType(exchName, assetType)
	if err != nil {
		return err
	}

	exchCfg.CurrencyPairs.StorePairs(assetType, pairs, enabled)
	return nil
}
