package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/openware/pkg/common"
	"github.com/openware/pkg/common/file"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)
// TODO: remove this file after all configs are moved using simpler API
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

	return nil
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
