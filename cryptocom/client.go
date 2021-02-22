package cryptocom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	userEndpoint   = "/v2/user"
	marketEndpoint = "/v2/market"
)

type LogFunc func(format string, args ...interface{})

func defaultLogFunc(format string, args ...interface{}) {
	log.Printf(format, args...)
}

type Transport interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	Close() error
}

func (c *Connection) Type() string {
	if c.IsPrivate {
		return "private"
	}

	return "public"
}

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
}

type Connection struct {
	Endpoint  string
	IsPrivate bool
	Transport
	sync.Mutex
}

type Client struct {
	publicConn    *Connection
	privateConn   *Connection
	isTerminating bool
	wsRootURL     string
	restRootURL   string
	key           string
	secret        string
	privateSubs   []string
	publicSubs    []string
	httpClient    HTTPClient
	outbox        chan Response
	LogFunc       LogFunc
	wg            sync.WaitGroup
}

// New returns a pointer of Client struct
func New(wsRootURL, restRootURL, key, secret string) *Client {
	return &Client{
		key:           key,
		secret:        secret,
		wsRootURL:     wsRootURL,
		restRootURL:   restRootURL,
		outbox:        make(chan Response),
		privateSubs:   make([]string, 0),
		publicSubs:    make([]string, 0),
		httpClient:    &http.Client{},
		LogFunc:       defaultLogFunc,
		isTerminating: false,
	}
}

// Connect instansiate WS Connections
func (c *Client) Connect() error {
	publicWsEndpoint := c.wsRootURL + marketEndpoint
	privateWsEndpoint := c.wsRootURL + userEndpoint

	cnx, err := c.createConnection(publicWsEndpoint, false)
	if err != nil {
		return err
	}
	c.publicConn = &cnx

	cnx, err = c.createConnection(privateWsEndpoint, true)
	if err != nil {
		return err
	}
	c.privateConn = &cnx

	time.Sleep(3 * time.Second) // Cryptocom requires this sleep.
	c.authenticate()

	return nil
}

func (c *Client) Listen() <-chan Response {
	go c.readConnection(c.publicConn)
	go c.readConnection(c.privateConn)
	return c.outbox
}

func (c *Client) Shutdown() {
	c.isTerminating = true
	c.privateConn.Close()
	c.publicConn.Close()
	c.wg.Wait()
	close(c.outbox)
}

func (c *Client) createConnection(endpoint string, isPrivate bool) (Connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, http.Header{})
	if err != nil {
		return Connection{}, err
	}

	return Connection{Endpoint: endpoint, IsPrivate: isPrivate, Transport: conn}, nil
}

func (c *Client) readConnection(cnx *Connection) {
	defer c.wg.Done()
	c.wg.Add(1)

	c.LogFunc("Start listening connection ... %s", cnx.Endpoint)
	for {
		_, m, err := cnx.ReadMessage()
		if err != nil {
			c.LogFunc("error on read message in %s cnx\nError message: %s\n", cnx.Type(), err.Error())
			if c.isTerminating {
				c.LogFunc("Stop reading from %s cnx. Connection closed\n", cnx.Type())
				return
			}
			for {
				conn, _, err := websocket.DefaultDialer.Dial(cnx.Endpoint, http.Header{})
				if err != nil {
					c.LogFunc("Reconnection error in %s cnx\n Error message: %s\n", cnx.Type(), err.Error())
					time.Sleep(1 * time.Second)
					continue
				}

				time.Sleep(3 * time.Second) // Cryptocom requires this sleep
				newCnx := Connection{Endpoint: cnx.Endpoint, IsPrivate: cnx.IsPrivate, Transport: conn}

				if newCnx.IsPrivate {
					c.privateConn = &newCnx
					c.authenticate()
					if len(c.privateSubs) > 0 {
						c.subscribePrivateChannels(c.privateSubs, false)
					}
				} else {
					c.publicConn = &newCnx
					if len(c.publicSubs) > 0 {
						c.subscribePublicChannels(c.publicSubs, false)
					}
				}

				cnx.Close()
				cnx = &newCnx
				break
			}

			continue
		}

		c.LogFunc("Received [%s]: %s\n", cnx.Type(), string(m))

		var parsed Response
		err = json.Unmarshal(m, &parsed)
		if err != nil {
			c.LogFunc("error on JSON.Unmarshal")
			continue
		}

		if parsed.Method == "public/heartbeat" {
			c.respondHeartBeat(cnx.IsPrivate, parsed.Id)
			continue
		}

		c.outbox <- parsed
	}
}

func (c *Client) generateSignature(r *Request) {
	secret := c.secret
	var parameters []string

	for key, v := range r.Params {
		parameters = append(parameters, key+v.(string))
	}

	sort.Strings(parameters)

	concatenedParams := strings.Join(parameters, "")

	data := r.Method + strconv.Itoa(r.Id) + r.ApiKey + concatenedParams + r.Nonce
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))
	r.Signature = sha
}

func (c *Client) authenticate() {
	r := c.AuthRequest()
	c.sendPrivateRequest(r)
}

func (c *Client) subscribePrivateChannels(channels []string, record bool) error {
	r := c.subscribeRequest(channels)
	err := c.sendPrivateRequest(r)

	if err != nil && record {
		c.privateSubs = append(c.privateSubs, channels...)
	}

	return err
}

func (c *Client) subscribePublicChannels(channels []string, record bool) error {
	r := c.subscribeRequest(channels)
	err := c.sendPublicRequest(r)

	if err != nil && record {
		c.publicSubs = append(c.publicSubs, channels...)
	}

	return err
}

func (c *Client) sendPrivateRequest(r *Request) error {
	b, err := r.Encode()
	if err != nil {
		return err
	}

	c.LogFunc("Sending private: %s\n", string(b))

	c.privateConn.Lock()
	response := c.privateConn.WriteMessage(websocket.TextMessage, b)
	c.privateConn.Unlock()

	return response
}

func (c *Client) sendPublicRequest(r *Request) error {
	b, err := r.Encode()
	if err != nil {
		return err
	}

	c.LogFunc("Sending public: %s\n", string(b))

	c.publicConn.Lock()
	response := c.publicConn.WriteMessage(websocket.TextMessage, b)
	c.publicConn.Unlock()

	return response
}
