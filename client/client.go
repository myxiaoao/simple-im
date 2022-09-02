package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端链接
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	// 链接 server
	dial, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = dial

	// 返回对象
	return client
}

var serverIp string
var serverPort int

func init() {
	// ./client -ip 127.0.0.1 -port 8888
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器 IP 地址，默认 127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口，默认 8888")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("客户端链接失败!")
		return
	}

	fmt.Println("客户端链接成功!")

	select {}
}
