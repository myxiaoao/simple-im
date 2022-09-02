package server

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个用户的 API
func NewUser(conn net.Conn, server *Server) *User {
	s := conn.RemoteAddr().String()

	user := &User{
		Name:   s,
		Addr:   s,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

// ListenMessage 监听当前 User channel 的方法，一但有消息，就直接发送给端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("write message err:", err)
		}
	}
}

// Online 用户上线业务
func (u *User) Online() {
	// 用户上线，将用户加入到 online map 中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}

// Offline 用户下线业务
func (u *User) Offline() {
	// 用户下线，将用户从 online map 中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户下线消息
	u.server.BroadCast(u, "下线")
}

// DoMessage 用户发消息
func (u *User) DoMessage(msg string) {
	u.server.BroadCast(u, msg)
}
