package mqtt

import (
	"regexp"
	"log"
	"strings"
)

type Subscription struct {
	TopicFilter string
	Qos         byte
	Regexp      *regexp.Regexp
}



func NewSubscription(topic string, qos byte) *Subscription {
	re, err := toRegexPattern(topic)
	if err != nil {
		log.Fatal(err)
	}
	s := &Subscription{
		TopicFilter: topic,
		Qos: qos,
		Regexp: re,
	}
	return s
}
func toRegexPattern(subscribedTopic string) (*regexp.Regexp, error) {
	var regexPattern string
	regexPattern = subscribedTopic
	regexPattern = strings.Replace(regexPattern, "#", ".*", -1)
	regexPattern = strings.Replace(regexPattern, "+", "[^/]*", -1)
	pattern, err := regexp.Compile("^"+regexPattern+"$")
	return pattern, err
}

func (s *Subscription) IsSubscribed(publishingTopic string) (bool, byte) {
	if strings.Compare(s.TopicFilter, publishingTopic) == 0 {
		return true, s.Qos
	} else {
		topicMatches := s.Regexp.MatchString(publishingTopic)
		calculatedQos := s.Qos
		return topicMatches, calculatedQos
	}
}
