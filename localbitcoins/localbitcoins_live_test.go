//+build mock_test_off

// This will build if build tag mock_test_off is parsed and will do live testing
// using all tests in (exchange)_test.go
package localbitcoins

import (
	"log"
	"os"
	"testing"

	"github.com/openware/irix/sharedtestvalues"
)

var mockTests = false

func TestMain(m *testing.M) {
	localbitcoinsConfig, err := configTest()
	if err != nil {
		log.Fatal("LocalBitcoins Setup() init error", err)
	}
	localbitcoinsConfig.API.AuthenticatedSupport = true
	localbitcoinsConfig.API.Credentials.Key = apiKey
	localbitcoinsConfig.API.Credentials.Secret = apiSecret
	l.SetDefaults()
	err = l.Setup(localbitcoinsConfig)
	if err != nil {
		log.Fatal("Localbitcoins setup error", err)
	}
	log.Printf(sharedtestvalues.LiveTesting, l.Name)
	os.Exit(m.Run())
}
