package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/chen102/ggbond/conn/connect"
)

func main() {

	//tcp连接8080
	conn, err := net.Dial("tcp", ":8089")
	if err != nil {
		panic(err)
	}
	//捕获退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		conn.Close()
		os.Exit(1)
	}()
	//获取命令行
	var id int32
	go func() {
		for id < 10 {
			select {
			case <-time.After(time.Microsecond * 1000000):
				id++
				SendMsg(conn, 1, id, nil)
			}
		}
		SendMsg(conn, 10, 1000, nil)
	}()

	for {
		msg := RevMsg(conn)
		if string(msg.Body()) == "ok" {
			break
		}
		fmt.Println(string(msg.Body()))

	}

}
func SendMsg(conn net.Conn, routeID, msgID int32, body []byte) error {
	msg := connect.NewMessage("tcp")
	msg.Write(body, msgID, routeID)
	return msg.PackAndWrite(conn)
}
func RevMsg(conn net.Conn) connect.IMessage {
	msg := connect.NewMessage("tcp")
	if err := msg.ReadAndUnpack(conn); errors.Is(err, io.EOF) {
		panic(err)
	}
	return msg
}
