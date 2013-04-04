// Tideland Common Go Library - Redis - Subscription
//
// Copyright (C) 2009-2013 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// SUBSCRIPTION VALUE
//--------------------

// SubscriptionValue is a result value plus
// channel pattern and channel.
type SubscriptionValue struct {
	Value
	ChannelPattern string
	Channel        string
}

// newSubscriptionValue creates a new subscription value
// based on the raw data.
func newSubscriptionValue(data [][]byte) *SubscriptionValue {
	switch len(data) {
	case 3:
		return &SubscriptionValue{
			Value:          Value(data[2]),
			ChannelPattern: "*",
			Channel:        string(data[1]),
		}
	case 4:
		return &SubscriptionValue{
			Value:          Value(data[3]),
			ChannelPattern: string(data[1]),
			Channel:        string(data[2]),
		}
	}

	return nil
}

//--------------------
// SUBSCRIPTION
//--------------------

// Subscription manages a subscription one or more channels in Redis.
type Subscription struct {
	urp          *unifiedRequestProtocol
	error        error
	channelCount int
	valueChan    chan *SubscriptionValue
}

// newSubscription creates a new subscription.
func newSubscription(urp *unifiedRequestProtocol, channels ...string) *Subscription {
	sub := &Subscription{
		urp:       urp,
		valueChan: make(chan *SubscriptionValue, 10),
	}
	sub.channelCount = sub.urp.subscribe(channels...)
	go sub.backend()
	return sub
}

// Subscribe adds one or more channels to the subscription.
func (s *Subscription) Subscribe(channels ...string) int {
	s.channelCount = s.urp.subscribe(channels...)
	return s.channelCount
}

// Unsubscribe removes one or more channels from the subscription.
func (s *Subscription) Unsubscribe(channels ...string) int {
	s.channelCount = s.urp.unsubscribe(channels...)
	return s.channelCount
}

// ChannelCount returns the number of subscribed channels.
func (s *Subscription) ChannelCount() int {
	return s.channelCount
}

// Values returns a channel emitting the subscription valies.
func (s *Subscription) Values() <-chan *SubscriptionValue {
	return s.valueChan
}

// Stop ends the subscription..
func (s *Subscription) Stop() {
	s.urp.stop()
	close(s.valueChan)
}

// backend is the serving goroutine for the subscription.
func (s *Subscription) backend() {
	for epd := range s.urp.publishedDataChan {
		// Received a published data, republish
		// as subscription value.
		sv := newSubscriptionValue(epd.data)
		// Send the subscription value.
		select {
		case s.valueChan <- sv:
			// OK.
		default:
			// Not sent!
			return
		}
	}
}

// EOF
