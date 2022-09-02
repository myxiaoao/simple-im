package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 现在用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播 channel
	Message chan string
}

// NewServer 创建一个 server 的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.Message <- sendMsg
}

// ListenMessage 监听 message 广播消息 channel 的 goroutine， 一但有消息就发送给全部在线 user
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		// 将 msg 发送给全部在线的 user
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// Handler 处理业务链接
func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)

	user.Online()

	// 监听用户是否活跃的 channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		bytes := make([]byte, 4096)
		for {
			n, err := conn.Read(bytes)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}

			// 提前用户消息，并去除 "\n"
			msg := string(bytes[:n-1])

			// 用户针对 msg 进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是活跃状态
			isLive <- true
		}
	}()

	// 当前 handler 阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不错任何处理，为了激活 select ，更新下面的定时器
		case <-time.After(time.Second * 600):
			// 已经超时
			// 将当前 user 强制关闭
			user.SendMsg("你被踢了.\n")
			user.Online()
			close(isLive)
			// 销毁用户资源
			close(user.C)
			// 关闭链接
			err := conn.Close()
			if err != nil {
				fmt.Println("server conn close err:", err)
			}

			// 退出当前 handler
			return // runtime.GoExit()
		}
	}
}

// Start 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
	}

	// close listen socket
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("net.listen err:", err)
		}
	}(listen)

	// 启动监听 message 的 goroutine
	go s.ListenMessage()

	for {
		// accept
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}

		// do handler
		go s.Handler(accept)
	}
}
