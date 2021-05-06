//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package poloniex

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockfile = "./poloniex.mock.json"

var mockTests = true

func configTest() (*config.ExchangeConfig, error) {
	wd, _ := os.Getwd()
	return config.FromFile(filepath.Join(wd, "poloniex.conf.json"))
}

func TestMain(m *testing.M) {
	poloniexConfig, err := configTest()
	if err != nil {
		log.Fatal("Poloniex Setup() init error", err)
	}
	p.SkipAuthCheck = true
	poloniexConfig.API.AuthenticatedSupport = true
	poloniexConfig.API.Credentials.Key = apiKey
	poloniexConfig.API.Credentials.Secret = apiSecret
	p.SetDefaults()
	p.Websocket = sharedtestvalues.NewTestWebsocket()
	err = p.Setup(poloniexConfig)
	if err != nil {
		log.Fatal("Poloniex setup error", err)
	}

	serverDetails, newClient, err := mock.NewVCRServer(mockfile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}

	p.HTTPClient = newClient
	endpoints := p.API.Endpoints.GetURLMap()
	for k := range endpoints {
		err = p.API.Endpoints.SetRunning(k, serverDetails)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf(sharedtestvalues.MockTesting, p.Name)
	os.Exit(m.Run())
}
