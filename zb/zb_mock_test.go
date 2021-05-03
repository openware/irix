//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package zb

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockfile = "./zb.mock.json"

var mockTests = true

func configTest() (*config.ExchangeConfig, error) {
	wd, _ := os.Getwd()
	return config.FromFile(filepath.Join(wd, "zb.conf.json"))
}
func TestMain(m *testing.M) {
	zbConfig, err := configTest()
	if err != nil {
		log.Fatal("ZB Setup() init error", err)
	}
	zbConfig.API.AuthenticatedSupport = true
	zbConfig.API.AuthenticatedWebsocketSupport = true
	zbConfig.API.Credentials.Key = apiKey
	zbConfig.API.Credentials.Secret = apiSecret
	z.SkipAuthCheck = true
	z.SetDefaults()
	z.Websocket = sharedtestvalues.NewTestWebsocket()
	err = z.Setup(zbConfig)
	if err != nil {
		log.Fatal("ZB setup error", err)
	}

	serverDetails, newClient, err := mock.NewVCRServer(mockfile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}

	z.HTTPClient = newClient
	endpoints := z.API.Endpoints.GetURLMap()
	for k := range endpoints {
		err = z.API.Endpoints.SetRunning(k, serverDetails)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf(sharedtestvalues.MockTesting,
		z.Name)

	os.Exit(m.Run())
}
