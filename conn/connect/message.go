package connect

import (
	"io"

	"github.com/chen102/ggbond/message"
)

type IMessage interface {
	PackAndWrite(w io.Writer) error
	ReadAndUnpack(io.Reader) error
	GetBody() []byte
	GetMessageID() int32
	GetRouteID() int32
	GetLength() int32
	Write(body []byte, messageID int32, routeID int32) error
}

func NewMessage(messagetype string) IMessage {
	switch messagetype {
	case "tcp":
		return &message.TCPMessage{}
	}
	return nil
}
