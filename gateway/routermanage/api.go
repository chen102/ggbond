package routermanage

import (
	"github.com/chen102/ggbond/store"
)

type RouterHandle func([]byte) error
type IRouterManage interface {
	RegisterRoute(id int32, route RouterHandle) error
	HandleMessage(id int32, parameter []byte) error
}

func NewRouterManage(name string, store store.IStore) IRouterManage {
	switch name {
	case "router":
		return NewTCPRouterManager(store)
	}
	return nil
}
