package message

import (
	"errors"
	"io"
	"sync"
)

type IMessage interface {
	PackAndWrite(w io.Writer) error
	ReadAndUnpack(io.Reader) error
	Body() []byte
	MessageID() int32
	RouteID() int32
	Length() int32
	Write(body []byte, messageID int32, routeID int32) error
	Reset()
}

// 消息对象池
type Pool struct {
	pool map[string]sync.Pool
}

func NewPool(msgtype ...string) *Pool {
	msgpool := &Pool{
		pool: make(map[string]sync.Pool),
	}
	for _, v := range msgtype {
		msgpool.pool[v] = sync.Pool{
			New: func() interface{} {
				if v == "tcp" {
					return &TCPMessage{}
				}
				return nil
			},
		}
	}
	return msgpool
}

func (p *Pool) Get(msgtype string) (IMessage, error) {
	data, ok := p.pool[msgtype]
	if !ok {
		return nil, errors.New("pool msgtype is null")
	}

	return data.Get().(IMessage), nil
}
func (p *Pool) Put(msgtype string, msg IMessage) error {
	data, ok := p.pool[msgtype]
	if !ok {
		return errors.New("pool msgtype is null")
	}
	msg.Reset()
	data.Put(msg)
	return nil
}
