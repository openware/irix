//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package bitstamp

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockfile = "../../testdata/http_mock/bitstamp/bitstamp.json"

func configTest() (*config.ExchangeConfig, error) {
	wd, _ := os.Getwd()
	return config.FromFile(filepath.Join(wd, "bitstamp.conf.json"))
}

var mockTests = true

func TestMain(m *testing.M) {
	bitstampConfig, err := configTest()
	if err != nil {
		log.Fatal("Bitstamp Setup() init error", err)
	}
	b.SkipAuthCheck = true
	bitstampConfig.API.AuthenticatedSupport = true
	bitstampConfig.API.Credentials.Key = apiKey
	bitstampConfig.API.Credentials.Secret = apiSecret
	bitstampConfig.API.Credentials.ClientID = customerID
	b.SetDefaults()
	b.Websocket = sharedtestvalues.NewTestWebsocket()
	err = b.Setup(bitstampConfig)
	if err != nil {
		log.Fatal("Bitstamp setup error", err)
	}

	serverDetails, newClient, err := mock.NewVCRServer(mockfile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}

	b.HTTPClient = newClient
	endpointMap := b.API.Endpoints.GetURLMap()
	for k := range endpointMap {
		err = b.API.Endpoints.SetRunning(k, serverDetails+"/api")
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf(sharedtestvalues.MockTesting, b.Name)
	os.Exit(m.Run())
}
