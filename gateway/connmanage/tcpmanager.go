package connmanage

import (
	"errors"
	"fmt"
	"log"

	"github.com/chen102/ggbond/store"
)

var ErrorTCPManager error = errors.New("tcp connmanager error")

type TCPConnManager struct {
	store.IStore
	hook    Hook
	TCPNums int32
}

var _ IConnManage = (*TCPConnManager)(nil)

func NewTCPConnManager(v store.IStore, h Hook) *TCPConnManager {
	return &TCPConnManager{v, h, 0}
}

func (m *TCPConnManager) AddConn(conn IConn) error {
	connid := conn.GetConnID()
	_, err := m.Set(connid, conn)
	if err != nil {
		return fmt.Errorf("%w: %w ", ErrorTCPManager, err)
	}
	m.TCPNums++
	return nil
}

func (m *TCPConnManager) RemoveConn(conn IConn, err error) error {
	if err := m.Del(conn.GetConnID()); err != nil {
		return fmt.Errorf("%w: %s", ErrorTCPManager, err)
	}
	m.TCPNums--
	return conn.Close(err)
}

func (m *TCPConnManager) FindConn(connID int32) (IConn, error) {
	conn, err := m.Get(connID)
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(IConn)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrorTCPManager, "connection is not a TCP connection")
	}
	return tcpConn, nil
}
func (m *TCPConnManager) CheckHealths() error {
	m.RangeStroe(func(key, value interface{}) bool {
		conn := value.(IConn)
		id := conn.GetConnID()
		log.Println("check health:", id)
		if !conn.CheckHealth() {
			log.Println("check health:", id, "failed")
			if err := m.RemoveConn(conn, errors.New("conn time out")); err != nil {
				log.Println("remove conn:", id, "failed")
			}
			m.Del(key.(int32))
			log.Println("remove conn:", id, "success")
		}
		return true
	})
	return nil
}
func (m *TCPConnManager) SetHook(hook Hook) {
	m.hook = hook
}
func (m *TCPConnManager) GetHook() Hook {
	return m.hook
}
func (c *TCPConnManager) GetAllConn() map[int32]IConn {
	conns := make(map[int32]IConn, c.TCPNums)
	c.RangeStroe(func(key, value interface{}) bool {
		conn := value.(IConn)
		conns[key.(int32)] = conn
		return true
	})
	return conns
}
