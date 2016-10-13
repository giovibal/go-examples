package routing

import (
	"github.com/giovibal/go-examples/mqtt-server-2/client"
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"net"
)

type Router struct {
	clients map[net.Conn]*client.Client

	publishchan chan *packets.PublishPacket
	addchan     chan *client.Client
	rmchan      chan *client.Client
}

func New() *Router {
	return &Router{
		clients: make(map[net.Conn]*client.Client),
		publishchan: make(chan *packets.PublishPacket, 8),
		addchan: make(chan *client.Client, 4),
		rmchan: make(chan *client.Client, 4),
	}
}

func (router *Router) addClient(client *client.Client) {
	router.clients[client.Conn] = client
}
func (router *Router) removeClient(client *client.Client) {
	delete(router.clients, client.Conn)
}

func (router *Router) Subscribe(c *client.Client)  {
	router.addchan <- c
}
func (router *Router) Unsubscribe(c *client.Client)  {
	router.rmchan <- c
}
func (router *Router) Publish(pubMsg *packets.PublishPacket) {
	router.publishchan <- pubMsg
}


//func (router *Router) RouteMessages(publishchan <-chan *packets.PublishPacket, addchan <-chan *client.Client, rmchan <-chan *client.Client) {
func (router *Router) RouteMessages() {
	clients := router.clients
	for {
		select {
		case msg := <-router.publishchan:
			// example of publish to all clients...
			for _, c := range clients {
				// check subscription match
				publishingTopic := msg.TopicName
				ok, _ := c.IsSubscribed(publishingTopic)
				if ok {
					// TODO: look at mqtt specs for qos handling...
					//msg.Qos = subscriptionQos
					c.WritePublishMessage(msg)
				}
			}

		case c := <-router.addchan:
			router.addClient(c)
		case c := <-router.rmchan:
			router.removeClient(c)
		}
	}
}

