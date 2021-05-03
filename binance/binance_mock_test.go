//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package binance

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockfile = "./binance.mock.json"

var mockTests = true

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	binanceConfig, err := config.FromFile(filepath.Join(wd, "binance.conf.json"))
	if err != nil {
		log.Fatal("Binance Setup() init error", err)
	}
	b.SkipAuthCheck = true
	binanceConfig.API.AuthenticatedSupport = true
	binanceConfig.API.Credentials.Key = apiKey
	binanceConfig.API.Credentials.Secret = apiSecret
	b.SetDefaults()
	b.Websocket = sharedtestvalues.NewTestWebsocket()
	err = b.Setup(binanceConfig)
	if err != nil {
		log.Fatal("Binance setup error", err)
	}

	b.setupOrderbookManager()

	serverDetails, newClient, err := mock.NewVCRServer(mockfile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}
	b.HTTPClient = newClient
	endpointMap := b.API.Endpoints.GetURLMap()
	for k := range endpointMap {
		err = b.API.Endpoints.SetRunning(k, serverDetails)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf(sharedtestvalues.MockTesting, b.Name)
	os.Exit(m.Run())
}
