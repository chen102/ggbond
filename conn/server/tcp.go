package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"

	"github.com/chen102/ggbond/conn/connect"
	"github.com/chen102/ggbond/message"
	uuid "github.com/satori/go.uuid"
	"github.com/spaolacci/murmur3"
)

type TCPServer struct {
	connManager ITCPConnManage
	group       IConnGroupMagage
	listener    net.Listener
	router      IRouterManage
	stopChannel chan struct{}
	ip          string
	port        int64
	servername  string
	msgpool     *message.Pool
}

// NewTCPServer 创建一个tcp服务器
// IConnManage:连接管理器实例，IRouterManage:路由管理实例 ,ServerOption:服务器选项
func NewTCPServer(connManager ITCPConnManage, router IRouterManage, opt ...ServerOption) *TCPServer {

	var options serveroptions
	for _, o := range opt {
		if err := o(&options); err != nil {
			panic(fmt.Errorf("apply option error:%w", err))
		}
	}
	var (
		ip, servername string = "127.0.0.1", "server001"
		port           int64  = 8080
	)

	if options.ip != nil {
		//验证ip是否合法
		if net.ParseIP(*options.ip) == nil {
			panic("ip is not valid")
		}
		ip = *options.ip
	}
	if options.port != nil {
		if *options.port < 0 || *options.port > 65535 {
			panic("port is not valid")
		}
		port = *options.port
	}
	if options.servername != nil {
		if *options.servername == "" {
			panic("servername is not valid")
		}
		servername = *options.servername
	}

	return &TCPServer{
		connManager: connManager,
		group:       NewConnGroup(),
		router:      router,
		stopChannel: make(chan struct{}),
		ip:          ip,
		port:        port,
		servername:  servername,
		msgpool:     message.NewPool("tcp"),
	}
}

// 启动服务
func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return err
	}
	log.Println("TCP server started on " + fmt.Sprintf("%s:%d", s.ip, s.port))
	go s.acceptConnections()
	return nil
}

// 停止服务
func (s *TCPServer) Stop() error {
	close(s.stopChannel)
	if err := s.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (s *TCPServer) acceptConnections() {
	for {
		select {
		case <-s.stopChannel:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v\n", err)
				continue
			}
			go s.handle(conn)
		}
	}
}
func (s *TCPServer) handle(tcpconn net.Conn) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	conn := connect.NewConn(tcpconn, GenerateConnID(), "tcp")
	//连接前的OOK
	if s.connManager.Hook() != nil {
		if err := s.connManager.Hook().BeforConn(conn); err != nil {
			return err
		}
	}
	if err := s.connManager.AddConn(conn); err != nil {
		tcpconn.Close()
		return err
	}
	go s.tcpreader(ctx, &wg, conn)
	go tcpwrite(ctx, &wg, conn)
	if s.connManager.Hook() != nil {
		if err := s.connManager.Hook().AfterConn(conn); err != nil {
			return err
		}
	}
	defer func() {
		if s.connManager.Hook() != nil {
			s.connManager.Hook().CloseConn(conn)
		}
	}()
	for {
		select {
		case <-s.stopChannel:
			cancel()
			wg.Wait()
			return s.connManager.RemoveConn(conn, errors.New("server stop"))
		case err := <-conn.WaitForClosed(): //读写协程出错，或者正常关闭
			cancel()
			wg.Wait()
			return s.connManager.RemoveConn(conn, err)
		}
	}

}

// 生成UUIDV4的murmur3算法int32 hash值
func GenerateConnID() int32 {
	//UUIDV4 HASH
	hasher := murmur3.New32()
	_, _ = hasher.Write([]byte(uuid.NewV4().String()))
	return int32(hasher.Sum32() % math.MaxInt32)
}
func (s *TCPServer) tcpreader(ctx context.Context, wg *sync.WaitGroup, conn connect.ITCPConn) {
	reader := bufio.NewReader(conn.Reader())
	wg.Add(1)
	defer wg.Done()
	log.Printf("conn %d tcpreader start...", conn.ConnID())
	for {
		select {
		case <-ctx.Done():
			log.Printf("conn %d tcpreader done", conn.ConnID())
			return
		default:
			if !conn.CheckHealth(s.connManager.OutTimeOption("detectionTimeout")) {
				log.Printf("conn %d tcpreader done", conn.ConnID())
				conn.SignalClose(errors.New("conn time out"))
				return
			}
			msg, err := s.msgpool.Get("tcp")
			if err != nil {
				log.Printf("conn %d tcpreader done", conn.ConnID())
				conn.SignalClose(fmt.Errorf("get msg err:%w", err))
				return
			}
			if err := msg.ReadAndUnpack(reader); errors.Is(err, io.EOF) {
				log.Printf("conn %d tcpreader done", conn.ConnID())
				conn.SignalClose(errors.New("read eof"))
				return
			}
			log.Println("read from conn:", string(msg.Body()), msg.MessageID())
			// switch msg.GetRouteID() {
			// case 1:
			// 	conn.UpdateLastActiveTime()
			// 	continue
			// case 9:
			// 	log.Printf("user %s logout", string(msg.GetBody()))
			// 	conn.SignalClose(nil)
			// 	return
			// }
			if err := s.router.HandleMessage(msg.RouteID(), conn.ConnID(), msg.MessageID(), msg.Body()); err != nil {
				log.Println("route error:", err, "msg:", msg)
			}
			if err := s.msgpool.Put("tcp", msg); err != nil {
				log.Printf("conn %d tcpreader done", conn.ConnID())
				conn.SignalClose(fmt.Errorf("put msg err:%w", err))
				return
			}
		}
	}
}
func tcpwrite(ctx context.Context, wg *sync.WaitGroup, conn connect.ITCPConn) {
	writer := bufio.NewWriter(conn.Sender())
	wg.Add(1)
	defer wg.Done()
	log.Printf("conn %d tcpwrite start...", conn.ConnID())
	for {
		select {
		case <-ctx.Done():
			log.Printf("conn %d tcpwrite done", conn.ConnID())
			return
		case msg := <-conn.MessageChan():
			// log.Println("write to conn...")
			if err := msg.PackAndWrite(writer); err != nil {
				log.Printf("conn %d tcpwrite done", conn.ConnID())
				conn.SignalClose(errors.New("write eof"))
				return
			}
			// log.Println("write to conn:", msg)
		}
	}
}
