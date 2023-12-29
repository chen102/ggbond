package connmanage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/chen102/ggbond/conn/store"
)

var ErrorTCPManager error = errors.New("tcp connmanager error")

type TCPConnManager struct {
	store.ITCPStore
	hook                Hook
	tcpnums             int32
	maximumConnection   int32 //最大连接数
	connectionTimedOut  int64 //连接超时时间
	transmissionTimeout int64 //传输超时时间
	explorationCycle    int64 //探测周期
	detectionTimeout    int64 //探测超时时间 每个连接探测超时时间，用次参数来监控连接是否正常
}

// NewTCPConnManager 创建一个tcp连接管理器
// ITCPStore:存储器实例,Hook:钩子函数,connManageroptions:连接管理器选项
func NewTCPConnManager(v store.ITCPStore, h Hook, opt ...ConnManagerOption) *TCPConnManager {
	m := &TCPConnManager{}
	m.ITCPStore = v
	m.hook = h
	if err := m.setoption(opt...); err != nil {
		panic(fmt.Errorf("%w:%w", ErrorTCPManager, err))
	}
	return m
}

// AddConn 添加一个连接
//ITCPConn :连接实例
func (m *TCPConnManager) AddConn(conn ITCPConn) error {
	if m.maximumConnection > 0 && m.tcpnums >= m.maximumConnection {
		return fmt.Errorf("%w: %s", ErrorTCPManager, "maximum connection")
	}
	connid := conn.GetConnID()
	_, err := m.Set(connid, conn)
	if err != nil {
		return fmt.Errorf("%w: %w ", ErrorTCPManager, err)
	}
	m.tcpnums++
	return nil
}

// RemoveConn 移除一个连接
//ITCPConn :连接实例
func (m *TCPConnManager) RemoveConn(conn ITCPConn, err error) error {
	if err := m.Del(conn.GetConnID()); err != nil {
		return fmt.Errorf("%w: %s", ErrorTCPManager, err)
	}
	m.tcpnums--
	return conn.Close(err)
}

// FindConn 查找一个连接
//connID 连接ID
func (m *TCPConnManager) FindConn(connID int32) (ITCPConn, error) {
	conn, err := m.Get(connID)
	if err != nil {
		return nil, err
	}
	store, ok := conn.(ITCPConn)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrorTCPManager, "connection is not a TCP connection")
	}
	return store, nil
}

//健康检查
func (m *TCPConnManager) CheckHealths(ctx context.Context) {
	close := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(m.explorationCycle)):
				m.RangeStroe(func(key, value interface{}) bool {
					conn := value.(ITCPConn)
					id := conn.GetConnID()
					log.Println("check health:", id)
					if !conn.CheckHealth(m.detectionTimeout) {
						log.Println("check health:", id, "failed")
						if err := m.RemoveConn(conn, errors.New("conn time out")); err != nil {
							log.Println("remove conn:", id, "failed")
						}
						m.Del(key.(int32))
						log.Println("remove conn:", id, "success")
					}
					return true
				})
			case <-ctx.Done():
				log.Println("check healths done")
				close <- struct{}{}
				break
			}
		}
	}()
	<-close
}
func (m *TCPConnManager) SetHook(hook Hook) {
	m.hook = hook
}
func (m *TCPConnManager) GetHook() Hook {
	return m.hook
}

// GetOutTimeOption 获取超时时间相关选项
func (m *TCPConnManager) GetOutTimeOption(name string) int64 {
	if name == "connectionTimedOut" {
		return m.connectionTimedOut
	}
	if name == "transmissionTimeout" {
		return m.transmissionTimeout
	}
	if name == "explorationCycle" {
		return m.explorationCycle
	}
	if name == "detectionTimeout" {
		return m.detectionTimeout
	}
	return 0
}

//GetAllConn 获取所有连接
func (c *TCPConnManager) GetAllConn() map[int32]ITCPConn {
	conns := make(map[int32]ITCPConn, c.tcpnums)
	c.RangeStroe(func(key, value interface{}) bool {
		conn := value.(ITCPConn)
		conns[key.(int32)] = conn
		return true
	})
	return conns
}
func (m *TCPConnManager) setoption(opt ...ConnManagerOption) error {
	var options connManageroptions
	baseerr := errors.New("option is not valid")
	for _, o := range opt {
		if err := o(&options); err != nil {
			panic(fmt.Errorf("apply option error:%w", err))
		}
	}
	var (
		maximumConnection   int32 = 10000
		connectionTimedOut  int64 = 2
		transmissionTimeout int64 = 2
		explorationCycle    int64 = 15
		detectionTimeout    int64 = 30
	)
	if options.maximumConnection != nil {
		if *options.maximumConnection < 0 || *options.maximumConnection > math.MaxInt32 {
			return fmt.Errorf("%w:maximumConnection is not valid", baseerr)
		}
		maximumConnection = *options.maximumConnection
	}
	if options.connectionTimedOut != nil {
		if *options.connectionTimedOut < 0 || *options.connectionTimedOut > math.MaxInt32 {
			fmt.Errorf("%w:connectionTimedOut is not valid", baseerr)
		}
		connectionTimedOut = *options.connectionTimedOut
	}
	if options.transmissionTimeout != nil {
		if *options.transmissionTimeout < 0 || *options.transmissionTimeout > math.MaxInt32 {
			fmt.Errorf("%w:transmissionTimeout is not valid", baseerr)
		}
	}
	if options.explorationCycle != nil {
		if *options.explorationCycle < 0 || *options.explorationCycle > math.MaxInt32 {
			fmt.Errorf("%w:explorationCycle is not valid", baseerr)
		}
		explorationCycle = *options.explorationCycle
	}
	if options.detectionTimeout != nil {
		if *options.detectionTimeout < 0 || *options.detectionTimeout > math.MaxInt32 {
			fmt.Errorf("%w:detectionTimeout is not valid", baseerr)
		}
		detectionTimeout = *options.detectionTimeout
	}
	m.maximumConnection = maximumConnection
	m.connectionTimedOut = connectionTimedOut
	m.transmissionTimeout = transmissionTimeout
	m.explorationCycle = explorationCycle
	m.detectionTimeout = detectionTimeout
	return nil
}
