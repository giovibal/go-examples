package mqtt

import (
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"log"
)

type Router struct {
	clients          map[string]*Client
	connectedClients map[string]*Client

	publishchan      chan *packets.PublishPacket
	addchan          chan *Client
	rmchan           chan *Client

	retainStore map[string]*packets.PublishPacket
}

func NewRouter() *Router {
	return &Router{
		clients:     make(map[string]*Client),
		connectedClients:     make(map[string]*Client),
		publishchan: make(chan *packets.PublishPacket, 8),
		addchan:     make(chan *Client, 2),
		rmchan:      make(chan *Client, 2),
		retainStore: make(map[string]*packets.PublishPacket),
	}
}

func (router *Router) addClient(c *Client) {
	router.clients[c.ID] = c
}
func (router *Router) removeClient(c *Client) {
	delete(router.clients, c.ID)
}

func (router *Router) Subscribe(c *Client) {
	router.addchan <- c
}
func (router *Router) Unsubscribe(c *Client) {
	router.rmchan <- c
}
func (router *Router) Publish(pubMsg *packets.PublishPacket) {
	if pubMsg.Retain {
		log.Printf("payload retained: %s\n", len(pubMsg.Payload))
		if len(pubMsg.Payload) == 0 {
			log.Printf("payload retained empty: %s, delete retained message\n", len(pubMsg.Payload))
			delete(router.retainStore, pubMsg.TopicName)
		} else {
			router.retainStore[pubMsg.TopicName] = pubMsg
		}
	}
	router.publishchan <- pubMsg
}
func (router *Router) RepublishRetainedMessages(c *Client, subMsg *packets.SubscribePacket) {
	msgid := subMsg.MessageID
	for _, pubMsg := range router.retainStore {
		msgid++
		msg := pubMsg.Copy()
		msg.FixedHeader = pubMsg.FixedHeader
		msg.MessageID = msgid

		// check subscription match
		publishingTopic := msg.TopicName
		subscribed := c.IsSubscribed(publishingTopic)
		if subscribed {
			c.WritePublishMessage(msg)
		}
	}
}

func (router *Router) Start() {
	go router.routeMessages()
}

func (router *Router) routeMessages() {
	clients := router.clients
	for {
		select {
		case msg := <-router.publishchan:
			for _, c := range clients {
				// check subscription match
				publishingTopic := msg.TopicName
				subscribed := c.IsSubscribed(publishingTopic)
				if subscribed {
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

func (r *Router) Connect(c *Client) {
	r.connectedClients[c.ID] = c
}
func (r *Router) Disconnect(c *Client) {
	delete(r.connectedClients, c.ID)
}
func (r *Router) Connected(c *Client) bool {
	_, present := r.connectedClients[c.ID]
	return present
}
