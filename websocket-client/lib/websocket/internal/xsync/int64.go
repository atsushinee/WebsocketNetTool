package xsync

import (
	"sync/atomic"
)


type Int64 struct {
	
	
	i atomic.Value
}


func (v *Int64) Load() int64 {
	i, _ := v.i.Load().(int64)
	return i
}


func (v *Int64) Store(i int64) {
	v.i.Store(i)
}
