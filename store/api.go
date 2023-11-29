package store

// IStore 接口定义了基本的数据存储功能。
type IStore interface {
	Get(key int32) (interface{}, error)
	Set(key int32, value interface{}) (int32, error)
	Del(key int32) error
	Exist(key int32) bool
	RangeStroe(f func(key, value interface{}) bool)
}

func NewStoe(name string) IStore {
	switch name {
	case SYNCMAPSTORE:
		return NewSyncMapStore()
	}
	return nil
}
