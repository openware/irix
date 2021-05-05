//+build mock_test_off

// This will build if build tag mock_test_off is parsed and will do live testing
// using all tests in (exchange)_test.go
package bitstamp

import (
	"log"
	"os"
	"testing"

	"github.com/openware/irix/sharedtestvalues"
)

var mockTests = false

func TestMain(m *testing.M) {
	bitstampConfig, err := configTest()
	if err != nil {
		log.Fatal("Bitstamp Setup() init error", err)
	}
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
	log.Printf(sharedtestvalues.LiveTesting, b.Name)
	os.Exit(m.Run())
}
