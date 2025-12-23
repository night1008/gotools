package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/night1008/gotools/aksk-sdk-go/internal/utils"
)

type HmacSha256 struct{}

func NewHmacSha256() *HmacSha256 {
	return &HmacSha256{}
}

func (s *HmacSha256) Sign(req *http.Request, body []byte, sk string) Result {
	ts := utils.NowUnix()
	nonce := utils.RandString(16)

	signStr := strings.Join([]string{
		req.Method,
		req.URL.Path,
		CanonicalQuery(req.URL.RawQuery),
		utils.SHA256Hex(body),
		ts,
		nonce,
	}, "\n")

	h := hmac.New(sha256.New, []byte(sk))
	h.Write([]byte(signStr))

	return Result{
		Signature: hex.EncodeToString(h.Sum(nil)),
		Timestamp: ts,
		Nonce:     nonce,
	}
}
