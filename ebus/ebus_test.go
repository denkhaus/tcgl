// Tideland Common Go Library - Event Bus - Unit Tests
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus_test

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/applog"
	"github.com/denkhaus/tcgl/asserts"
	"github.com/denkhaus/tcgl/config"
	"github.com/denkhaus/tcgl/ebus"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// TestStartStopSingle
func TestStartStopSingle(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	config := config.New(provider)

	config.Set("backend", "single")

	err := ebus.Init(config)
	assert.Nil(err, "single node backend started")

	agent := ebus.NewTestAgent(1)
	_, err = ebus.Register(agent)
	assert.Nil(err, "agent registered")

	err = ebus.Stop()
	assert.Nil(err, "stopped the bus")

	time.Sleep(100 * time.Millisecond)
	assert.True(agent.Stopped, "agent is stopped")
}

// TestSimpleSingle the single node backend.
func TestSimpleSingle(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	config := config.New(provider)

	config.Set("backend", "single")

	err := ebus.Init(config)
	assert.Nil(err, "single node backend started")
	defer ebus.Stop()

	applog.Debugf("registering agents")
	agent1 := ebus.NewTestAgent(1)
	_, err = ebus.Register(agent1)
	assert.Nil(err, "registered agent 1")
	agent2 := ebus.NewTestAgent(2)
	_, err = ebus.Register(agent2)
	assert.Nil(err, "registered agent 2")
	agent3 := ebus.NewTestAgent(3)
	_, err = ebus.Register(agent3)
	assert.Nil(err, "registered agent 3")
	agent4 := ebus.NewTestAgent(4)
	_, err = ebus.Register(agent4)
	assert.Nil(err, "registered agent 4")

	applog.Debugf("subscribing agents to topics")
	assert.Nil(ebus.Subscribe(agent1, "foo", "bar"), "subscribing agent 1")
	assert.Nil(ebus.Subscribe(agent2, "foo", "baz"), "subscribing agent 2")
	assert.Nil(ebus.Subscribe(agent3, "baz", "yadda", "zong"), "subscribing agent 3")
	assert.Nil(ebus.Subscribe(agent4, "foo", "bar"), "subscribing agent 4")

	applog.Debugf("unsubscribing agents from topics")
	assert.Nil(ebus.Unsubscribe(agent3, "zong"), "usubscribing agent 3")

	applog.Debugf("unsubscribing agents from not subscribed topics")
	assert.Nil(ebus.Unsubscribe(agent3, "argle"), "usubscribing agent 3 from not subscribed topic")

	applog.Debugf("deregistering agents")
	assert.Nil(ebus.Deregister(agent4), "deregistered agent 4")
	time.Sleep(100 * time.Millisecond)
	assert.True(agent4.Stopped, "agent 4 is stopped")

	applog.Debugf("emitting events with subscribers")
	ebus.Emit(ebus.EmptyPayload, "foo")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent1.Counters["foo"], 1, "counter foo for agent 1")
	assert.Equal(agent2.Counters["foo"], 1, "counter foo for agent 2")
	assert.Equal(agent3.Counters["foo"], 0, "counter foo for agent 3")
	assert.Equal(agent4.Counters["foo"], 0, "counter foo for agent 4")

	applog.Debugf("emitting events without subscribers")
	ebus.Emit(ebus.EmptyPayload, "iirks")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(agent1.Counters["iirks"], 0, "counter iirks for agent 1")
	assert.Equal(agent2.Counters["iirks"], 0, "counter iirks for agent 2")
	assert.Equal(agent3.Counters["iirks"], 0, "counter iirks for agent 3")
	assert.Equal(agent4.Counters["iirks"], 0, "counter iirks for agent 4")

	applog.Debugf("pushing events to many subscribers")
	agents := []*ebus.TestAgent{}
	for i := 5; i < 10000; i++ {
		agent := ebus.NewTestAgent(i)
		agents = append(agents, agent)
		ebus.Register(agent)
		ebus.Subscribe(agent, "flirp")
	}
	time.Sleep(100 * time.Millisecond)
	ebus.Emit(ebus.EmptyPayload, "flirp")
	ebus.Emit(ebus.EmptyPayload, "flirp")
	ebus.Emit(ebus.EmptyPayload, "flirp")
	time.Sleep(100 * time.Millisecond)
	for i := 5; i < 10000; i++ {
		agent := agents[i-5]
		assert.Equal(agent.Counters["flirp"], 3, "counter flirp for agent")
	}
}

// TestTicker tests the usage of tickers.
func TestTicker(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	config := config.New(provider)

	config.Set("backend", "single")

	err := ebus.Init(config)
	assert.Nil(err, "single node backend started")

	testAgent := ebus.NewTestAgent(1)
	_, err = ebus.Register(testAgent)
	assert.Nil(err, "test agent registered")
	err = ebus.Subscribe(testAgent, "tick", "tock")

	logAgent := ebus.NewLogAgent("logger")
	_, err = ebus.Register(logAgent)
	assert.Nil(err, "log agent registered")
	err = ebus.Subscribe(logAgent, "tick")

	err = ebus.AddTicker("foo", 100*time.Millisecond, "tick", "tock")
	assert.Nil(err, "ticker foo added")
	err = ebus.AddTicker("bar", 500*time.Millisecond, "tock")
	assert.Nil(err, "ticker bar added")

	time.Sleep(1050 * time.Millisecond)
	assert.Equal(testAgent.Counters["tick"], 10, "got all ticks")
	assert.Equal(testAgent.Counters["tock"], 12, "got all tocks")

	err = ebus.AddTicker("foo", 100*time.Millisecond, "tick", "tock")
	assert.True(ebus.IsDuplicateTickerError(err), "can't add ticker twice")

	err = ebus.RemoveTicker("bar")
	assert.Nil(err, "ticker bar removed")

	err = ebus.RemoveTicker("bar")
	assert.True(ebus.IsTickerNotFoundError(err), "ticker bar is already removed")

	err = ebus.Stop()
	assert.Nil(err, "ebus stopped")

	err = ebus.RemoveTicker("foo")
	assert.True(ebus.IsTickerNotFoundError(err), "ticker foo is removed by ebus stopping")
}

//--------------------
// HELPER
//--------------------

// EOF
