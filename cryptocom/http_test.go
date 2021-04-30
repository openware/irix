package cryptocom

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type httpClientMock struct {
	mock.Mock
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestHttpClient_SendGet(t *testing.T) {
	testCases := []struct {
		method       string
		payload      *Request
		out interface{}
		mockResponse string
		mockResponseCode int
		shouldError bool
	}{
		{
			"GET",
			&Request{
				Method:    "something",
			},
			&InstrumentResult{},
			`{"code":10001, "message": "invalid code"}`,
			400,
			true,
		},
	}
	for _, c := range testCases {
		mocker := &httpClientMock{}
		resp := &http.Response{
			StatusCode: c.mockResponseCode,
			Body: ioutil.NopCloser(strings.NewReader(c.mockResponse)),
		}
		mocker.On("Do", mock.Anything).Once().Return(resp, nil)
		root := fmt.Sprintf("%s%s/%s", "https://", sandboxHost, apiVersion)
		cli := newHttpClient(mocker, root)
		res, err := cli.Send(c.method, c.payload, c.out)
		mocker.AssertExpectations(t)
		req := mocker.Calls[0].Arguments[0].(*http.Request)
		assert.Equal(t, fmt.Sprintf("/%s/%s", apiVersion, c.payload.Method), req.URL.Path)
		assert.Equal(t, "", req.URL.Query().Encode())
		assert.Equal(t, "GET", req.Method)
		if c.shouldError {
			assert.NotNil(t, err)
		} else if err == nil {
			assert.Equal(t, 0, res.Code)
			assert.NotNil(t, c.out)
		}
	}
}
