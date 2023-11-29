package main

import (
	"errors"
	"fmt"
	"ggbond/message"
	"io"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {

	//tcp连接8080
	conn, err := net.Dial("tcp", ":8080")
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
	name := os.Args[1]
	var id int32
	SendMsg(conn, 1, id, []byte(name))

	go func() {
		for {
			select {
			case <-time.After(time.Microsecond * 1000):
				id++
				SendMsg(conn, 2, id, []byte(name+"hello"))
			}
		}
	}()
	for {
		fmt.Println(string(RevMsg(conn).GetBody()))
	}

}
func SendMsg(conn net.Conn, routeID, msgID int32, body []byte) error {
	msg := message.NewMessage("tcp")
	msg.Write(body, msgID, routeID)
	return msg.PackAndWrite(conn)
}
func RevMsg(conn net.Conn) message.IMessage {
	msg := message.NewMessage("tcp")
	if err := msg.ReadAndUnpack(conn); errors.Is(err, io.EOF) {
		panic(err)
	}
	return msg
}
