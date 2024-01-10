package store

import "github.com/chen102/ggbond/store"

type ITCPStore interface {
	Get(key int32) (interface{}, error)
	Set(key int32, value interface{}) (int32, error)
	Del(key int32) error
	Exist(key int32) bool
	RangeStroe(f func(key, value interface{}) bool)
}

func NewTCPSyncMap() ITCPStore {
	return store.NewSyncMap()
}
