package connect

//使用epoll
import (
	"io"
	"log"
	"time"
)

type AsyncTcpConn struct {
	fd               int
	connID           int32
	connType         string
	lastactivatetime int64
	r                io.Reader
	w                io.Writer
	sendChan         chan IMessage
	close            chan error
}

func NewAsyncTcpConn(fd int, connID int32, connType string) *AsyncTcpConn {
	return &AsyncTcpConn{}
}
func (t *AsyncTcpConn) ConnType() (string, error) {
	return t.connType, nil
}
func (t *AsyncTcpConn) ConnID() int32 {
	return t.connID
}
func (t *AsyncTcpConn) Conn() (interface{}, error) {
	return t.fd, nil
}
func (t *AsyncTcpConn) CheckHealth(timeout int64) bool {
	log.Println("check health:", t.connID)
	return time.Now().Unix()-t.lastactivatetime < timeout
}
func (t *AsyncTcpConn) Close(err error) error {
	log.Println("连接关闭:", t.connID)
	return nil
}
func (t *AsyncTcpConn) WaitForClosed() chan error {
	return t.close
}
func (t *AsyncTcpConn) SignalClose(err error) {
	t.close <- err
}
func (t *AsyncTcpConn) Sender() io.Writer {
	return t.w
}
func (t *AsyncTcpConn) Reader() io.Reader {
	return t.r
}
func (t *AsyncTcpConn) UpdateLastActiveTime() {
	t.lastactivatetime = time.Now().Unix()
}
func (t *AsyncTcpConn) SendMessage(msg IMessage) error {
	t.sendChan <- msg
	return nil
}
func (t *AsyncTcpConn) MessageChan() chan IMessage {
	return t.sendChan
}
