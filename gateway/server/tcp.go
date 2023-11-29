package server

import (
	"bufio"
	"context"
	"errors"
	"ggbond/gateway/connmanage"
	"ggbond/gateway/routermanage"
	"io"
	"log"
	"math"
	"net"
	"sync"

	"ggbond/message"

	uuid "github.com/satori/go.uuid"
	"github.com/spaolacci/murmur3"
)

type TCPServer struct {
	connManager connmanage.IConnManage
	listener    net.Listener
	router      routermanage.IRouterManage
	stopChannel chan struct{}
}

var _ IServer = (*TCPServer)(nil)

func NewTCPServer(connManager connmanage.IConnManage, router routermanage.IRouterManage) *TCPServer {
	return &TCPServer{
		connManager: connManager,
		router:      router,
		stopChannel: make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	log.Println("TCP server started on :8080")
	go s.acceptConnections()
	return nil
}
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
	conn := connmanage.NewConn(tcpconn, GenerateConnID(), "tcp")
	//连接前的OOK
	if s.connManager.GetHook() != nil {
		if err := s.connManager.GetHook().BeforConn(conn); err != nil {
			return err
		}
	}
	if err := s.connManager.AddConn(conn); err != nil {
		return err
	}
	go s.tcpreader(ctx, &wg, conn)
	go tcpwrite(ctx, &wg, conn)
	<-s.stopChannel
	cancel()
	wg.Wait()
	return s.connManager.RemoveConn(conn)
}
func GenerateConnID() int32 {
	//UUIDV4 HASH
	hasher := murmur3.New32()
	_, _ = hasher.Write([]byte(uuid.NewV4().String()))
	return int32(hasher.Sum32() % math.MaxInt32)
}
func (s *TCPServer) tcpreader(ctx context.Context, wg *sync.WaitGroup, conn connmanage.IConn) {
	reader := bufio.NewReader(conn.GetReader())
	wg.Add(1)
	defer wg.Done()
	log.Printf("conn %d tcpreader start...", conn.GetConnID())
	for {
		select {
		case <-ctx.Done():
			log.Printf("conn %d tcpreader done", conn.GetConnID())
			return
		default:
			if !conn.CheckHealth() {
				log.Printf("conn %d tcpreader done", conn.GetConnID())
				return
			}
			msg := message.NewMessage("tcp")
			if err := msg.ReadAndUnpack(reader); errors.Is(err, io.EOF) {
				log.Printf("conn %d tcpreader done", conn.GetConnID())
				return
			}
			log.Println("read from conn:", string(msg.GetBody()), msg.GetMessageID())
			if msg.GetRouteID() == 1 {
				conn.UpdateLastActiveTime()
				continue
			}
			if err := s.router.HandleMessage(msg.GetRouteID(), msg.GetBody()); err != nil {
				log.Println("route error:", err, "msg:", msg)
			}

		}
	}
}
func tcpwrite(ctx context.Context, wg *sync.WaitGroup, conn connmanage.IConn) {
	writer := bufio.NewWriter(conn.GetSender())
	wg.Add(1)
	defer wg.Done()
	log.Printf("conn %d tcpwrite start...", conn.GetConnID())
	for {
		select {
		case <-ctx.Done():
			log.Printf("conn %d tcpwrite done", conn.GetConnID())
			return
		case msg := <-conn.GetMessageChan():
			// log.Println("write to conn...")
			msg.PackAndWrite(writer)
			// log.Println("write to conn:", msg)
		}
	}
}
