// rand.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func RandString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
