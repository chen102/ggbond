package connmanage

import (
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
	CheckHealth() bool
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
}

// IConnManage 接口用于管理连接。
type IConnManage interface {
	AddConn(conn IConn) error
	RemoveConn(conn IConn, reason string) error
	FindConn(id int32) (IConn, error)
	CheckHealths() error
	SetHook(Hook)
	GetHook() Hook
	GetAllConn() map[int32]IConn
}

func NewConnManage(conntype string, store store.IStore, hook Hook) IConnManage {
	switch conntype {
	case "tcp":
		return NewTCPConnManager(store, hook)
	}
	return nil
}
func NewConn(conn net.Conn, connID int32, conntype string) IConn {
	switch conntype {
	case "tcp":
		return NewTCPConn(conn, connID, conntype)
	}
	return nil
}
