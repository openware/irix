package cryptocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type connectionMock struct {
	Buffer   *bytes.Buffer
	Response chan *bytes.Buffer
}

func (c *Client) connectMock(privateResponse chan *bytes.Buffer, publicResponse chan *bytes.Buffer, privateWriter, publicWriter *bytes.Buffer) {
	c.publicConn = Connection{
		Transport: &connectionMock{Buffer: publicWriter, Response: publicResponse},
		IsPrivate: false,
		Mutex:     &sync.Mutex{},
	}
	c.privateConn = Connection{
		Transport: &connectionMock{Buffer: privateWriter, Response: privateResponse},
		IsPrivate: true,
		Mutex:     &sync.Mutex{},
	}
}

func (cm *connectionMock) ReadMessage() (int, []byte, error) {
	response := <-cm.Response
	return 1, response.Bytes(), nil
}

func (cm *connectionMock) WriteMessage(messageType int, data []byte) error {
	cm.Buffer.Write(data)
	return nil
}

func (cm *connectionMock) Close() error {
	cm.Buffer.Reset()
	return nil
}

func TestConnectionWrite(t *testing.T) {
	client := New("test", "test", "test", "test")
	// publicBuffer := bytes.NewBuffer(nil)
	privateBuffer := bytes.NewBuffer(nil)

	t.Run("auth", func(t *testing.T) {
		client.privateConn = Connection{Transport: &connectionMock{Buffer: privateBuffer}, Mutex: &sync.Mutex{}}

		req := client.AuthRequest()
		client.sendPrivateRequest(req)

		expected := fmt.Sprintf("{\"api_key\":\"test\",\"id\":1,\"method\":\"public/auth\",\"nonce\":\"%v\",\"sig\":\"%v\"}", req.Nonce, req.Signature)
		assert.Equal(t, expected, privateBuffer.String())
	})
}

func TestConnectionRead(t *testing.T) {
	// prepare mock
	client := New("", "", "test", "test")
	privateResponse := make(chan *bytes.Buffer, 1)
	publicResponse := make(chan *bytes.Buffer, 1)
	client.connectMock(privateResponse, publicResponse, bytes.NewBuffer(nil), bytes.NewBuffer(nil))

	// expectations
	expectStr := `{"id":12,"method":"public/auth","code":10002,"message":"UNAUTHORIZED"}`
	var expectedResponse Response
	err := json.Unmarshal([]byte(expectStr), &expectedResponse)
	if err != nil {
		fmt.Println("error on parse expected message")
	}

	// mocked responses
	privateResponse <- bytes.NewBufferString(expectStr)
	publicResponse <- bytes.NewBufferString(expectStr)

	// Running client
	msgs := client.Listen()

	// assertion
	assert.Equal(t, expectedResponse, <-msgs)
}
