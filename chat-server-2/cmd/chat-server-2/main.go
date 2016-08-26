package main

import (
	"net"
	"github.com/giovibal/go-examples/chat-server-2/chat"
)


func main() {
	chatRoom := chat.NewRoom()

	listener, _ := net.Listen("tcp", ":6000")

	for {
		conn, _ := listener.Accept()
		chatRoom.Joins <- conn
	}
}
