// Tideland Common Go Library - Event Bus - Utilities
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"cgl.tideland.biz/applog"
	"cgl.tideland.biz/monitoring"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
)

//--------------------
// ID
//--------------------

// Id creates identifiers used e.g. for topics.
func Id(stem string, parts ...interface{}) string {
	iparts := make([]string, len(parts)+1)
	iparts[0] = stem
	for i, p := range parts {
		switch pv := p.(type) {
		case string:
			iparts[i+1] = pv
		default:
			iparts[i+1] = fmt.Sprintf("%v", pv)
		}
	}
	return strings.Join(iparts, "/")
}

//--------------------
// SIMPLE EVENT
//--------------------

// simpleEvent implements the Event interface.
type simpeEvent struct {
	payload []byte
	topic   string
}

// newSimpleEvent creates a new event instance.
func newSimpleEvent(payload interface{}, topic string) (Event, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(payload)
	if err != nil {
		return nil, err
	}
	payloadBytes := buf.Bytes()
	return &simpeEvent{payloadBytes, topic}, nil
}

// Payload returns the payload of the event.
func (e *simpeEvent) Payload(value interface{}) error {
	buf := bytes.NewBuffer(e.payload)
	dec := gob.NewDecoder(buf)
	return dec.Decode(value)
}

// Topic returns the topic of the event.
func (e *simpeEvent) Topic() string {
	return e.topic
}

//--------------------
// AGENT BOX
//--------------------

type boxMsgKind int

const (
	msgEvent boxMsgKind = iota
	msgSubscribe
	msgUnsubscribe
	msgStop
)

// boxMessage controls an agent.
type boxMessage struct {
	kind  boxMsgKind
	event Event
	topic string
}

// boxEntry is one entry in the linked list of entries
// for messages.
type boxEntry struct {
	message *boxMessage
	next    *boxEntry
}

// box is an inbox for agent control messages.
type box struct {
	cond  *sync.Cond
	first *boxEntry
	last  *boxEntry
}

// newBox creates a new inbox.
func newBox() *box {
	var locker sync.Mutex
	return &box{sync.NewCond(&locker), nil, nil}
}

// push appends a new message to the box.
func (b *box) push(message *boxMessage) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	switch {
	case b.first == nil:
		b.first = &boxEntry{message, nil}
		b.last = nil
	case b.last == nil:
		b.last = &boxEntry{message, nil}
		b.first.next = b.last
	default:
		b.last.next = &boxEntry{message, nil}
		b.last = b.last.next
	}
	b.cond.Signal()
}

// pop retrieves the first message out of the box. If it's 
// empty pop is waiting.
func (b *box) pop() (message *boxMessage) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for {
		if b.first == nil {
			b.cond.Wait()
		} else {
			message = b.first.message
			b.first = b.first.next
			break
		}
	}
	return
}

// len returns the number of messages in the box.
func (b *box) len() int {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	if b.first == nil {
		return 0
	}
	ctr := 1
	current := b.first
	for current.next != nil {
		ctr++
		current = current.next
	}
	return ctr
}

//--------------------
// AGENT RUNNER
//--------------------

// agentRunner manages an agent and lets it process events.
type agentRunner struct {
	agent       Agent
	measuringId string
	inbox       *box
	topics      map[string]bool
}

// newAgentRunner creates a new agent runner
func newAgentRunner(agent Agent) *agentRunner {
	a := &agentRunner{
		agent:       agent,
		measuringId: Id("agent", agent.Id()),
		inbox:       newBox(),
		topics:      make(map[string]bool),
	}
	go a.backend()
	return a
}

// push appends an event for processing.
func (a *agentRunner) push(event Event) {
	message := &boxMessage{msgEvent, event, ""}
	a.inbox.push(message)
}

// subscribe tells the runner to subscribe to a topic.
func (a *agentRunner) subscribe(topic string) {
	message := &boxMessage{msgSubscribe, nil, topic}
	a.inbox.push(message)
}

// unsubscribe tells the runner to unsubscribe from a topic.
func (a *agentRunner) unsubscribe(topic string) {
	message := &boxMessage{msgUnsubscribe, nil, topic}
	a.inbox.push(message)
}

// stop tells the agent to end working.
func (a *agentRunner) stop() {
	message := &boxMessage{msgStop, nil, ""}
	a.inbox.push(message)
}

// backend runs the endless processing loop.
func (a *agentRunner) backend() {
	defer Deregister(a.agent)
	defer a.agent.Stop()
	for {
		message := a.inbox.pop()
		switch message.kind {
		case msgStop:
			return
		case msgSubscribe:
			a.topics[message.topic] = true
		case msgUnsubscribe:
			delete(a.topics, message.topic)
		default:
			if err := a.process(message.event); err != nil {
				applog.Errorf("agent %q is not recoverable after error: %v", a.agent.Id(), err)
				return
			}
		}
	}
}

// process processes one event.
func (a *agentRunner) process(event Event) (err error) {
	// Error recovering.
	defer func() {
		if r := recover(); r != nil {
			applog.Errorf("agent %q has panicked: %v", a.agent.Id(), r)
			err = a.agent.Recover(r, event)
		}
	}()
	// Handle the event inside a measuring.
	measuring := monitoring.BeginMeasuring(a.measuringId)
	defer measuring.EndMeasuring()
	if err = a.agent.Process(event); err != nil {
		applog.Errorf("agent %q has failed: %v", a.agent.Id(), err)
		return a.agent.Recover(err, event)
	}
	return nil
}

//--------------------
// NODE ROUTER
//--------------------

type opRegister struct {
	agent    Agent
	response chan *response
}

type opDeregister struct {
	agent    Agent
	response chan *response
}

type opLookup struct {
	id       string
	response chan *response
}

type opSubscribe struct {
	agent    Agent
	topic    string
	response chan *response
}

type opUnsubscribe struct {
	agent    Agent
	topic    string
	response chan *response
}

type opPush struct {
	event    Event
	response chan *response
}

type opStop struct{}

type response struct {
	agent Agent
	err   error
}

// nodeRouter manages registrations and subsciptions per node.
type nodeRouter struct {
	registry      map[string]*agentRunner
	topic2Runners map[string]map[string]*agentRunner
	ops           chan interface{}
}

// newNodeRouter create a new node router.
func newNodeRouter() *nodeRouter {
	n := &nodeRouter{
		registry:      make(map[string]*agentRunner),
		topic2Runners: make(map[string]map[string]*agentRunner),
		ops:           make(chan interface{}),
	}
	go n.backend()
	return n
}

// register registers an agent at the router.
func (n *nodeRouter) register(agent Agent) error {
	op := &opRegister{agent, make(chan *response)}
	n.ops <- op
	response := <-op.response
	return response.err
}

// deregister unsubscribes an agent from all topics and removes 
// it from the router.
func (n *nodeRouter) deregister(agent Agent) error {
	op := &opDeregister{agent, make(chan *response)}
	n.ops <- op
	response := <-op.response
	return response.err
}

// lookup retrieves a registered agent by id.
func (n *nodeRouter) lookup(id string) (Agent, error) {
	op := &opLookup{id, make(chan *response)}
	n.ops <- op
	response := <-op.response
	if response.err != nil {
		return nil, response.err
	}
	return response.agent, nil
}

// subscribe subscribes the agent to the topic.
func (n *nodeRouter) subscribe(agent Agent, topic string) error {
	op := &opSubscribe{agent, topic, make(chan *response)}
	n.ops <- op
	response := <-op.response
	return response.err
}

// unsubscribe removes the subscription of the agent from the topic.
func (n *nodeRouter) unsubscribe(agent Agent, topic string) error {
	op := &opUnsubscribe{agent, topic, make(chan *response)}
	n.ops <- op
	response := <-op.response
	return response.err
}

// push pushes an event to the router so that will be delivered
// to all subscribers.
func (n *nodeRouter) push(event Event) error {
	op := &opPush{event, make(chan *response)}
	n.ops <- op
	response := <-op.response
	return response.err
}

// stop tells the router to stop working.
func (n *nodeRouter) stop() {
	n.ops <- &opStop{}
}

// backend runs the endless processing loop.
func (n *nodeRouter) backend() {
	defer n.stopAgents()
	for next := range n.ops {
		switch op := next.(type) {
		case *opRegister:
			id := op.agent.Id()
			if n.registry[id] != nil {
				op.response <- &response{nil, &DuplicateAgentIdError{id}}
				continue
			}
			// Regiser new agent runner.
			n.registry[id] = newAgentRunner(op.agent)
			op.response <- &response{}
		case *opDeregister:
			id := op.agent.Id()
			runner := n.registry[id]
			if runner == nil {
				op.response <- &response{nil, &AgentNotRegisteredError{id}}
				continue
			}
			// Deregister and unsubscribe agent runner.
			delete(n.registry, id)
			runner.stop()
			for topic := range runner.topics {
				delete(n.topic2Runners[topic], id)
			}
			op.response <- &response{}
		case *opLookup:
			id := op.id
			runner := n.registry[id]
			if runner == nil {
				op.response <- &response{nil, &AgentNotRegisteredError{id}}
				continue
			}
			op.response <- &response{runner.agent, nil}
		case *opSubscribe:
			id := op.agent.Id()
			runner := n.registry[id]
			if runner == nil {
				op.response <- &response{nil, &AgentNotRegisteredError{id}}
				continue
			}
			// Subscribe agent runner.
			runner.subscribe(op.topic)
			if n.topic2Runners[op.topic] == nil {
				n.topic2Runners[op.topic] = make(map[string]*agentRunner)
			}
			n.topic2Runners[op.topic][id] = runner
			op.response <- &response{}
		case *opUnsubscribe:
			id := op.agent.Id()
			runner := n.registry[id]
			if runner == nil {
				op.response <- &response{nil, &AgentNotRegisteredError{id}}
				continue
			}
			// Unsubscribe agent runner.
			runner.unsubscribe(op.topic)
			if n.topic2Runners[op.topic] != nil {
				delete(n.topic2Runners[op.topic], id)
			}
			if len(n.topic2Runners[op.topic]) == 0 {
				delete(n.topic2Runners, op.topic)
			}
			op.response <- &response{}
		case *opPush:
			runners := n.topic2Runners[op.event.Topic()]
			if runners == nil {
				op.response <- &response{nil, &NoSubscriberError{op.event.Topic()}}
				continue
			}
			for _, runner := range runners {
				runner.push(op.event)
			}
			op.response <- &response{}
		case *opStop:
			return
		}
	}
}

// stopAgents stops the remaining agent runner when the
// router stops.
func (n *nodeRouter) stopAgents() {
	for _, runner := range n.registry {
		runner.stop()
	}
}

// EOF
