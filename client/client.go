package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端链接
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       -1,
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

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}

		// 根据不同的模式处理不同的业务
		switch c.flag {
		case 1:
			c.PublicChat()
		case 2:
			c.PrivateChat()
		case 3:
			c.UpdateName()
		}
	}
}

// DealResponse 处理 server 回应的消息，直接显示标准输出
func (c *Client) DealResponse() {
	// 一但 c.conn 有数据，就直接 copy 到 stdout 标准输出上，永久阻塞监听
	_, err := io.Copy(os.Stdout, c.conn)
	if err != nil {
		return
	}
}

// PublicChat 公聊模式
func (c *Client) PublicChat() {
	// 提示用户输入信息
	var chatMsg string

	fmt.Println("> 请输入聊天内容，exit 退出")
	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}

	for chatMsg != "exit" {
		// 发送服务器
		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("> 请输入聊天内容，exit 退出")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}
	}
}

// SelectUsers 查询在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

// PrivateChat 私聊模式
func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUsers()

	fmt.Println("> 请输入聊天对象[用户名]，exit 退出")
	_, err := fmt.Scanln(&remoteName)
	if err != nil {
		return
	}

	for remoteName != "exit" {
		fmt.Println("> 请输入聊天内容，exit 退出")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}

			fmt.Println("> 请输入聊天内容，exit 退出")
			_, err := fmt.Scanln(&chatMsg)
			if err != nil {
				return
			}
		}

		c.SelectUsers()

		fmt.Println("> 请输入聊天对象[用户名]，exit 退出")
		_, err = fmt.Scanln(&remoteName)
		if err != nil {
			return
		}
	}

}

// UpdateName 更新用户名
func (c *Client) UpdateName() bool {
	fmt.Println("> 请输入用户名:")
	_, err := fmt.Scanln(&c.Name)
	if err != nil {
		return false
	}

	sendMsg := "rename|" + c.Name + "\n"
	_, err = c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (c *Client) menu() bool {
	var clientFlag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&clientFlag)
	if err != nil {
		fmt.Println("fmt.Scan err:", err)
	}

	if clientFlag >= 0 && clientFlag <= 3 {
		c.flag = clientFlag
		return true
	}

	fmt.Println("> 请输入合法范围内编号:")
	return false
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

	// 单独开启一个 goroutine 去处理 server 回执消息
	go client.DealResponse()

	fmt.Println("客户端链接成功!")

	// 启动客户端业务
	client.Run()
}
