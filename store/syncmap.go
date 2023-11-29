package store

import (
	"errors"
	"fmt"
	"sync"
)

const SYNCMAPSTORE = "syncmap"

var ErrNotFound = errors.New("syncmap store:not found")

type SyncMapStore struct {
	sync.Map
}

var _ IStore = (*SyncMapStore)(nil)

func NewSyncMapStore() *SyncMapStore {
	return &SyncMapStore{}
}
func (s *SyncMapStore) Get(key int32) (interface{}, error) {
	res, ok := s.Load(key)
	if !ok {
		return nil, fmt.Errorf("%w:%d", ErrNotFound, key)
	}
	return res, nil
}
func (s *SyncMapStore) Set(key int32, value interface{}) (int32, error) {
	s.Store(key, value)
	return key, nil
}
func (s *SyncMapStore) Del(key int32) error {
	s.Delete(key)
	return nil
}
func (s *SyncMapStore) Exist(key int32) bool {
	_, ok := s.Load(key)
	return ok
}
func (s *SyncMapStore) RangeStroe(f func(key, value interface{}) bool) {
	s.Range(f)
}
