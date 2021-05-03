package config

// TODO: remove this file after all configs are moved using simpler API
type Config struct {
	Name      string           `json:"name"`
	Exchanges []ExchangeConfig `json:"exchanges"`
}
