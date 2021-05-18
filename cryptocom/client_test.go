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

func TestConnectionRead(t *testing.T) {
	// prepare mock
	client := Default("test", "test", true)
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

func TestDefaultClient(t *testing.T) {
	key, secret := "a", "b"
	cl := Default(key, secret, false)
	assert.Equal(t, fmt.Sprintf("wss://%s/%s", streamHost, apiVersion), cl.wsRootURL)
	assert.Equal(t, fmt.Sprintf("https://%s/%s", host, apiVersion), cl.restRootURL)
	assert.Equal(t, key, cl.key)
	assert.Equal(t, secret, cl.secret)

	cl = Default(key, secret, true)
	assert.Equal(t, fmt.Sprintf("wss://%s/%s", sandboxStreamHost, apiVersion), cl.wsRootURL)
	assert.Equal(t, fmt.Sprintf("https://%s/%s", sandboxHost, apiVersion), cl.restRootURL)
	assert.Equal(t, key, cl.key)
	assert.Equal(t, secret, cl.secret)
}
