package main

import (
	"context"

	"github.com/chen102/ggbond/conn/routermanage"
	"github.com/chen102/ggbond/conn/server"
	"github.com/chen102/ggbond/conn/store"
	"github.com/chen102/ggbond/service/hook"
	"github.com/chen102/ggbond/service/router"
)

type RouterInstance interface {
	Handles() map[int32]routermanage.RouterHandle
}
type IServer interface {
	Start() error
	Stop() error
}

func main() {

	var (
		connmanager   server.ITCPConnManage = server.NewConnManage("tcp", store.NewTCPSyncMapStore(), &hook.Hook{})
		routermanager server.IRouterManage  = server.NewRouterManage("router", store.NewTCPSyncMapStore())
		connsvc       IServer               = server.NewTCPServer(connmanager, routermanager)
		systemsvc     RouterInstance        = router.NewSystemService(connmanager)
	)
	for id, handle := range systemsvc.Handles() {
		routermanager.RegisterRoute(id, handle)
	}
	connsvc.Start()
	connmanager.CheckHealths(context.Background())
}
