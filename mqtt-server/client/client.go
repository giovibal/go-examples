package client

import (
	"bufio"
	"github.com/giovibal/go-examples/mqtt-server/packets"
	"log"
	"net"
	"regexp"
	"strings"
	"io"
)

type Client struct {
	Conn          net.Conn
	DataReader    *bufio.Reader
	DataWriter    io.Writer
	Ch            chan *packets.PublishPacket
	Subscriptions map[string]Subscription
}
type Subscription struct {
	TopicFilter string
	Qos         byte
	Regexp      *regexp.Regexp
}

func New(conn net.Conn) *Client {
	return &Client{
		Conn:          conn,
		Ch:            make(chan *packets.PublishPacket),
		Subscriptions: make(map[string]Subscription),
		DataReader:    bufio.NewReader(conn),
		DataWriter:    conn,
	}
}

// tcp conn >>> channel
func (c *Client) ReadMessagesInto(publishchan chan<- *packets.PublishPacket, addchan chan<- *Client, rmchan chan<- *Client) {
	var err error
	var cp packets.ControlPacket
	for {
		if cp, err = packets.ReadPacket(c.DataReader); err != nil {
			break
		}
		//log.Printf(">> %v \n", cp.String())
		// handle mqtt

		msgType := cp.GetMessageType()
		switch msgType {

		case packets.Connect:
			connackMsg := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
			connackMsg.Write(c.DataWriter)
			break
		case packets.Subscribe:
			addchan <- c
			subackMsg := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
			subackMsg.MessageID = cp.Details().MessageID
			subackMsg.Write(c.DataWriter)

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
			break
		case packets.Unsubscribe:
			rmchan <- c
			c.Subscriptions = make(map[string]Subscription)
			unsubackMsg := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
			unsubackMsg.MessageID = cp.Details().MessageID
			unsubackMsg.Write(c.DataWriter)
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
				pubAck.Write(c.DataWriter)
				break
			case 2:
				publishchan <- pubMsg
				pubRec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
				pubRec.MessageID = cp.Details().MessageID
				pubRec.Write(c.DataWriter)
				break
			}
		case packets.Pubrec:
			pubRel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
			pubRel.MessageID = cp.Details().MessageID
			pubRel.Write(c.DataWriter)
			break

		case packets.Pubrel:
			pubComp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
			pubComp.MessageID = cp.Details().MessageID
			pubComp.Write(c.DataWriter)
			break

		case packets.Puback:
			break

		case packets.Pingreq:
			pingresp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
			pingresp.Write(c.DataWriter)
			break

		default:
		}

		//c.DataWriter.Flush()
		//ch <- cp
	}
}

// tcp conn <<< channel
func (c *Client) WriteMessagesFrom(ch <-chan *packets.PublishPacket) {
	for msg := range ch {
		err := msg.Write(c.DataWriter)
		if err != nil {
			return
		}
	}
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
	go func(mch chan *packets.PublishPacket) {
		mch <- msg
	}(c.Ch)
}

func toRegexPattern(subscribedTopic string) (*regexp.Regexp, error) {
	var regexPattern string
	regexPattern = subscribedTopic
	regexPattern = strings.Replace(regexPattern, "#", ".*", -1)
	regexPattern = strings.Replace(regexPattern, "+", "[^/]*", -1)
	pattern, err := regexp.Compile(regexPattern)
	return pattern, err
}
