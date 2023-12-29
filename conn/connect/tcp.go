package connect

import (
	"io"
	"log"
	"net"
	"time"
)

type TCPConn struct {
	conn             net.Conn
	connID           int32
	connType         string
	lastactivatetime int64
	r                io.Reader
	w                io.Writer
	sendChan         chan IMessage
	close            chan error
}

//初始化一个TCP连接
func NewTCPConn(conn net.Conn, connID int32, connType string) *TCPConn {
	return &TCPConn{
		conn:             conn,
		connID:           connID,
		connType:         connType,
		lastactivatetime: time.Now().Unix(),
		r:                conn,
		w:                conn,
		sendChan:         make(chan IMessage, 100),
		close:            make(chan error),
	}
}

//获取TCP读者
func (c *TCPConn) GetSender() io.Writer {
	return c.w
}

//获取TCP写者
func (c *TCPConn) GetReader() io.Reader {
	return c.r
}

//获取连接类型
func (c *TCPConn) GetConnType() (string, error) {
	return c.connType, nil
}

//获取连接ID
func (c *TCPConn) GetConnID() int32 {
	return c.connID
}

//获取连接
func (c *TCPConn) GetConn() (interface{}, error) {
	return c.conn, nil
}

//检查连接是否健康 timeout:超时时间 单位秒
func (c *TCPConn) CheckHealth(timeout int64) bool {
	log.Println("check health:", c.connID)
	return time.Now().Unix()-c.lastactivatetime < timeout
}

//关闭连接
func (c *TCPConn) Close(err error) error {
	log.Println("连接关闭:", c.connID)
	return c.conn.Close()
}

//更新最后活跃时间
func (c *TCPConn) UpdateLastActiveTime() {
	c.lastactivatetime = time.Now().Unix()
}

//发送消息
func (c *TCPConn) SendMessage(msg IMessage) error {
	c.sendChan <- msg
	return nil
}

//获取消息通道
func (c *TCPConn) GetMessageChan() chan IMessage {
	return c.sendChan
}

//等待连接关闭
func (c *TCPConn) WaitForClosed() chan error {
	return c.close
}

//通知连接关闭
func (c *TCPConn) SignalClose(err error) {
	c.close <- err
}
