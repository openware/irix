package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/openware/openfinex/pkg/log"
	"github.com/openware/openfinex/pkg/msg"
)

type Client struct {
	conn *websocket.Conn

	key    string
	secret string

	done chan struct{}
	msgs chan *msg.Msg
}

// NewClient initializes new Client object.
func New(key, secret string) *Client {
	return &Client{
		key:    key,
		secret: secret,
		done:   make(chan struct{}),
		msgs:   make(chan *msg.Msg),
	}
}

func (c *Client) apikeyHeaders(header http.Header) http.Header {
	kid := c.key
	secret := c.secret
	nonce := fmt.Sprintf("%d", time.Now().Unix()*1000)
	data := nonce + kid

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))

	header.Add("X-Auth-Apikey", kid)
	header.Add("X-Auth-Nonce", nonce)
	header.Add("X-Auth-Signature", sha)

	return header
}

func (c *Client) Connect(url string) error {
	body := bytes.NewBuffer(nil)
	conn, resp, err := websocket.DefaultDialer.Dial(url, c.apikeyHeaders(http.Header{}))
	if resp.Body != nil {
		defer resp.Body.Close()
		io.Copy(body, resp.Body)
	}
	if err != nil {
		return fmt.Errorf("%w: %s", err, body.String())
	}
	resp.Body.Close()
	c.conn = conn

	return nil
}

func (c *Client) Listen() <-chan *msg.Msg {
	go func() {
		defer func() {
			close(c.done)
			close(c.msgs)
		}()

		for {
			_, m, err := c.conn.ReadMessage()
			if err != nil {
				log.Debug("websocket", "error on read message", err.Error())
				return
			}

			parsed, err := msg.Parse(m)
			if err != nil {
				log.Debug("websocket", "error on parse message", err.Error())
				continue
			}

			c.msgs <- parsed
		}
	}()

	return c.msgs
}

func (c *Client) Shutdown() error {
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}

	return c.conn.Close()
}

func (c *Client) Send(msg []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) SendMsg(msg *msg.Msg) error {
	b, err := msg.Encode()
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, b)
}
