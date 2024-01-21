package connect

import (
	"io"
	"net"
)

// ITCPConn 接口定义了连接的基本操作。
type ITCPConn interface {
	ConnType() (string, error)
	ConnID() int32
	Conn() (interface{}, error)
	CheckHealth(timeout int64) bool
	Close(err error) error
	WaitForClosed() chan error
	SignalClose(err error)
	Sender() io.Writer
	Reader() io.Reader
	UpdateLastActiveTime()
	SendMessage(IMessage) error
	MessageChan() chan IMessage
	Stat() ConnStat
	SetStat(ConnStat)
}

type Hook interface {
	BeforConn(ITCPConn) error
	AfterConn(ITCPConn) error
	CloseConn(ITCPConn) error
}

// NewConn 创建一个新的连接。
func NewConn(conn interface{}, connID int32, conntype string) ITCPConn {
	switch conntype {
	case "tcp":
		return NewTCPConn(conn.(net.Conn), connID, conntype)
	case "atcp":
		return NewAsyncTcpConn(conn.(int), connID, conntype)
	}
	return nil
}
