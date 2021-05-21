package cryptocom

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) ReadMessage() (int, []byte, error) {
	args := m.Called()
	return args.Get(0).(int), args.Get(1).([]byte), args.Error(2)
}

func (m *mockTransport) WriteMessage(i int, i2 []byte) error {
	return m.Called(i, i2).Error(0)
}

func (m *mockTransport) Close() error {
	return m.Called().Error(0)
}

func mockWsClient() (cli *Client, public *mockTransport, private *mockTransport) {
	cli = New("test", "test", "test", "test")
	public = &mockTransport{}
	private = &mockTransport{}
	cli.publicConn = Connection{
		IsPrivate: false,
		Transport: public,
		Mutex:     &sync.Mutex{},
	}
	cli.privateConn = Connection{
		IsPrivate: true,
		Transport: private,
		Mutex:     &sync.Mutex{},
	}
	return
}

func TestFormat(t *testing.T) {
	markets := []string{"ETH_BTC", "ETH_COV", "XRP_BTC"}
	expected := []string{"trade.ETH_BTC", "trade.ETH_COV", "trade.XRP_BTC"}

	result := format(markets, func(s string) string {
		return fmt.Sprintf("trade.%s", s)
	})

	assert.Equal(t, result, expected)
}
