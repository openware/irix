package stats

import (
	"testing"

	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
)

const (
	testExchange = "OKEX"
)

func TestLenByPrice(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	Items = []Item{
		{
			Exchange:  testExchange,
			Pair:      p,
			AssetType: asset.Spot,
			Price:     1200,
			Volume:    5,
		},
	}

	if ByPrice.Len(Items) < 1 {
		t.Error("stats LenByPrice() length not correct.")
	}
}

func TestLessByPrice(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	Items = []Item{
		{
			Exchange:  "alphapoint",
			Pair:      p,
			AssetType: asset.Spot,
			Price:     1200,
			Volume:    5,
		},
		{
			Exchange:  "bitfinex",
			Pair:      p,
			AssetType: asset.Spot,
			Price:     1198,
			Volume:    20,
		},
	}

	if !ByPrice.Less(Items, 1, 0) {
		t.Error("stats LessByPrice() incorrect return.")
	}
	if ByPrice.Less(Items, 0, 1) {
		t.Error("stats LessByPrice() incorrect return.")
	}
}

func TestSwapByPrice(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	Items = []Item{
		{
			Exchange:  "bitstamp",
			Pair:      p,
			AssetType: asset.Spot,
			Price:     1324,
			Volume:    5,
		},
		{
			Exchange:  "bitfinex",
			Pair:      p,
			AssetType: asset.Spot,
			Price:     7863,
			Volume:    20,
		},
	}

	ByPrice.Swap(Items, 0, 1)
	if Items[0].Exchange != "bitfinex" || Items[1].Exchange != "bitstamp" {
		t.Error("stats SwapByPrice did not swap values.")
	}
}

func TestLenByVolume(t *testing.T) {
	if ByVolume.Len(Items) != 2 {
		t.Error("stats lenByVolume did not swap values.")
	}
}

func TestLessByVolume(t *testing.T) {
	if !ByVolume.Less(Items, 1, 0) {
		t.Error("stats LessByVolume() incorrect return.")
	}
	if ByVolume.Less(Items, 0, 1) {
		t.Error("stats LessByVolume() incorrect return.")
	}
}

func TestSwapByVolume(t *testing.T) {
	ByPrice.Swap(Items, 0, 1)

	if Items[1].Exchange != "bitfinex" || Items[0].Exchange != "bitstamp" {
		t.Error("stats SwapByVolume did not swap values.")
	}
}

func TestAdd(t *testing.T) {
	Items = Items[:0]
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	err = Add(testExchange, p, asset.Spot, 1200, 42)
	if err != nil {
		t.Fatal(err)
	}

	if len(Items) < 1 {
		t.Error("stats Add did not add exchange info.")
	}

	err = Add("", p, "", 0, 0)
	if err == nil {
		t.Fatal("error cannot be nil")
	}

	if len(Items) != 1 {
		t.Error("stats Add did not add exchange info.")
	}

	p.Base = currency.XBT
	err = Add(testExchange, p, asset.Spot, 1201, 43)
	if err != nil {
		t.Fatal(err)
	}

	if Items[1].Pair.String() != "XBTUSD" {
		t.Fatal("stats Add did not add exchange info.")
	}

	p, err = currency.NewPairFromStrings("ETH", "USDT")
	if err != nil {
		t.Fatal(err)
	}
	Add(testExchange, p, asset.Spot, 300, 1000)

	if Items[2].Pair.String() != "ETHUSD" {
		t.Fatal("stats Add did not add exchange info.")
	}
}

func TestAppend(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	Append("sillyexchange", p, asset.Spot, 1234, 45)
	if len(Items) < 2 {
		t.Error("stats Append did not add exchange values.")
	}

	Append("sillyexchange", p, asset.Spot, 1234, 45)
	if len(Items) == 3 {
		t.Error("stats Append added exchange values")
	}
}

func TestAlreadyExists(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	if !AlreadyExists(testExchange, p, asset.Spot, 1200, 42) {
		t.Error("stats AlreadyExists exchange does not exist.")
	}
	p.Base = currency.NewCode("dii")
	if AlreadyExists("bla", p, asset.Spot, 1234, 123) {
		t.Error("stats AlreadyExists found incorrect exchange.")
	}
}

func TestSortExchangesByVolume(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	topVolume := SortExchangesByVolume(p, asset.Spot, true)
	if topVolume[0].Exchange != "sillyexchange" {
		t.Error("stats SortExchangesByVolume incorrectly sorted values.")
	}

	topVolume = SortExchangesByVolume(p, asset.Spot, false)
	if topVolume[0].Exchange != testExchange {
		t.Error("stats SortExchangesByVolume incorrectly sorted values.")
	}
}

func TestSortExchangesByPrice(t *testing.T) {
	p, err := currency.NewPairFromStrings("BTC", "USD")
	if err != nil {
		t.Fatal(err)
	}
	topPrice := SortExchangesByPrice(p, asset.Spot, true)
	if topPrice[0].Exchange != "sillyexchange" {
		t.Error("stats SortExchangesByPrice incorrectly sorted values.")
	}

	topPrice = SortExchangesByPrice(p, asset.Spot, false)
	if topPrice[0].Exchange != testExchange {
		t.Error("stats SortExchangesByPrice incorrectly sorted values.")
	}
}
