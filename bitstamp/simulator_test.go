package bitstamp

import (
	"testing"

	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/asset"
)

func TestSimulate(t *testing.T) {
	b := Bitstamp{}
	b.SetDefaults()
	o, err := b.FetchOrderbook(currency.NewPair(currency.BTC, currency.USD), asset.Spot)
	if err != nil {
		t.Error(err)
	}

	r := o.SimulateOrder(10000000, true)
	t.Log(r.Status)
	r = o.SimulateOrder(2171, false)
	t.Log(r.Status)
}
