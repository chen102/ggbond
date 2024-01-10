package server

import (
	"context"

	"github.com/chen102/ggbond/conn/connect"
	"github.com/chen102/ggbond/conn/connmanage"
	"github.com/chen102/ggbond/conn/store"
)

// IConnManage 接口用于管理连接。
type ITCPConnManage interface {
	AddConn(conn connect.ITCPConn) error
	RemoveConn(conn connect.ITCPConn, err error) error
	FindConn(id int32) (connect.ITCPConn, error)
	CheckHealths(context.Context)
	SetHook(connect.Hook)
	Hook() connect.Hook
	AllConn() map[int32]connect.ITCPConn
	OutTimeOption(string) int64
}

type IConnGroupMagage interface {
	AddGroup(g connmanage.GroupHook) error
	RemoveGroup(g connmanage.GroupHook) error
	Group(g connmanage.GroupHook) (map[int32]struct{}, error)
	AddConnToGroup(g connmanage.GroupHook, conn int32) error
	RemoveConnFromGroup(g connmanage.GroupHook, conn int32) error
	ClearGroup(g connmanage.GroupHook) error
}

// NewConnManage 创建一个新的连接管理器。
func NewConnManage(conntype string, store store.ITCPStore, hook connect.Hook, opt ...connmanage.ConnManagerOption) ITCPConnManage {
	switch conntype {
	case "tcp":
		return connmanage.NewTCPConn(store, hook, opt...)
	}
	return nil
}
func NewConnGroup() IConnGroupMagage {
	return connmanage.NewConnGroup()
}
