//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package localbitcoins

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockfile = "./localbitcoins.mock.json"

var mockTests = true

func configTest() (*config.ExchangeConfig, error) {
	wd, _ := os.Getwd()
	return config.FromFile(filepath.Join(wd, "localbitcoins.conf.json"))
}

func TestMain(m *testing.M) {
	localbitcoinsConfig, err := configTest()
	if err != nil {
		log.Fatal("Localbitcoins Setup() init error", err)
	}
	l.SkipAuthCheck = true
	localbitcoinsConfig.API.AuthenticatedSupport = true
	localbitcoinsConfig.API.Credentials.Key = apiKey
	localbitcoinsConfig.API.Credentials.Secret = apiSecret
	l.SetDefaults()
	err = l.Setup(localbitcoinsConfig)
	if err != nil {
		log.Fatal("Localbitcoins setup error", err)
	}

	serverDetails, newClient, err := mock.NewVCRServer(mockfile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}

	l.HTTPClient = newClient
	endpoints := l.API.Endpoints.GetURLMap()
	for k := range endpoints {
		err = l.API.Endpoints.SetRunning(k, serverDetails)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf(sharedtestvalues.MockTesting, l.Name)
	os.Exit(m.Run())
}
