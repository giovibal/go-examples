package main

import (
	"github.com/giovibal/go-examples/chat-server-2/chat"
	"net"
)

func main() {
	chatRoom := chat.NewRoom()

	listener, _ := net.Listen("tcp", ":6000")

	for {
		conn, _ := listener.Accept()
		chatRoom.HandleNewConnection(conn)
	}
}
