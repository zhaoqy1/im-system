package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	//启动监听当前user channel消息的goroutine

	go user.ListenMessage()

	return user
}

// Online 用户上线功能
func (this *User) Online() {
	//用户上线，将用户加入到onlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// Offline 用户上线功能
func (this *User) Offline() {
	//用户上线，将用户加入到onlineMap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "下线")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 处理用户消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些

		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"

			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]

		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
			return
		}

		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发\n")
			return
		}

		remoteUser.SendMsg(this.Name + "对您说：" + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
