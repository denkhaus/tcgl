// Tideland Common Go Library - Cells - Unit Tests
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/asserts"
	"testing"
	"strings"
	"time"
)

//--------------------
// HELPER FUNCTIONS
//--------------------

// Counter is a function for the counter behavior. Here
// it just dispatches the topic to the counter variable.
func Counter(e Event) []string {
	return []string{e.Topic()}
}

//--------------------
// TESTS
//--------------------

// TestContext tests the event context.
func TestContext(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Check setting and getting.
	c := newContext()
	c.Set("foo", 4711)
	c.Set("bar", "BAR")
	i, err := c.Get("foo")
	assert.Nil(err, "No error getting 'foo'.")
	assert.Equal(i, 4711, "Right value getting 'foo'.")
	s, err := c.Get("bar")
	assert.Nil(err, "No error getting 'bar'.")
	assert.Equal(s, "BAR", "Right value getting 'bar'.")
	x, err := c.Get("baz")
	assert.ErrorMatch(err, `no context value for "baz"`, "Right error for illegal key.")
	assert.Nil(x, "No value found for 'baz'.")

	// Check activity and wait.
	step := 0
	c.incrActivity()
	err = c.Wait(50 * time.Millisecond)
	assert.ErrorMatch(err, "timeout during context wait", "Wait doesn't finish, so timeout.")
	go func() {
		step = 1
		c.incrActivity()
		step = 2
		c.incrActivity()
		step = 3
		c.decrActivity()
		step = 4
		c.incrActivity()
		step = 5
		c.decrActivity()
		step = 6
		c.decrActivity()
		step = 7
		c.decrActivity()
		step = 8
		c.decrActivity()
		step = 9
	}()
	err = c.Wait(50 * time.Millisecond)
	assert.Nil(err, "No timeout during wait.")
	assert.Equal(step, 9, "Right increments and decrements before end of waiting.")

	// Check iteration and value count.
	valueCount := 0
	c.Do(func(key Id, value interface{}) {
		valueCount++
		switch key {
		case "foo":
			assert.Equal(value, 4711, "Right value getting 'foo'.")
		case "bar":
			assert.Equal(value, "BAR", "Right value getting 'bar'.")
		}
	})
	assert.Equal(valueCount, 2, "Right number of values.")
}

// TestEnvironment tests the environment creation and shutdown.
func TestEnvironment(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("environment")

	assert.NotNil(env, "Environment is created.")
	assert.Equal(env.id, Id("environment"), "Environment id is 'environment'.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestAddCell tests the adding of cells.
func TestAddCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("add-cell")

	counter, err := env.AddCell("counter", NewCounterBehaviorFactory(Counter))
	assert.Nil(err, "Added counter.")
	assert.NotNil(counter, "Added counter.")
	assert.True(env.HasCell("counter"), "Added counter.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestDoubleAddCell tests the adding of cells with
// the same id.
func TestDoubleAddCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("double-add-cell")

	counter, err := env.AddCell("counter", NewCounterBehaviorFactory(Counter))
	assert.Nil(err, "Added counter the first time.")
	assert.NotNil(counter, "Added counter the first time.")
	counter, err = env.AddCell("counter", NewCounterBehaviorFactory(Counter))
	assert.Nil(counter, "Counter can't be added twice.")
	assert.ErrorMatch(err, `cell "counter" already exists`, "Counter can't added twice.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestAddCells tests the adding of multiple cells.
func TestAddCells(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("add-cells")
	bfm := BehaviorFactoryMap{
		"counter-a": NewCounterBehaviorFactory(Counter),
		"counter-b": NewCounterBehaviorFactory(Counter),
		"counter-c": NewCounterBehaviorFactory(Counter),
		"counter-d": NewCounterBehaviorFactory(Counter),
		"counter-e": NewCounterBehaviorFactory(Counter),
		"counter-f": NewCounterBehaviorFactory(Counter),
		"counter-g": NewCounterBehaviorFactory(Counter),
		"counter-h": NewCounterBehaviorFactory(Counter),
	}

	err := env.AddCells(bfm)
	assert.Nil(err, "Add cells.")

	counterH, err := env.CellBehavior("counter-h")
	assert.Nil(err, "Get cell H.")
	assert.NotNil(counterH, "Get cell H.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestRemoveCell tests the removing of cells.
func TestRemoveCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("remove-cell")

	counterA, err := env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))
	assert.Nil(err, "Add cell A.")
	assert.NotNil(counterA, "Add cell A.")
	env.RemoveCell("counter-a")
	assert.False(env.HasCell("counter-a"), "Remove of cell A worked.")
	counterA, err = env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))
	assert.Nil(err, "Add cell A after removal.")
	assert.NotNil(counterA, "Add cell A after removal.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSubscribe tests the subscription of cells.
func TestSubscribe(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("subscribe")

	env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-b", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-c", NewCounterBehaviorFactory(Counter))

	err := env.Subscribe("counter-a", "counter-b")
	assert.Nil(err, "Subscribe cell B to cell A.")
	err = env.Subscribe("counter-a", "counter-c")
	assert.Nil(err, "Subscribe cell C to cell A.")

	err = env.Subscribe("counter-x", "counter-a")
	assert.ErrorMatch(err, `cell "counter-x" does not exist`, "Subscribe cell X to cell A.")
	err = env.Subscribe("counter-a", "counter-x")
	assert.ErrorMatch(err, `cell "counter-x" does not exist`, "Subscribe cell A to cell X.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestUnsubscribe tests the unsubscription of cells.
func TestUnsubscribe(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("unsubscribe")

	env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-b", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-c", NewCounterBehaviorFactory(Counter))

	err := env.Subscribe("counter-a", "counter-b")
	assert.Nil(err, "Subscribe cell B to cell A.")
	err = env.Subscribe("counter-a", "counter-c")
	assert.Nil(err, "Subscribe cell C to cell A.")
	err = env.Unsubscribe("counter-a", "counter-b")
	assert.Nil(err, "Unsubscribe cell B from cell A.")

	err = env.Unsubscribe("counter-x", "counter-a")
	assert.ErrorMatch(err, `cell "counter-x" does not exist`, "Subscribe cell X to cell A.")
	err = env.Unsubscribe("counter-a", "counter-x")
	assert.ErrorMatch(err, `cell "counter-x" does not exist`, "Subscribe cell A to cell X.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSimpleEmit tests the emitting of events.
func TestSimpleEmit(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	env := NewEnvironment("simple-emit")

	env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))

	c, _ := env.Emit("counter-a", NewSimpleEvent("event:1", "data"))

	c.Wait(time.Second)

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSubscribedEmit tests the emitting of events for subscribed cells.
func TestSubscribedEmit(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("subscribed-emit")
	env.AddCell("trace", LogBehaviorFactory)
	env.AddCell("counter-a", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-b", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-c", NewCounterBehaviorFactory(Counter))
	env.AddCell("counter-d", NewCounterBehaviorFactory(Counter))

	env.Subscribe("counter-a", "counter-b")
	env.Subscribe("counter-b", "counter-c", "counter-d")
	env.Subscribe("counter-c", "trace")
	env.Subscribe("counter-d", "trace")

	c, err := env.EmitSimple("counter-a", "event:1", "data")
	assert.Nil(err, "No error during raise.")

	err = c.Wait(time.Second)
	assert.Nil(err, "No error during wait.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestBroadcastBehavior tests the pass-through of events by the
// broadcast behavior.
func TestBroadcastBehavior(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("broadcast-behavior")
	env.AddCell("broadcast", BroadcastBehaviorFactory)
	env.AddCell("collector", CollectorBehaviorFactory)

	env.Subscribe("broadcast", "collector")

	env.EmitSimple("broadcast", "event:1", true)
	env.EmitSimple("broadcast", "event:2", true)
	env.EmitSimple("broadcast", "event:3", true)

	time.Sleep(100 * time.Millisecond)

	b, _ := env.CellBehavior("collector")
	collector := b.(EventCollector)
	events := collector.Events()

	assert.Length(events, 3, "All events received")
}

// TestFilterBehavior tests the filtering of events by the
// filter behavior.
func TestFilterBehavior(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ff := func(e Event) bool {
		if strings.Contains(e.Topic(), "yes") {
			return true
		}
		return false
	}
	env := NewEnvironment("filter-behavior")
	env.AddCell("filter", NewFilteredBroadcastBehaviorFactory(ff))
	env.AddCell("collector", CollectorBehaviorFactory)

	env.Subscribe("filter", "collector")

	env.EmitSimple("filter", "yes", true)
	env.EmitSimple("filter", "no", true)
	env.EmitSimple("filter", "yes", true)

	time.Sleep(100 * time.Millisecond)

	b, _ := env.CellBehavior("collector")
	collector := b.(EventCollector)
	events := collector.Events()

	assert.Length(events, 2, "All events received")
}

// TestSimpleActionBehavior tests the performing of simple actions 
// with events by the simple action behavior.
func TestSimpleActionBehavior(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	saf := func(e Event, emitter EventEmitter) {
		values := e.Payload().([]int)
		result := 0
		switch e.Topic() {
		case "add":
			for _, value := range values {
				result += value
			}
		case "sub":
			for _, value := range values {
				result -= value
			}
		}
		emitter.EmitSimple("result", result)
	}
	env := NewEnvironment("action-behavior")
	env.AddCell("action", NewSimpleActionBehaviorFactory(saf))
	env.AddCell("collector", CollectorBehaviorFactory)

	env.Subscribe("action", "collector")

	env.EmitSimple("action", "add", []int{1, 2, 3, 4, 5})
	env.EmitSimple("action", "sub", []int{1, 2, 3, 4, 5})
	env.EmitSimple("action", "add", []int{1, 1})

	time.Sleep(100 * time.Millisecond)

	b, _ := env.CellBehavior("collector")
	collector := b.(EventCollector)
	events := collector.Events()

	assert.Equal(events[0].Payload().(int), 15, "First result is ok.")
	assert.Equal(events[1].Payload().(int), -15, "Second result is ok.")
	assert.Equal(events[2].Payload().(int), 2, "Third result is ok.")
}

// TestCounterBehavior tests the counting of events 
// by the counter behavior.
func TestCounterBehavior(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cf := func(e Event) []string {
		return []string{e.Topic()}
	}
	env := NewEnvironment("counter-behavior")
	env.AddCell("counter", NewCounterBehaviorFactory(cf))
	env.AddCell("collector", CollectorBehaviorFactory)

	env.Subscribe("counter", "collector")

	env.EmitSimple("counter", "a", true)
	env.EmitSimple("counter", "b", true)
	env.EmitSimple("counter", "c", true)
	env.EmitSimple("counter", "a", true)

	time.Sleep(100 * time.Millisecond)

	b, _ := env.CellBehavior("collector")
	collector := b.(EventCollector)
	events := collector.Events()

	assert.Equal(events[0].Payload().(int64), int64(1), "First result is ok.")
	assert.Equal(events[1].Payload().(int64), int64(1), "Second result is ok.")
	assert.Equal(events[2].Payload().(int64), int64(1), "Third result is ok.")
	assert.Equal(events[3].Payload().(int64), int64(2), "Fourth result is ok.")
}

// EOF
