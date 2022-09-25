package xrand

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)


func Bytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Reader.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate rand bytes: %v", err))
	}
	return b
}


func String(n int) string {
	s := strings.ToValidUTF8(string(Bytes(n)), "_")
	s = strings.ReplaceAll(s, "\x00", "_")
	if len(s) > n {
		return s[:n]
	}
	if len(s) < n {
		
		extra := n - len(s)
		return s + strings.Repeat("=", extra)
	}
	return s
}


func Bool() bool {
	return Int(2) == 1
}


func Int(max int) int {
	x, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(fmt.Sprintf("failed to get random int: %v", err))
	}
	return int(x.Int64())
}
