package client

import (
	"bytes"
	"io"
	"net/http"

	"github.com/night1008/gotools/aksk-sdk-go/signer"
	"github.com/night1008/gotools/aksk-sdk-go/transport"
)

type Client struct {
	ak       string
	sk       string
	endpoint string
	signer   signer.Signer
	httpCli  *http.Client
}

func New(ak, sk, endpoint string) *Client {
	c := &Client{
		ak:       ak,
		sk:       sk,
		endpoint: endpoint,
		signer:   signer.NewHmacSha256(),
		httpCli:  transport.DefaultHTTPClient(),
	}

	return c
}

func (c *Client) Do(req *http.Request, body []byte) (*http.Response, error) {
	sig := c.signer.Sign(req, body, c.sk)

	req.Header.Set("X-AccessKey-ID", c.ak)
	req.Header.Set("X-Timestamp", sig.Timestamp)
	req.Header.Set("X-Nonce", sig.Nonce)
	req.Header.Set("X-Signature", sig.Signature)
	req.Header.Set("Content-Type", "application/json")

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return c.httpCli.Do(req)
}
