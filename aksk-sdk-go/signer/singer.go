package signer

import "net/http"

type Result struct {
	Signature string
	Timestamp string
	Nonce     string
}

type Signer interface {
	Sign(req *http.Request, body []byte, secret string) Result
}
