package connmanage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/chen102/ggbond/conn/connect"
	"github.com/chen102/ggbond/conn/store"
)

var ErrorTCPManager error = errors.New("tcp connmanager error")

type TCPConnManager struct {
	store.ITCPStore
	hook                connect.Hook
	tcpnums             int32
	maximumConnection   int32 //最大连接数
	connectionTimedOut  int64 //连接超时时间
	transmissionTimeout int64 //传输超时时间
	explorationCycle    int64 //探测周期
	detectionTimeout    int64 //探测超时时间 每个连接探测超时时间，用次参数来监控连接是否正常
	readwriteTimeout    int64 //读写超时时间
	readTimeout         int64 //读超时时间
	writeTimeout        int64 //写超时时间
	readbuffer          int32 //读缓冲区大小
	writebuffer         int32 //写缓冲区大小
}

// NewTCPConnManager 创建一个tcp连接管理器
// ITCPStore:存储器实例,Hook:钩子函数,connManageroptions:连接管理器选项
func NewTCPConn(v store.ITCPStore, h connect.Hook, opt ...ConnManagerOption) *TCPConnManager {
	m := &TCPConnManager{}
	m.ITCPStore = v
	m.hook = h
	if err := m.setoption(opt...); err != nil {
		panic(fmt.Errorf("%w:%w", ErrorTCPManager, err))
	}
	return m
}

// AddConn 添加一个连接
// ITCPConn :连接实例
func (m *TCPConnManager) AddConn(conn connect.ITCPConn) error {
	if m.maximumConnection > 0 && m.tcpnums >= m.maximumConnection {
		return fmt.Errorf("%w: %s", ErrorTCPManager, "maximum connection")
	}
	connid := conn.ConnID()
	_, err := m.Set(connid, conn)
	if err != nil {
		return fmt.Errorf("%w: %w ", ErrorTCPManager, err)
	}
	m.tcpnums++
	return nil
}

// RemoveConn 移除一个连接
// ITCPConn :连接实例
func (m *TCPConnManager) RemoveConn(conn connect.ITCPConn, err error) error {
	if err := m.Del(conn.ConnID()); err != nil {
		return fmt.Errorf("%w: %s", ErrorTCPManager, err)
	}
	m.tcpnums--
	return conn.Close(err)
}

// FindConn 查找一个连接
// connID 连接ID
func (m *TCPConnManager) FindConn(connID int32) (connect.ITCPConn, error) {
	conn, err := m.Get(connID)
	if err != nil {
		return nil, err
	}
	store, ok := conn.(connect.ITCPConn)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrorTCPManager, "connection is not a TCP connection")
	}
	return store, nil
}

// 健康检查
// 防止客户端因为网络原因管理器误删除连接:
// 连接共有三张状态 活动:2 超时:1 关闭：0 单个连接，检查失败，依次递减直到状态为-1 连接管理器删除连接
func (m *TCPConnManager) CheckHealths(ctx context.Context) {
	close := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(m.explorationCycle)):
				m.RangeStroe(func(key, value interface{}) bool {
					conn := value.(connect.ITCPConn)
					id := conn.ConnID()
					log.Println("check health:", id)
					stat := conn.Stat()
					if stat < 0 {
						log.Println("check health:", id, "failed")
						if err := m.RemoveConn(conn, errors.New("connect timeout")); err != nil {
							log.Println("remove conn:", id, "failed")
						}
						m.Del(key.(int32))
						log.Println("remove conn:", id, "success")
						return true
					}
					if !conn.CheckHealth(m.detectionTimeout) {
						conn.SetStat(stat - 1)
						return true
					}
					if stat < connect.ACTIVE {
						conn.SetStat(stat + 1)
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
func (m *TCPConnManager) SetHook(hook connect.Hook) {
	m.hook = hook
}
func (m *TCPConnManager) Hook() connect.Hook {
	return m.hook
}

// GetOutTimeOption 获取超时时间相关选项
func (m *TCPConnManager) OutTimeOption(name string) int64 {
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
	if name == "readwriteTimeout" {
		return m.readwriteTimeout
	}
	if name == "readTimeout" {
		return m.readTimeout
	}
	if name == "writeTimeout" {
		return m.writeTimeout
	}
	return 0
}
func (m *TCPConnManager) ReadBuffer() int32 {
	return m.readbuffer
}
func (m *TCPConnManager) WriteBuffer() int32 {
	return m.writebuffer
}

// GetAllConn 获取所有连接
func (c *TCPConnManager) AllConn() map[int32]connect.ITCPConn {
	conns := make(map[int32]connect.ITCPConn, c.tcpnums)
	c.RangeStroe(func(key, value interface{}) bool {
		conn := value.(connect.ITCPConn)
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
		readwriteTimeout    int64 = 0
		readTimeout         int64 = 6
		writeTimeout        int64 = 1
		readbuffer          int32 = 1024
		writebuffer         int32 = 1024
	)
	if options.maximumConnection != nil {
		if *options.maximumConnection < 0 || *options.maximumConnection > math.MaxInt32 {
			return fmt.Errorf("%w:maximumConnection is not valid", baseerr)
		}
		maximumConnection = *options.maximumConnection
	}
	if options.connectionTimedOut != nil {
		if *options.connectionTimedOut < 0 || *options.connectionTimedOut > math.MaxInt64 {
			fmt.Errorf("%w:connectionTimedOut is not valid", baseerr)
		}
		connectionTimedOut = *options.connectionTimedOut
	}
	if options.transmissionTimeout != nil {
		if *options.transmissionTimeout < 0 || *options.transmissionTimeout > math.MaxInt64 {
			fmt.Errorf("%w:transmissionTimeout is not valid", baseerr)
		}
	}
	if options.explorationCycle != nil {
		if *options.explorationCycle < 0 || *options.explorationCycle > math.MaxInt64 {
			fmt.Errorf("%w:explorationCycle is not valid", baseerr)
		}
		explorationCycle = *options.explorationCycle
	}
	if options.detectionTimeout != nil {
		if *options.detectionTimeout < 0 || *options.detectionTimeout > math.MaxInt64 {
			fmt.Errorf("%w:detectionTimeout is not valid", baseerr)
		}
		detectionTimeout = *options.detectionTimeout
	}
	if options.readwriteTimeout != nil {
		if *options.readwriteTimeout < 0 || *options.readwriteTimeout > math.MaxInt64 {
			fmt.Errorf("%w:readwriteTimeout is not valid", baseerr)
		}
		readwriteTimeout = *options.readwriteTimeout
	}
	if options.readTimeout != nil {
		if *options.readTimeout < 0 || *options.readTimeout > math.MaxInt64 {
			fmt.Errorf("%w:readTimeout is not valid", baseerr)
		}
		readTimeout = *options.readTimeout
	}
	if options.writeTimeout != nil {
		if *options.writeTimeout < 0 || *options.writeTimeout > math.MaxInt64 {
			fmt.Errorf("%w:writeTimeout is not valid", baseerr)
		}
		writeTimeout = *options.writeTimeout
	}
	if options.readbuffer != nil {
		if *options.readbuffer < 0 || *options.readbuffer > math.MaxInt32 {
			fmt.Errorf("%w:readbuffer is not valid", baseerr)
		}
		readbuffer = *options.readbuffer
	}
	if options.writebuffer != nil {
		if *options.writebuffer < 0 || *options.writebuffer > math.MaxInt32 {
			fmt.Errorf("%w:writebuffer is not valid", baseerr)
		}
		writebuffer = *options.writebuffer
	}

	m.maximumConnection = maximumConnection
	m.connectionTimedOut = connectionTimedOut
	m.transmissionTimeout = transmissionTimeout
	m.explorationCycle = explorationCycle
	m.detectionTimeout = detectionTimeout
	m.readwriteTimeout = readwriteTimeout
	m.readTimeout = readTimeout
	m.writeTimeout = writeTimeout
	m.readbuffer = readbuffer
	m.writebuffer = writebuffer

	return nil
}
