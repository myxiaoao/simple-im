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
}

// NewUser 创建一个用户的 API
func NewUser(conn net.Conn) *User {
	s := conn.RemoteAddr().String()

	user := &User{
		Name: s,
		Addr: s,
		C:    make(chan string),
		conn: conn,
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
