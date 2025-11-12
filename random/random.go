package random

import "crypto/rand"

const (
	alphanum             = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	digit                = "0123456789"
	digitLowercaseLetter = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// RandomString generate random string by specify chars.
// source: https://github.com/gogits/gogs/blob/9ee80e3e5426821f03a4e99fad34418f5c736413/modules/base/tool.go#L58
func RandomString(n int, alphabets ...byte) (string, error) {
	var bytes = make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		if len(alphabets) == 0 {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return string(bytes), nil
}

// 随机生成只包含数字的字符串
func RandomDigitString(n int) (string, error) {
	return RandomString(n, []byte(digit)...)
}

// 随机生成只包含数字和小写字母的字符串
func GetRandomDigitLowercaseLetterString(n int) (string, error) {
	return RandomString(n, []byte(digitLowercaseLetter)...)
}
