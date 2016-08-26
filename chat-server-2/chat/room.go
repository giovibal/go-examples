package chat

import "net"

type Room struct {
	clients  []*Client
	Joins    chan net.Conn
	incoming chan string
	outgoing chan string
}

func (room *Room) Broadcast(data string) {
	for _, client := range room.clients {
		client.outgoing <- data
	}
}

func (room *Room) Join(connection net.Conn) {
	client := NewClient(connection)
	room.clients = append(room.clients, client)
	go func() {
		for {
			room.incoming <- <-client.incoming
		}
	}()
}

func (room *Room) Listen() {
	go func() {
		for {
			select {
			case data := <-room.incoming:
				room.Broadcast(data)
			case conn := <-room.Joins:
				room.Join(conn)
			}
		}
	}()
}

func NewRoom() *Room {
	room := &Room{
		clients:  make([]*Client, 0),
		Joins:    make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}

	room.Listen()

	return room
}