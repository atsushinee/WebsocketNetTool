package bpool

import (
	"bytes"
	"sync"
)

var bpool sync.Pool



func Get() *bytes.Buffer {
	b := bpool.Get()
	if b == nil {
		return &bytes.Buffer{}
	}
	return b.(*bytes.Buffer)
}


func Put(b *bytes.Buffer) {
	b.Reset()
	bpool.Put(b)
}
