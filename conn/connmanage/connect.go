package connmanage

import (
	"io"
	"net"

	"github.com/chen102/ggbond/conn/connect"
)

// ITCPConn 接口定义了连接的基本操作。
type ITCPConn interface {
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
	SendMessage(connect.IMessage) error
	GetMessageChan() chan connect.IMessage
}

type Hook interface {
	BeforConn(ITCPConn) error
	AfterConn(ITCPConn) error
	CloseConn(ITCPConn) error
}

// NewConn 创建一个新的连接。
func NewConn(conn net.Conn, connID int32, conntype string) ITCPConn {
	switch conntype {
	case "tcp":
		return connect.NewTCPConn(conn, connID, conntype)
	}
	return nil
}
