package connmanage

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/chen102/ggbond/message"
)

type TCPConn struct {
	conn             net.Conn
	connID           int32
	connType         string
	lastactivatetime int64
	timeout          int64
	r                io.Reader
	w                io.Writer
	sendChan         chan message.IMessage
}

var _ IConn = (*TCPConn)(nil)

func NewTCPConn(conn net.Conn, connID int32, connType string) *TCPConn {
	return &TCPConn{
		conn:             conn,
		connID:           connID,
		connType:         connType,
		timeout:          30,
		lastactivatetime: time.Now().Unix(),
		r:                conn,
		w:                conn,
		sendChan:         make(chan message.IMessage, 100),
	}
}
func (c *TCPConn) GetSender() io.Writer {
	return c.w
}
func (c *TCPConn) GetReader() io.Reader {
	return c.r
}
func (c *TCPConn) GetConnType() (string, error) {
	return c.connType, nil
}

func (c *TCPConn) GetConnID() int32 {
	return c.connID
}

func (c *TCPConn) GetConn() (interface{}, error) {
	return c.conn, nil
}
func (c *TCPConn) CheckHealth() bool {
	log.Println("check health:", c.connID)
	return time.Now().Unix()-c.lastactivatetime < c.timeout
}
func (c *TCPConn) Close() error {
	return c.conn.Close()
}

func (c *TCPConn) UpdateLastActiveTime() {
	c.lastactivatetime = time.Now().Unix()
}

func (c *TCPConn) SendMessage(msg message.IMessage) error {
	c.sendChan <- msg
	return nil
}

func (c *TCPConn) GetMessageChan() chan message.IMessage {
	return c.sendChan
}
