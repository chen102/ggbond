package server

import (
	"github.com/chen102/ggbond/conn/routermanage"
	"github.com/chen102/ggbond/conn/store"
)

type IRouterManage interface {
	RegisterRoute(id int32, route routermanage.RouterHandle) error
	HandleMessage(routerid, connid, msgid int32, parameter []byte) error
}

func NewRouterManage(name string, store store.ITCPStore) IRouterManage {
	switch name {
	case "router":
		return routermanage.NewTCPRouterManager(store)
	}
	return nil
}
