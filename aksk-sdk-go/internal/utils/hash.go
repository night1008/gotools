// hash.go
package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256Hex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
