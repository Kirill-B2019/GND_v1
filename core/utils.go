package core

import (
	"crypto/sha256"
	"strings"
)

func AddPrefix(addr string) string {
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(addr, string(prefix)) {
			return addr
		}
	}
	return "GND_" + addr
}

func Checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}
