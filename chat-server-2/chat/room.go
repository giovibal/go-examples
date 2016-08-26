package chat

import "net"

type Room struct {
	clients  []*Client
	Joins    chan net.Conn
	incoming chan string
	outgoing chan string
}

func (chatRoom *Room) Broadcast(data string) {
	for _, client := range chatRoom.clients {
		client.outgoing <- data
	}
}

func (chatRoom *Room) Join(connection net.Conn) {
	client := NewClient(connection)
	chatRoom.clients = append(chatRoom.clients, client)
	go func() {
		for {
			chatRoom.incoming <- <-client.incoming
		}
	}()
}

func (chatRoom *Room) Listen() {
	go func() {
		for {
			select {
			case data := <-chatRoom.incoming:
				chatRoom.Broadcast(data)
			case conn := <-chatRoom.Joins:
				chatRoom.Join(conn)
			}
		}
	}()
}

func NewRoom() *Room {
	chatRoom := &Room{
		clients:  make([]*Client, 0),
		Joins:    make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}

	chatRoom.Listen()

	return chatRoom
}