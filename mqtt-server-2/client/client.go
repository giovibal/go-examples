package client

import (
	"bufio"
	"github.com/giovibal/go-examples/mqtt-server-2/packets"
	"log"
	"net"
	"regexp"
	"strings"
)

type Client struct {
	Conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	incoming      chan packets.ControlPacket
	outgoing      chan packets.ControlPacket
	Subscriptions map[string]Subscription
}
type Subscription struct {
	TopicFilter string
	Qos         byte
	Regexp      *regexp.Regexp
}
//
//func New(conn net.Conn) *Client {
//	return &Client{
//		Conn:          conn,
//		Subscriptions: make(map[string]Subscription),
//		reader:    bufio.NewReader(conn),
//		writer:    bufio.NewWriter(conn),
//	}
//}
func New(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		Conn:     connection,
		incoming: make(chan packets.ControlPacket),
		outgoing: make(chan packets.ControlPacket),
		reader:   reader,
		writer:   writer,
	}

	client.Listen()

	return client
}

// tcp conn >>> channel
func (c *Client) HandleMqttProtocol(publishchan chan<- *packets.PublishPacket, addchan chan<- *Client, rmchan chan<- *Client) {
	//var err error
	//var cp packets.ControlPacket
	for {
		select {
		case cp := <- c.incoming:
			log.Printf(">> %v \n", cp.String())

			msgType := cp.GetMessageType()
			switch msgType {

			case packets.Connect:
				connackMsg := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
				c.outgoing <- connackMsg
				break
			case packets.Subscribe:
				addchan <- c
				subackMsg := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
				subackMsg.MessageID = cp.Details().MessageID

				subMsg := cp.(*packets.SubscribePacket)
				topics := subMsg.Topics
				qoss := subMsg.Qoss

				for i := 0; i < len(topics); i++ {
					topic := topics[i]
					var qos byte
					if i < len(qoss) {
						qos = qoss[i]
					} else {
						qos = 0x0
					}
					re, err := toRegexPattern(topic)
					if err != nil {
						log.Fatal(err)
					}
					s := Subscription{TopicFilter: topic, Qos: qos, Regexp: re}
					c.Subscriptions[topic] = s
				}

				c.outgoing <- subackMsg
				break
			case packets.Unsubscribe:
				rmchan <- c
				c.Subscriptions = make(map[string]Subscription)
				unsubackMsg := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
				unsubackMsg.MessageID = cp.Details().MessageID
				c.outgoing <- unsubackMsg
				break
			case packets.Disconnect:
				break
			case packets.Publish:
				qos := cp.Details().Qos
				pubMsg := cp.(*packets.PublishPacket)
				switch qos {
				case 0:
					publishchan <- pubMsg
					break
				case 1:
					publishchan <- pubMsg
					pubAck := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
					pubAck.MessageID = cp.Details().MessageID
					pubAck.Write(c.writer)
					break
				case 2:
					publishchan <- pubMsg
					pubRec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
					pubRec.MessageID = cp.Details().MessageID
					pubRec.Write(c.writer)
					break
				}
			case packets.Pubrec:
				pubRel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
				pubRel.MessageID = cp.Details().MessageID
				c.outgoing <- pubRel
				break

			case packets.Pubrel:
				pubComp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
				pubComp.MessageID = cp.Details().MessageID
				c.outgoing <- pubComp
				break

			case packets.Puback:
				break

			case packets.Pingreq:
				pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
				c.outgoing <- pingresp
				break

			default:
			}

		}
	}
}


// tcp conn <<< channel
func (c *Client) WriteMessagesFrom(ch <-chan *packets.PublishPacket) {
	for msg := range ch {
		err := msg.Write(c.writer)
		if err != nil {
			return
		}
		c.writer.Flush()
	}
}

func (client *Client) Read() {
	for {
		data, err := packets.ReadPacket(client.reader)
		if err != nil {
			break
		}
		client.incoming <- data
	}
}
func (client *Client) Write() {
	for data := range client.outgoing {
		data.Write(client.writer)
		client.writer.Flush()
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}




func (c *Client) IsSubscribed(publishingTopic string) (bool, byte) {
	var topicMatches bool = false
	var calculatedQos byte = 0x0
	for _, subscription := range c.Subscriptions {
		//regexp, err := toRegexPattern(subscriptionTopicFilter)
		//if err != nil {
		//	log.Fatal(err)
		//	topicMatches = false
		//	calculatedQos = 0x0
		//}
		regexp := subscription.Regexp
		topicMatches := regexp.MatchString(publishingTopic)
		calculatedQos = subscription.Qos
		return topicMatches, calculatedQos
	}
	return topicMatches, calculatedQos
}
func (c *Client) WritePublishMessage(msg *packets.PublishPacket) {
	go func(ch chan packets.ControlPacket) {
		ch <- msg
	}(c.incoming)
}

func toRegexPattern(subscribedTopic string) (*regexp.Regexp, error) {
	var regexPattern string
	regexPattern = subscribedTopic
	regexPattern = strings.Replace(regexPattern, "#", ".*", -1)
	regexPattern = strings.Replace(regexPattern, "+", "[^/]*", -1)
	pattern, err := regexp.Compile(regexPattern)
	return pattern, err
}
