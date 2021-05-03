//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package gemini

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openware/irix/config"
	"github.com/openware/irix/sharedtestvalues"
	"github.com/openware/pkg/mock"
)

const mockFile = "./gemini.mock.json"

var mockTests = true
func configTest() (*config.ExchangeConfig, error) {
	wd, _ := os.Getwd()
	return config.FromFile(filepath.Join(wd, "gemini.conf.json"))
}

func TestMain(m *testing.M) {
	geminiConfig, err := configTest()
	if err != nil {
		log.Fatal("Mock server error", err)
	}
	g.SkipAuthCheck = true
	geminiConfig.API.AuthenticatedSupport = true
	geminiConfig.API.Credentials.Key = apiKey
	geminiConfig.API.Credentials.Secret = apiSecret
	g.SetDefaults()
	g.Websocket = sharedtestvalues.NewTestWebsocket()
	err = g.Setup(geminiConfig)
	if err != nil {
		log.Fatal("Gemini setup error", err)
	}

	serverDetails, newClient, err := mock.NewVCRServer(mockFile)
	if err != nil {
		log.Fatalf("Mock server error %s", err)
	}

	g.HTTPClient = newClient
	endpointMap := g.API.Endpoints.GetURLMap()
	for k := range endpointMap {
		err = g.API.Endpoints.SetRunning(k, serverDetails)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf(sharedtestvalues.MockTesting, g.Name)
	os.Exit(m.Run())
}
