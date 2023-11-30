package connmanage

import (
	"context"
	"io"
	"net"

	"github.com/chen102/ggbond/message"
	"github.com/chen102/ggbond/store"
)

// IConn 接口定义了连接的基本操作。
type IConn interface {
	GetConnType() (string, error)
	GetConnID() int32
	GetConn() (interface{}, error)
	CheckHealth(timeout int64) bool
	Close(err error) error
	WaitForClosed() chan error
	SignalClose(err error)
	GetSender() io.Writer
	GetReader() io.Reader
	UpdateLastActiveTime()
	SendMessage(message.IMessage) error
	GetMessageChan() chan message.IMessage
}
type Hook interface {
	BeforConn(IConn) error
	AfterConn(IConn) error
	CloseConn(IConn) error
}

// IConnManage 接口用于管理连接。
type IConnManage interface {
	AddConn(conn IConn) error
	RemoveConn(conn IConn, err error) error
	FindConn(id int32) (IConn, error)
	CheckHealths(context.Context)
	SetHook(Hook)
	GetHook() Hook
	GetAllConn() map[int32]IConn
	GetOutTimeOption(string) int64
}

// NewConnManage 创建一个新的连接管理器。
func NewConnManage(conntype string, store store.IStore, hook Hook) IConnManage {
	switch conntype {
	case "tcp":
		return NewTCPConnManager(store, hook)
	}
	return nil
}

// NewConn 创建一个新的连接。
func NewConn(conn net.Conn, connID int32, conntype string) IConn {
	switch conntype {
	case "tcp":
		return NewTCPConn(conn, connID, conntype)
	}
	return nil
}
