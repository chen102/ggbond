package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chen102/ggbond/gateway/connmanage"
	"github.com/chen102/ggbond/gateway/routermanage"
	"github.com/chen102/ggbond/gateway/server"
	"github.com/chen102/ggbond/message"
	"github.com/chen102/ggbond/store"
)

type Hook struct {
}

func (c *Hook) BeforConn(conn connmanage.IConn) error {
	fmt.Println("正在连接")
	return nil
}
func (c *Hook) AfterConn(conn connmanage.IConn) error {
	fmt.Println("连接成功")
	return nil
}
func (c *Hook) CloseConn(conn connmanage.IConn) error {
	fmt.Println("关闭连接")
	return nil
}

const (
	_ = iota //提示上线
	ONLINE
	PING
)

func main() {
	var (
		connmanager connmanage.IConnManage     = connmanage.NewConnManage("tcp", store.NewStoe(store.SYNCMAPSTORE), &Hook{})
		router      routermanage.IRouterManage = routermanage.NewRouterManage("router", store.NewStoe(store.SYNCMAPSTORE))
		svc         server.IServer             = server.NewTCPServer(connmanager, router)
	)
	router.RegisterRoute(ONLINE, func(b []byte) error {
		allconns := connmanager.GetAllConn()
		for _, v := range allconns {
			msg := message.NewMessage("tcp")
			msg.Write([]byte(string(b)+"上线了"), server.GenerateConnID(), ONLINE)
			if err := v.SendMessage(msg); err != nil {
				panic(err)
			}
			log.Println("发送消息:", string(msg.GetBody()))

		}
		return nil
	})
	router.RegisterRoute(PING, func(b []byte) error {
		allconns := connmanager.GetAllConn()
		for _, v := range allconns {
			msg := message.NewMessage("tcp")
			msg.Write([]byte("PONG"), server.GenerateConnID(), 5)
			if err := v.SendMessage(msg); err != nil {
				panic(err)
			}
			log.Println("发送消息:", string(msg.GetBody()))
		}
		return nil
	})
	svc.Start()
	connmanager.CheckHealths(context.Background())
}
