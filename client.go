package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var (
	serverIp   string
	serverPort int
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	signal     int
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		signal:     999,
	}

	//链接服务端
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	return client
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var signal int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&signal)

	if signal >= 0 && signal <= 3 {
		client.signal = signal
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<")
		return false
	}

}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit推出。")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit推出。")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) Run() {
	for client.signal != 0 {
		for client.menu() != true {

		}

		//根据不同模式处理业务
		switch client.signal {
		case 1:
			//fmt.Println("公聊模式选择...")
			client.PublicChat()
			break
		case 2:
			//fmt.Println("私聊模式选择...")
			client.PrivateChat()
			break
		case 3:
			//fmt.Println("更改用户名选择...")
			client.UpdateName()
			break

		}
	}
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"

	_, err := client.conn.Write([]byte(sendMsg))

	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}

}

//私聊模式
func (client *Client) PrivateChat() {
	var (
		remoteName string
		chatMsg    string
	)

	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象[用户名],exit推出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入聊天内容，exit推出。")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入聊天内容，exit推出。")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象[用户名],exit推出：")
		fmt.Scanln(&remoteName)
	}
}

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认8888）")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println(">>>>>> 连接服务器失败...")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>>>连接服务器成功...")

	client.Run()
}
