package hook

import (
	"fmt"

	"github.com/chen102/ggbond/conn/connect"
)

type ConnHook struct {
}

func (c *ConnHook) AfterConn(conn connect.ITCPConn) error {
	fmt.Println("连接成功")
	// time.Sleep(10 * time.Second)
	return nil
}
