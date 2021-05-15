package cryptocom

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type mockBody struct {
	code int
	body []byte
}
type httpClientMock struct {
	mock.Mock
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
