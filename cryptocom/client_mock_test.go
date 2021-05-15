package cryptocom

import "github.com/stretchr/testify/mock"

type HttpMock struct {
	mock.Mock
}

func (h *HttpMock) Send(httpMethod string, request *Request, out interface{}) (RawResponse, error) {
	args := h.Called(httpMethod, request, out)
	return args.Get(0).(RawResponse), args.Error(1)
}

type WsMock struct {
	mock.Mock
}

func (w *WsMock) ReadMessage(out interface{}) (RawResponse, error) {
	args := w.Called(out)
	return args.Get(0).(RawResponse), args.Error(1)
}

func (w *WsMock) WriteMessage(request Request) error {
	return w.Called(request).Error(0)
}

func (w *WsMock) Close() error {
	return w.Called().Error(0)
}



