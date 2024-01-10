package connect

import (
	"io"

	"github.com/chen102/ggbond/message"
)

type IMessage interface {
	PackAndWrite(w io.Writer) error
	ReadAndUnpack(io.Reader) error
	Body() []byte
	MessageID() int32
	RouteID() int32
	Length() int32
	Write(body []byte, messageID int32, routeID int32) error
}

func NewMessage(messagetype string) IMessage {
	switch messagetype {
	case "tcp":
		return &message.TCPMessage{}
	}
	return nil
}
