package server

import (
	"context"

	"github.com/chen102/ggbond/conn/connmanage"
	"github.com/chen102/ggbond/conn/store"
)

// IConnManage 接口用于管理连接。
type ITCPConnManage interface {
	AddConn(conn connmanage.ITCPConn) error
	RemoveConn(conn connmanage.ITCPConn, err error) error
	FindConn(id int32) (connmanage.ITCPConn, error)
	CheckHealths(context.Context)
	SetHook(connmanage.Hook)
	GetHook() connmanage.Hook
	GetAllConn() map[int32]connmanage.ITCPConn
	GetOutTimeOption(string) int64
}

// NewConnManage 创建一个新的连接管理器。
func NewConnManage(conntype string, store store.ITCPStore, hook connmanage.Hook, opt ...connmanage.ConnManagerOption) ITCPConnManage {
	switch conntype {
	case "tcp":
		return connmanage.NewTCPConnManager(store, hook, opt...)
	}
	return nil
}
