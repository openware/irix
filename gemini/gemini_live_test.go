//+build mock_test_off

// This will build if build tag mock_test_off is parsed and will do live testing
// using all tests in (exchange)_test.go
package gemini

import (
	"log"
	"os"
	"testing"

	exchange "github.com/openware/irix"
	"github.com/openware/irix/sharedtestvalues"
)

var mockTests = false

func TestMain(m *testing.M) {
	geminiConfig, err := configTest()
	if err != nil {
		log.Fatal("Gemini Setup() init error", err)
	}
	geminiConfig.API.AuthenticatedSupport = true
	geminiConfig.API.Credentials.Key = apiKey
	geminiConfig.API.Credentials.Secret = apiSecret
	g.SetDefaults()
	g.Websocket = sharedtestvalues.NewTestWebsocket()
	err = g.Setup(geminiConfig)
	if err != nil {
		log.Fatal("Gemini setup error", err)
	}
	err = g.API.Endpoints.SetRunning(exchange.RestSpot.String(), geminiSandboxAPIURL)
	if err != nil {
		log.Fatalf("endpoint setting failed. key: %s, val: %s", exchange.RestSpot.String(), geminiSandboxAPIURL)
	}
	log.Printf(sharedtestvalues.LiveTesting, g.Name)
	os.Exit(m.Run())
}
