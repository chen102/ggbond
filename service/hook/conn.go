package hook

import (
	"fmt"

	"github.com/chen102/ggbond/conn/connect"
)

type ConnHook struct {
}

func (c *ConnHook) BeforConn(conn connect.ITCPConn) error {
	fmt.Println("正在连接")
	return nil
}
func (c *ConnHook) AfterConn(conn connect.ITCPConn) error {
	fmt.Println("连接成功")
	return nil
}
func (c *ConnHook) CloseConn(conn connect.ITCPConn) error {
	fmt.Println("关闭连接")
	return nil
}
