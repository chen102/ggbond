package main

import (
	"fmt"
	"ggbond/gateway/connmanage"
	"ggbond/gateway/routermanage"
	"ggbond/gateway/server"
	"ggbond/message"
	"ggbond/store"
	"log"
	"time"
)

type Hook struct {
}

func (c *Hook) BeforConn(conn connmanage.IConn) error {
	fmt.Println("马上连接")
	return nil
}
func (c *Hook) AfterConn(conn connmanage.IConn) error {
	fmt.Println("连接成功")
	return nil
}

const (
	_ = iota //提示上线
	ONLINE
	Boadcast
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
	router.RegisterRoute(Boadcast, func(b []byte) error {
		allconns := connmanager.GetAllConn()
		for _, v := range allconns {
			msg := message.NewMessage("tcp")
			msg.Write([]byte("已处理"+string(b)), server.GenerateConnID(), ONLINE)
			if err := v.SendMessage(msg); err != nil {
				panic(err)
			}
			log.Println("发送消息:", string(msg.GetBody()))
		}
		return nil
	})
	svc.Start()
	for {
		select {
		case <-time.After(time.Second * 20):
			if err := connmanager.CheckHealths(); err != nil {
				panic(err)
			}
		}
	}
}
