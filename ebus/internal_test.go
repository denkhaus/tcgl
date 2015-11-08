// Tideland Common Go Library - Event Bus - Unit Tests
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/applog"
	"github.com/denkhaus/tcgl/asserts"
	"github.com/denkhaus/tcgl/config"
	"fmt"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// TestBox tests the box usage.
func TestEventBox(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	inbox := newBox()

	assert.Equal(inbox.len(), 0, "box is empty")

	inbox.push(EventMessage(EmptyPayload, "Event", 1))
	inbox.push(EventMessage(EmptyPayload, "Event", 2))
	inbox.push(EventMessage(EmptyPayload, "Event", 3))

	assert.Equal(inbox.len(), 3, "box has right length")
	assert.Equal(inbox.pop().event.Topic(), Id("Event", 1), "first event")
	assert.Equal(inbox.pop().event.Topic(), Id("Event", 2), "second event")
	assert.Equal(inbox.pop().event.Topic(), Id("Event", 3), "third event")

	go func() {
		assert.Equal(inbox.pop().event.Topic(), Id("Event", 4), "fourth event")

		inbox.push(EventMessage(EmptyPayload, "Event", 5))
	}()

	inbox.push(EventMessage(EmptyPayload, "Event", 4))
	time.Sleep(100 * time.Millisecond)
	assert.Equal(inbox.pop().event.Topic(), Id("Event", 5), "fifth event")
}

// TestAgentRunner tests the runtime for an agent.
func TestAgentRunner(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	InitSingle()
	defer Stop()

	agent := NewTestAgent(1)
	runner := newAgentRunner(agent)

	// Standard processings.
	for i := 0; i < 10; i++ {
		runnerPush(assert, runner, EmptyPayload, "one")
		if i%2 == 0 {
			runnerPush(assert, runner, EmptyPayload, "two")
		}
		if i%3 == 0 {
			runnerPush(assert, runner, EmptyPayload, "three")
		}
	}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent.Counters["one"], 10, "counter one")
	assert.Equal(agent.Counters["two"], 5, "counter two")
	assert.Equal(agent.Counters["three"], 4, "counter three")

	// Error handling.
	runnerPush(assert, runner, EmptyPayload, "error")
	runnerPush(assert, runner, EmptyPayload, "error")
	runnerPush(assert, runner, EmptyPayload, "panic")

	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent.Recoverings["error"], 2, "error recovering")
	assert.Equal(agent.Recoverings["panic"], 1, "panic recovering")

	runnerPush(assert, runner, EmptyPayload, "hard-panic")
	time.Sleep(100 * time.Millisecond)
	runner.stop()

	assert.ErrorMatch(agent.Err(), "hard panic is too hard for me", "hard panic")
}

// TestNodeRouter tests the event router for one node.
func TestNodeRouter(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	assert.Nil(InitSingle(), "init the single backend")
	defer Stop()

	agent1 := NewTestAgent(1)
	agent2 := NewTestAgent(2)
	agent3 := NewTestAgent(3)
	agent4 := NewTestAgent(4)
	router := newNodeRouter()
	defer router.stop()

	applog.Debugf("registering agents")
	assert.Nil(router.register(agent1), "registered agent 1")
	assert.Nil(router.register(agent2), "registered agent 2")
	assert.Nil(router.register(agent3), "registered agent 3")
	assert.Nil(router.register(agent4), "registered agent 4")

	applog.Debugf("lookup agents")
	la1, err := router.lookup(agent1.Id())
	assert.Nil(err, "lookup for agent 1")
	assert.Equal(la1, agent1, "lookup for agent 1")

	applog.Debugf("subscribing agents to topics")
	assert.Nil(router.subscribe(agent1, "foo"), "subscribing agent 1")
	assert.Nil(router.subscribe(agent1, "bar"), "subscribing agent 1")
	assert.Nil(router.subscribe(agent2, "foo"), "subscribing agent 2")
	assert.Nil(router.subscribe(agent2, "baz"), "subscribing agent 2")
	assert.Nil(router.subscribe(agent3, "baz"), "subscribing agent 3")
	assert.Nil(router.subscribe(agent3, "yadda"), "subscribing agent 3")
	assert.Nil(router.subscribe(agent3, "zong"), "subscribing agent 3")
	assert.Nil(router.subscribe(agent4, "foo"), "subscribing agent 4")
	assert.Nil(router.subscribe(agent4, "bar"), "subscribing agent 4")

	applog.Debugf("unsubscribing agents from topics")
	assert.Nil(router.unsubscribe(agent3, "zong"), "usubscribing agent 3")

	applog.Debugf("unsubscribing agents from not subscribed topics")
	assert.Nil(router.unsubscribe(agent3, "argle"), "usubscribing agent 3 from not subscribed topic")

	applog.Debugf("deregistering agents")
	assert.Nil(router.deregister(agent4), "deregistered agent 4")
	time.Sleep(100 * time.Millisecond)
	assert.True(agent4.Stopped, "agent 4 is stopped")

	applog.Debugf("pushing events with subscribers")
	routerPush(assert, router, EmptyPayload, "one")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent1.Counters["foo"], 1, "counter foo for agent 1")
	assert.Equal(agent2.Counters["foo"], 1, "counter foo for agent 2")
	assert.Equal(agent3.Counters["foo"], 0, "counter foo for agent 3")
	assert.Equal(agent4.Counters["foo"], 0, "counter foo for agent 4")

	applog.Debugf("pushing events without subscribers")
	routerPush(assert, router, EmptyPayload, "iirks")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent1.Counters["iirks"], 0, "counter iirks for agent 1")
	assert.Equal(agent2.Counters["iirks"], 0, "counter iirks for agent 2")
	assert.Equal(agent3.Counters["iirks"], 0, "counter iirks for agent 3")
	assert.Equal(agent4.Counters["iirks"], 0, "counter iirks for agent 4")

	applog.Debugf("pushing events to many subscribers")
	agents := []*TestAgent{}
	for i := 5; i < 10000; i++ {
		agent := NewTestAgent(i)
		agents = append(agents, agent)
		router.register(agent)
		router.subscribe(agent, "flirp")
	}
	time.Sleep(100 * time.Millisecond)
	routerPush(assert, router, EmptyPayload, "flirp")
	routerPush(assert, router, EmptyPayload, "flirp")
	routerPush(assert, router, EmptyPayload, "flirp")
	time.Sleep(100 * time.Millisecond)
	for i := 5; i < 10000; i++ {
		agent := agents[i-5]
		assert.Equal(agent.Counters["flirp"], 3, "counter flirp for agent")
	}
}

//--------------------
// HELPERS
//--------------------

func runnerPush(a *asserts.Asserts, r *agentRunner, p interface{}, topic string) {
	event, err := newSimpleEvent(p, topic)
	a.Nil(err, "no error in new event")
	r.push(event)
}

func routerPush(a *asserts.Asserts, r *nodeRouter, p interface{}, topic string) {
	event, err := newSimpleEvent(p, topic)
	a.Nil(err, "no error in new event")
	err = r.push(event)
	a.Nil(err, "no error during push")
}

var EmptyPayload = struct {
	A int
	B string
}{
	A: 4711,
	B: "foobar",
}

// EventMessage create a box message with an event.
func EventMessage(payload interface{}, stem string, parts ...interface{}) *boxMessage {
	event, _ := newSimpleEvent(payload, Id(stem, parts...))
	return &boxMessage{msgEvent, event, ""}
}

func InitSingle() error {
	provider := config.NewMapConfigurationProvider()
	config := config.New(provider)

	config.Set("backend", "single")

	return Init(config)
}

// testAgent is used to test the agent runner.
type TestAgent struct {
	id          string
	Counters    map[string]int
	Recoverings map[string]int
	Stopped     bool
	err         error
}

// NewTestAgent creates a new test agent.
func NewTestAgent(no int) *TestAgent {
	return &TestAgent{
		id:          Id("TestAgent", no),
		Counters:    make(map[string]int),
		Recoverings: make(map[string]int),
	}
}

// Id returns the unique identifier of the agent.
func (t *TestAgent) Id() string {
	return t.id
}

// Process processes an event.
func (t *TestAgent) Process(event Event) error {
	switch event.Topic() {
	case "reset":
		t.Counters = make(map[string]int)
	case "error":
		return fmt.Errorf("ouch, an error")
	case "panic":
		panic("ouch, a panic")
	case "hard-panic":
		panic("ouch, a hard panic")
	default:
		t.Counters[event.Topic()]++
	}
	return nil
}

// Recover from an error during the processing of an event.
func (t *TestAgent) Recover(r interface{}, event Event) error {
	t.Recoverings[event.Topic()]++
	if event.Topic() == "hard-panic" {
		t.err = fmt.Errorf("hard panic is too hard for me")
		return t.err
	}
	return nil
}

// Stop tells the agent to cleanup.
func (t *TestAgent) Stop() {
	t.Stopped = true
}

// Err returns the error the agent possibly stopped with.
func (t *TestAgent) Err() error {
	return t.err
}

// EOF
