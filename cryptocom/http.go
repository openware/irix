package cryptocom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HttpTransport interface {
	Send(httpMethod string, request *Request, out interface{}) (RawResponse, error)
}

type HttpExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}
type httpClient struct {
	client HttpExecutor
	root string
}

func (h *httpClient) Send(verb string, request *Request, out interface{}) (res RawResponse, err error) {
	var req *http.Request
	switch verb {
	case "GET":
		req, _ = http.NewRequest(verb, fmt.Sprintf("%s/%s", h.root, request.Method), nil)
		req.Header.Add("Content-Type", "application/json")
		q := url.Values{}
		if request.Id > 0 {
			q.Set("id", fmt.Sprint(request.Id))
		}
		if request.Nonce > 0 {
			q.Set("nonce", fmt.Sprintf("%d", request.Nonce))
		}
		for k, v := range request.Params {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
		break
	case "POST":
		payload, err1 := request.Encode()
		if err1 != nil {
			err = err1
			return
		}
		urlString := fmt.Sprintf("%s/%s", h.root, request.Method)
		req, _ = http.NewRequest(verb, urlString, bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		break
	}
	var rawMsg json.RawMessage
	httpRes, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer httpRes.Body.Close()
	if rawMsg, err = ioutil.ReadAll(httpRes.Body); err != nil {
		return
	}
	if err = json.Unmarshal(rawMsg, &res); err != nil {
		return
	}
	if res.Code > 0 {
		err = fmt.Errorf("error call at %s code: %d. reason: %s", request.Method, res.Code, res.Message)
		return
	}
	if out != nil {
		err = json.Unmarshal(rawMsg, out)
	}
	return
}

func newHttpClient(client HttpExecutor, root string) HttpTransport {
	return &httpClient{client, root}
}

func defaultHttpClient(root string) HttpTransport {
	return newHttpClient(http.DefaultClient, root)
}
