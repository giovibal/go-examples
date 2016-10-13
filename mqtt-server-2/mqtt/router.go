package mqtt

import (
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
)

type Router struct {
	clients map[string]*Client

	publishchan chan *packets.PublishPacket
	addchan     chan *Client
	rmchan      chan *Client
}

func NewRouter() *Router {
	return &Router{
		clients: make(map[string]*Client),
		publishchan: make(chan *packets.PublishPacket, 8),
		addchan: make(chan *Client, 4),
		rmchan: make(chan *Client, 4),
	}
}

func (router *Router) addClient(client *Client) {
	router.clients[client.ID] = client
}
func (router *Router) removeClient(client *Client) {
	delete(router.clients, client.ID)
}

func (router *Router) Subscribe(c *Client)  {
	router.addchan <- c
}
func (router *Router) Unsubscribe(c *Client)  {
	router.rmchan <- c
}
func (router *Router) Publish(pubMsg *packets.PublishPacket) {
	router.publishchan <- pubMsg
}

func (router *Router) Start() {
	go router.routeMessages()
}
func (router *Router) routeMessages() {
	clients := router.clients
	for {
		select {
		case msg := <-router.publishchan:
			// example of publish to all clients...
			for _, c := range clients {
				// check subscription match
				publishingTopic := msg.TopicName
				ok := c.IsSubscribed(publishingTopic)
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
