package utils

import (
	"encoding/base64"
	"fmt"
	"time"
)

func GetShortCode() string {
	ts := time.Now().UnixNano()
	tsBytes := []byte(fmt.Sprintf("%d", ts))
	key := base64.StdEncoding.EncodeToString(tsBytes)
	key = key[:len(key)-2]

	return key[16:]
}
