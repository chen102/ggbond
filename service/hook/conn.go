package hook

import (
	"fmt"

	"github.com/chen102/ggbond/conn/connmanage"
)

type Hook struct {
}

func (c *Hook) BeforConn(conn connmanage.ITCPConn) error {
	fmt.Println("正在连接")
	return nil
}
func (c *Hook) AfterConn(conn connmanage.ITCPConn) error {
	fmt.Println("连接成功")
	return nil
}
func (c *Hook) CloseConn(conn connmanage.ITCPConn) error {
	fmt.Println("关闭连接")
	return nil
}
