package connect

import (
	"io"
	"log"
	"net"
	"time"
)

type ConnStat int

const (
	CLOSE = iota
	TIMEOUT
	ACTIVE
)

type TCP struct {
	conn             net.Conn
	connID           int32
	connType         string
	lastactivatetime int64
	r                io.Reader
	w                io.Writer
	sendChan         chan IMessage
	close            chan error
	stat             ConnStat
}

// 初始化一个TCP连接
func NewTCPConn(conn net.Conn, connID int32, connType string) *TCP {
	return &TCP{
		conn:             conn,
		connID:           connID,
		connType:         connType,
		lastactivatetime: time.Now().Unix(),
		r:                conn,
		w:                conn,
		sendChan:         make(chan IMessage, 100),
		close:            make(chan error),
		stat:             ACTIVE,
	}
}

// 获取TCP读者
func (c *TCP) Sender() io.Writer {
	return c.w
}

// 获取TCP写者
func (c *TCP) Reader() io.Reader {
	return c.r
}

// 获取连接类型
func (c *TCP) ConnType() (string, error) {
	return c.connType, nil
}

// 获取连接ID
func (c *TCP) ConnID() int32 {
	return c.connID
}

// 获取连接
func (c *TCP) Conn() (interface{}, error) {
	return c.conn, nil
}

// 检查连接是否健康 timeout:超时时间 单位秒
func (c *TCP) CheckHealth(timeout int64) bool {
	log.Println("check health:", c.connID)
	return time.Now().Unix()-c.lastactivatetime < timeout
}

// 关闭连接
func (c *TCP) Close(err error) error {
	log.Println("连接关闭:", c.connID)
	return c.conn.Close()
}

// 更新最后活跃时间
func (c *TCP) UpdateLastActiveTime() {
	c.lastactivatetime = time.Now().Unix()
}

// 发送消息
func (c *TCP) SendMessage(msg IMessage) error {
	c.sendChan <- msg
	return nil
}

// 获取消息通道
func (c *TCP) MessageChan() chan IMessage {
	return c.sendChan
}

// 等待连接关闭
func (c *TCP) WaitForClosed() chan error {
	return c.close
}

// 通知连接关闭
func (c *TCP) SignalClose(err error) {
	c.close <- err
}
func (c *TCP) Stat() ConnStat {
	return c.stat
}
func (c *TCP) SetStat(stat ConnStat) {
	c.stat = stat
}
