// time.go
package utils

import (
	"strconv"
	"time"
)

func NowUnix() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
