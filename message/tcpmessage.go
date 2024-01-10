package message

import (
	"encoding/binary"
	"io"
)

const headerSize = 12 // 数据包长度、路由id、消息id各4字节
type flusher interface {
	Flush() error
}
type TCPMessage struct {
	body      []byte
	messageID int32
	routeID   int32
	length    int32
}

func (m *TCPMessage) PackAndWrite(w io.Writer) error {
	if err := m.pack(w); err != nil {
		return err
	}
	f, ok := w.(flusher)
	if !ok {
		return nil
	}
	return f.Flush()

}
func (m *TCPMessage) ReadAndUnpack(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, &m.length); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.routeID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.messageID); err != nil {
		return err
	}
	if m.length > 0 {
		m.body = make([]byte, m.length)
		if err := binary.Read(r, binary.BigEndian, &m.body); err != nil {
			return err
		}
	}
	// fmt.Println("read message:", m)
	return nil
}

func (m *TCPMessage) pack(w io.Writer) error {
	// 写入数据包长度
	if err := binary.Write(w, binary.BigEndian, m.length); err != nil {
		return err
	}
	// 写入路由ID
	if err := binary.Write(w, binary.BigEndian, m.routeID); err != nil {
		return err
	}
	// 写入消息ID
	if err := binary.Write(w, binary.BigEndian, m.messageID); err != nil {
		return err
	}
	// 写入消息体
	if err := binary.Write(w, binary.BigEndian, m.body); err != nil {
		return err
	}
	return nil
}

func (m *TCPMessage) Body() []byte {
	return m.body
}

func (m *TCPMessage) MessageID() int32 {
	return m.messageID
}

func (m *TCPMessage) RouteID() int32 {
	return m.routeID
}

func (m *TCPMessage) Length() int32 {
	return m.length
}
func (m *TCPMessage) Write(body []byte, messageID int32, routeID int32) error {
	m.body = body
	m.messageID = messageID
	m.routeID = routeID
	m.length = int32(len(body))
	return nil
}
