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
	"time"

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
			if s.connManager.OutTimeOption("readwriteTimeout") != 0 {
				timeout := time.Duration(s.connManager.OutTimeOption("readwriteTimeout")) * time.Second
				if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
					log.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}
			if s.connManager.OutTimeOption("readTimeout") != 0 {
				timeout := time.Duration(s.connManager.OutTimeOption("readTimeout")) * time.Second
				if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
					log.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}
			if s.connManager.OutTimeOption("writeTimeout") != 0 {
				timeout := time.Duration(s.connManager.OutTimeOption("writeTimeout")) * time.Second
				if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
					log.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}

			//设置连接超时时间
			go s.handle(conn)
		}
	}
}
func (s *TCPServer) handle(tcpconn net.Conn) error {
	var wg sync.WaitGroup
	conn := connect.NewConn(tcpconn, GenerateConnID(), "tcp")
	timeout := time.Now().Add(time.Duration(s.connManager.OutTimeOption("connectionTimedOut")) * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.connManager.AddConn(conn); err != nil {
		return conn.Close(err)
	}
	if s.connManager.Hook() != nil {
		err := s.connManager.Hook().AfterConn(conn)
		if err != nil {
			log.Println("hook error:", err)
			return s.connManager.RemoveConn(conn, err)
		}
	}

	if time.Now().After(timeout) {
		log.Println("connect timeout...")
		return s.connManager.RemoveConn(conn, errors.New("connect timeout..."))
	}
	go s.tcpreader(ctx, &wg, conn, int(s.connManager.ReadBuffer()))
	go s.tcpwrite(ctx, &wg, conn, int(s.connManager.WriteBuffer()))
	for {
		select {
		case <-s.stopChannel:
			cancel()
			wg.Wait()
			return s.connManager.RemoveConn(conn, errors.New("server stop"))
		case err := <-conn.WaitForClosed(): //读写协程出错，或者正常关闭
			if err != nil {
				log.Println("conn closed:", err)
			}
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
func (s *TCPServer) tcpreader(ctx context.Context, wg *sync.WaitGroup, conn connect.ITCPConn, buffsize int) {
	reader := bufio.NewReaderSize(conn.Reader(), buffsize)
	wg.Add(1)
	if err := s.resetTimeOut(conn, "readwriteTimeout"); err != nil {
		return
	}
	if err := s.resetTimeOut(conn, "readTimeout"); err != nil {
		return
	}
	defer log.Printf("conn %d tcpreader done", conn.ConnID())
	defer wg.Done()
	log.Printf("conn %d tcpreader start...", conn.ConnID())
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := s.msgpool.Get("tcp")
			if err != nil {
				conn.SignalClose(fmt.Errorf("get msg err:%w", err))
				return
			}
			if err := msg.ReadAndUnpack(reader); errors.Is(err, io.EOF) {
				// log.Println("read eof")
				conn.SignalClose(fmt.Errorf("readandunpack error:%w", err))
				return
			} else if operr, ok := err.(net.Error); ok && operr.Timeout() { //若设置了读超时时间，读超时后关闭连接
				conn.SignalClose(fmt.Errorf("readandunpack error:%w", err))
				return
			}

			log.Println("read from conn:", string(msg.Body()), msg.MessageID())
			if err := s.router.HandleMessage(msg.RouteID(), conn.ConnID(), msg.MessageID(), msg.Body()); err != nil {
				log.Println("route error:", err, "msg:", msg)
			}
			if err := s.msgpool.Put("tcp", msg); err != nil {
				conn.SignalClose(fmt.Errorf("put msg err:%w", err))
				return
			}
			if err := s.resetTimeOut(conn, "readwriteTimeout"); err != nil {
				conn.SignalClose(fmt.Errorf("set readwriteTimeout err:%w", err))
				return
			}
			if err := s.resetTimeOut(conn, "readTimeout"); err != nil {
				conn.SignalClose(fmt.Errorf("set readTimeout err:%w", err))
				return
			}
		}
	}
}
func (s *TCPServer) tcpwrite(ctx context.Context, wg *sync.WaitGroup, conn connect.ITCPConn, buffsize int) {
	writer := bufio.NewWriterSize(conn.Sender(), buffsize)
	wg.Add(1)
	if err := s.resetTimeOut(conn, "readwriteTimeout"); err != nil {
		conn.SignalClose(fmt.Errorf("set readwriteTimeout err:%w", err))
		return
	}
	if err := s.resetTimeOut(conn, "writeTimeout"); err != nil {
		conn.SignalClose(fmt.Errorf("set writeTimeout err:%w", err))
		return
	}
	defer wg.Done()
	defer log.Printf("conn %d tcpwrite done", conn.ConnID())
	log.Printf("conn %d tcpwrite start...", conn.ConnID())
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-conn.MessageChan():
			// log.Println("write to conn...")
			if err := msg.PackAndWrite(writer); err != nil {
				conn.SignalClose(fmt.Errorf("packandwrite error:%w", err))
				return
			} else if operr, ok := err.(net.Error); ok && operr.Timeout() { //若设置了读超时时间，读超时后关闭连接
				conn.SignalClose(fmt.Errorf("packandwrite error:%w", err))
				return
			}
			if err := s.resetTimeOut(conn, "readwriteTimeout"); err != nil {
				conn.SignalClose(fmt.Errorf("set readwriteTimeout err:%w", err))
				return
			}
			if err := s.resetTimeOut(conn, "writeTimeout"); err != nil {
				conn.SignalClose(fmt.Errorf("set writeTimeout err:%w", err))
				return
			}
			// log.Println("write to conn:", msg)
		}
	}
}
func (s *TCPServer) resetTimeOut(conn connect.ITCPConn, timeouttype string) error {
	if s.connManager.OutTimeOption(timeouttype) != 0 {

		log.Println("set timeout:", timeouttype, s.connManager.OutTimeOption(timeouttype))
		if timeouttype == "readwriteTimeout" {
			if err := conn.SetDeadline(s.connManager.OutTimeOption(timeouttype)); err != nil {
				return err
			}
		} else if timeouttype == "readTimeout" {
			if err := conn.SetReadDeadline(s.connManager.OutTimeOption(timeouttype)); err != nil {
				return err
			}
		} else if timeouttype == "writeTimeout" {
			if err := conn.SetWriteDeadline(s.connManager.OutTimeOption(timeouttype)); err != nil {
				return err
			}
		}
	}
	return nil
}
