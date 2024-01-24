package router

import (
	"fmt"
	"strconv"

	"github.com/chen102/ggbond/conn/connect"
	"github.com/chen102/ggbond/conn/routermanage"
	"github.com/chen102/ggbond/conn/server"
)

type SystemService struct {
	server.ITCPConnManage
}

const (
	PING           = 1
	ACTIVESHUTDOWN = 10
)

//初始化系统服务
func NewSystemService(connmanager server.ITCPConnManage) *SystemService {
	return &SystemService{
		connmanager,
	}
}

//路由装载器
func (b *SystemService) Handles() map[int32]routermanage.RouterHandle {
	return map[int32]routermanage.RouterHandle{
		PING:           b.Ping(),
		ACTIVESHUTDOWN: b.ActiveShutdown(),
	}
}
func (b *SystemService) Ping() routermanage.RouterHandle {
	return func(msgid, connid int32, parameter []byte) error {
		conn, err := b.FindConn(connid)
		if err != nil {
			return fmt.Errorf("PING Router Error:%w,RouterId:%d", err, PING)
		}
		msg := connect.NewMessage("tcp")
		if err := msg.Write([]byte(strconv.Itoa(int(msgid))+":PONG"), server.GenerateConnID(), PING); err != nil {
			return fmt.Errorf("PING Router Error:%w,RouterId:%d", err, PING)
		}
		if err := conn.SendMessage(msg); err != nil {
			return fmt.Errorf("PING Router Error:%w,RouterId:%d", err, PING)
		}
		conn.UpdateLastActiveTime()
		return nil
	}
}
func (b *SystemService) ActiveShutdown() routermanage.RouterHandle {
	return func(msgid, connid int32, parameter []byte) error {
		conn, err := b.FindConn(connid)
		if err != nil {
			return fmt.Errorf("ActiveShutdown Router Error:%w,RouterId:%d", err, ACTIVESHUTDOWN)
		}
		msg := connect.NewMessage("tcp")
		if err := msg.Write([]byte("ok"), server.GenerateConnID(), ACTIVESHUTDOWN); err != nil {
			return fmt.Errorf("ActiveShutdown Router Error:%w,RouterId:%d", err, ACTIVESHUTDOWN)
		}
		if err := conn.SendMessage(msg); err != nil {
			return fmt.Errorf("ActiveShutdown Router Error:%w,RouterId:%d", err, ACTIVESHUTDOWN)
		}
		return nil
	}
}
