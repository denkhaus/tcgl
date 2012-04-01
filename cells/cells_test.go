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
	"code.google.com/p/tcgl/asserts"
	"code.google.com/p/tcgl/monitoring"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

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
	}()
	err = c.Wait(50 * time.Millisecond)
	assert.Nil(err, "No timeout during wait.")
	assert.Equal(step, 8, "Right increments and decrements before end of waiting.")

	// Check iteration and value count.
	valueCount := 0
	c.Do(func(key string, value interface{}) {
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

// TestAssignments tests the assignments of one to several ids.
func TestAssignments(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	a := make(assignments)
	assert.Empty(a.all("foo"), "Nothing assigned to 'foo' yet.")
	a.add("foo", "bar")
	assert.Equal(a.all("foo"), []string{"bar"}, "'bar' assigned to 'foo'.")
	a.add("foo", "baz")
	assert.Equal(a.all("foo"), []string{"bar", "baz"}, "'bar' and 'baz' assigned to 'foo'.")
	a.add("foo", "bar")
	assert.Equal(a.all("foo"), []string{"bar", "baz"}, "Second 'bar' not additionally assigned to 'foo'.")
	a.add("foo", "yadda")
	assert.Equal(a.all("foo"), []string{"bar", "baz", "yadda"}, "'bar', 'baz' and 'yadda' assigned to 'foo'.")

	a.remove("foo", "bar")
	assert.Equal(a.all("foo"), []string{"baz", "yadda"}, "'baz' and 'yadda' assigned to 'foo'.")
	a.remove("foo", "bar")
	assert.Equal(a.all("foo"), []string{"baz", "yadda"}, "Can't remove 'bar' twice.")

	// Remove should raise no error.
	a.remove("foobar", "barfoo")

	a.drop("foo")
	assert.Empty(a.all("foo"), "Nothing assigned to 'foo' yet.")

	a.add("foo", "bar", "baz", "yadda")
	assert.Length(a.all("foo"), 3, "All three are assigned to 'foo'.")
	oa := make(assignments)
	oa.add("bar", "dummy", "foo")
	oa.add("baz", "dummy", "foo", "yadda")
	oa.add("yadda", "foo")
	a.dropAll("foo", oa)
	assert.Equal(oa.all("bar"), []string{"dummy"}, "Other assignment 'bar' is OK.")
	assert.Equal(oa.all("baz"), []string{"dummy", "yadda"}, "Other assignment 'baz' is OK.")
	assert.Equal(oa.all("yadda"), []string{}, "Other assignment 'yadda' is OK.")
}

// TestEnvironment tests the envirment creation and shutdown.
func TestEnvironment(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	assert.NotNil(env, "Environment is created.")
	assert.Equal(env.id, "test", "Environment id is 'test'.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestAddCell tests the adding of cells.
func TestAddCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	ba, err := env.AddCell("ba", newDummyBehavior(), 5)
	assert.Nil(err, "Add cell A.")
	assert.NotNil(ba, "Add cell A.")
	dba := ba.(*dummyBehavior)
	assert.Equal(dba.initCounter, 1, "Behavior A has been initialized.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
	assert.Equal(dba.stopCounter, 1, "Behavior A has been stopped.")
}

// TestDoubleAddCell tests the adding of cells with
// the same id.
func TestDoubleAddCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	ba, err := env.AddCell("ba", newDummyBehavior(), 5)
	assert.Nil(err, "Add cell A.")
	assert.NotNil(ba, "Add cell A.")
	ba, err = env.AddCell("ba", newDummyBehavior(), 5)
	assert.Nil(ba, "Cell A can't be added twice.")
	assert.ErrorMatch(err, `cell behavior with id "ba" already added`, "Cell A can't be added twice.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestRemoveCell tests the removing of cells.
func TestRemoveCell(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	ba, err := env.AddCell("ba", newDummyBehavior(), 5)
	assert.Nil(err, "Add cell A.")
	assert.NotNil(ba, "Add cell A.")
	err = env.RemoveCell("ba")
	assert.Nil(err, "Remove cell A.")
	dba := ba.(*dummyBehavior)
	assert.Equal(dba.stopCounter, 1, "Behavior A has been stopped.")
	ba, err = env.AddCell("ba", newDummyBehavior(), 5)
	assert.Nil(err, "Add cell A after removal.")
	assert.NotNil(ba, "Add cell A after removal.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSubscribe tests the subscription of cells.
func TestSubscribe(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	env.AddCell("ba", newDummyBehavior(), 5)
	env.AddCell("bb", newDummyBehavior(), 5)
	env.AddCell("bc", newDummyBehavior(), 5)

	err := env.Subscribe("ba", "bb")
	assert.Nil(err, "Subscribe BB to BA.")
	err = env.Subscribe("ba", "bc")
	assert.Nil(err, "Subscribe BC to BA.")

	err = env.Subscribe("bx", "ba")
	assert.ErrorMatch(err, `emitter cell "bx" does not exist`, "Subscribe BX to BA.")
	err = env.Subscribe("ba", "bx")
	assert.ErrorMatch(err, `subscriber cell "bx" does not exist`, "Subscribe BA to BX.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestUnsubscribe tests the unsubscription of cells.
func TestUnsubscribe(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	env.AddCell("ba", newDummyBehavior(), 5)
	env.AddCell("bb", newDummyBehavior(), 5)
	env.AddCell("bc", newDummyBehavior(), 5)

	err := env.Subscribe("ba", "bb")
	assert.Nil(err, "Subscribe BB to BA.")
	err = env.Subscribe("ba", "bc")
	assert.Nil(err, "Subscribe BC to BA.")
	err = env.Unsubscribe("ba", "bb")
	assert.Nil(err, "Unsubscribe BB from BA.")

	err = env.Unsubscribe("bx", "ba")
	assert.ErrorMatch(err, `emitter cell "bx" does not exist`, "Subscribe BX to BA.")
	err = env.Unsubscribe("ba", "bx")
	assert.ErrorMatch(err, `subscriber cell "bx" does not exist`, "Subscribe BA to BX.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSimpleRaise tests the raising of events.
func TestSimpleRaise(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	ba, _ := env.AddCell("ba", newDummyBehavior(), 5)
	dba := ba.(*dummyBehavior)

	c, _ := env.Raise("ba", NewSimpleEvent("event:1", "data"))

	c.Wait(10 * time.Millisecond)
	assert.Length(dba.events, 1, "Number of processed events.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSubscribedRaise tests the raising of events for subscribed cells.
func TestSubscribedRaise(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	env.AddCell("trace", NewLogBehavior(), 1)
	env.AddCell("ba", newDummyBehavior(), 5)
	env.AddCell("bb", newDummyBehavior(), 5)
	bc, _ := env.AddCell("bc", newDummyBehavior(), 5)
	dbc := bc.(*dummyBehavior)
	bd, _ := env.AddCell("bd", newDummyBehavior(), 5)
	dbd := bd.(*dummyBehavior)

	env.Subscribe("ba", "bb")
	env.Subscribe("bb", "bc", "bd")
	env.Subscribe("bc", "trace")
	env.Subscribe("bd", "trace")

	c, err := env.Raise("ba", NewSimpleEvent("event:1", "data"))
	assert.Nil(err, "No error during raise.")

	c.Wait(10 * time.Millisecond)

	assert.Length(dbc.events, 1, "Number of processed events.")
	assert.Length(dbd.events, 1, "Number of processed events.")

	bcv, err := c.Get("bc")
	assert.Nil(err, "No error retrieving the value 'bc'.")
	assert.Equal(bcv, 1, "Right value of 'bc'.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestBroadcast tests the broadcast behavior.
func TestBroadcast(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	env.AddCell("input", NewBroadcastBehavior(), 1)
	env.AddCell("trace", NewLogBehavior(), 1)
	env.AddCell("ba", newDummyBehavior(), 1)

	env.Subscribe("input", "trace", "ba")

	env.Raise("input", NewSimpleEvent("event:broadcast", "data"))

	time.Sleep(10 * time.Millisecond)

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestPool tests the pool behavior.
func TestPool(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env := NewEnvironment("test")
	bf := func() Behavior {
		return newDummyBehavior()
	}
	env.AddCell("pool", NewPoolBehavior(bf, 10), 1)
	ps, _ := env.AddCell("pool-sub", newDummyBehavior(), 1)
	dbps := ps.(*dummyBehavior)

	env.Subscribe("pool", "pool-sub")

	for i := 0; i < 20; i++ {
		env.RaiseSimpleEvent("pool", "event:pool", true)
	}

	time.Sleep(10 * time.Millisecond)
	assert.Length(dbps.events, 20, "Number of processed events.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestSimpleAction tests the simple acion behavior.
func TestSimpleAction(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := ""
	action := func(e Event, emitChan EventChannel) {
		data = fmt.Sprintf("Payload: %v", e.Payload())
	}
	env := NewEnvironment("test")
	env.AddCell("simple", SimpleActionFunc(action), 1)

	env.Raise("simple", NewSimpleEvent("event:action", "data"))

	time.Sleep(10 * time.Millisecond)
	assert.Equal(data, "Payload: data", "Received data.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestFilteredSimpleAction tests the filtered simple acion behavior.
func TestFilteredSimpleAction(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := ""
	filter := func(e Event) bool {
		if e.Payload() == "data" {
			return true
		}
		return false
	}
	action := func(e Event, emitChan EventChannel) {
		data = fmt.Sprintf("Payload: %v", e.Payload())
	}
	env := NewEnvironment("test")
	env.AddCell("filter", NewFilteredSimpleActionBehavior(filter, action), 1)

	env.Raise("filter", NewSimpleEvent("event:action", "foo"))
	env.Raise("filter", NewSimpleEvent("event:action", "data"))
	env.Raise("filter", NewSimpleEvent("event:action", "bar"))

	time.Sleep(10 * time.Millisecond)
	assert.Equal(data, "Payload: data", "Received data.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestCounter tests the counter behavior.
func TestCounter(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cf := func(e Event) []string {
		return strings.Split(e.Topic(), ":")
	}
	env := NewEnvironment("test")
	env.AddCell("counter", NewCounterBehavior(cf), 1)
	env.AddCell("trace", NewLogBehavior(), 1)
	env.Subscribe("counter", "trace")

	env.RaiseSimpleEvent("counter", "foo:bar:baz", nil)
	env.RaiseSimpleEvent("counter", "bar:baz", nil)
	env.RaiseSimpleEvent("counter", "baz", nil)

	time.Sleep(10 * time.Millisecond)

	b, _ := env.Cell("counter")
	cb := b.(*CounterBehavior)

	assert.Equal(cb.Counter("foo"), 1, "Counter 'foo'.")
	assert.Equal(cb.Counter("bar"), 2, "Counter 'bar'.")
	assert.Equal(cb.Counter("baz"), 3, "Counter 'baz'.")
	assert.Equal(cb.Counter("yadda"), -1, "Counter 'yadda'.")

	err := env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestTicker tests the management of periodic events.
func TestTicker(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cf := func(e Event) []string {
		return []string{e.Topic()}
	}
	env := NewEnvironment("test")
	env.AddCell("counter", NewCounterBehavior(cf), 1)
	env.AddTicker("alpha", "counter", 5 * time.Millisecond)
	env.AddTicker("beta", "counter", 2 * time.Millisecond)

	// Test standard ticking.
	time.Sleep(500 * time.Millisecond)

	b, _ := env.Cell("counter")
	cb := b.(*CounterBehavior)
	va := float64(cb.Counter("ticker(alpha)"))
	vb := float64(cb.Counter("ticker(beta)"))

	assert.About(va, 100.0, 10.0, "Ticker Alpha should emit about 100 events during this time.")
	assert.About(vb, 250.0, 50.0, "Ticker Beta should emit about 250 events during this time.")

	// Test double add.
	err := env.AddTicker("alpha", "counter", 5 * time.Millisecond)
	assert.ErrorMatch(err, `ticker with id "alpha" already added`, "Can't add a ticker with an existing id.")

	// Test remove non-existing ticker.
	err = env.RemoveTicker("not-there")
	assert.ErrorMatch(err, `ticker with id "not-there" does not exist`, "Can't remove a not existing ticker.")

	// Test remove ticker.
	err = env.RemoveTicker("alpha")
	assert.Nil(err, "Ticker removal should work.")
	err = env.AddTicker("alpha", "counter", 5 * time.Millisecond)
	assert.Nil(err, "And so also a new adding of a now free ticker.")

	err = env.Shutdown()
	assert.Nil(err, "No error during shutdown.")
}

// TestMonitoring just prints the measuring values.
func TestMonitoring(t *testing.T) {
	monitoring.MeasuringPointsPrintAll()
}

//--------------------
// HELPERS
//--------------------

type rdata struct {
	r interface{}
	e Event
}

type dummyBehavior struct {
	id          string
	initCounter int
	events      []Event
	rdatas      []*rdata
	stopCounter int
}

func newDummyBehavior() *dummyBehavior {
	return &dummyBehavior{"", 0, []Event{}, []*rdata{}, 0}
}

func (d *dummyBehavior) Init(env *Environment, id string) error {
	d.id = id
	d.initCounter++
	return nil
}

func (d *dummyBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	ne := NewSimpleEvent(e.Topic()+":"+d.id, e.Payload())

	emitChan <- ne

	d.events = append(d.events, e)

	e.Context().Set(d.id, len(d.events))
}

func (d *dummyBehavior) Recover(r interface{}, e Event) {
	d.rdatas = append(d.rdatas, &rdata{r, e})
}

func (d *dummyBehavior) Stop() error {
	d.stopCounter++
	return nil
}

// EvenFilter filters even integer.
func EvenFilter(e Event) bool {
	if d, ok := e.Payload().(int); ok {
		return d%2 == 0
	}
	return false
}

// OddFilter filters even integer.
func OddFilter(e Event) bool {
	if d, ok := e.Payload().(int); ok {
		return d%2 != 0
	}
	return false
}

// SeparatorAction seperates integer events in odd and even.
func SeparatorAction(e Event, ec EventChannel) {
	if d, ok := e.Payload().(int); ok {
		if d%2 == 0 {
			ec <- NewSimpleEvent("even", d)
		} else {
			ec <- NewSimpleEvent("odd", d)
		}
	}
}

// ItoaAction maps an integer to an asciii string.
func ItoaAction(e Event, ec EventChannel) {
	if d, ok := e.Payload().(int); ok {
		ec <- NewSimpleEvent("itoa:"+e.Topic(), strconv.Itoa(d))
	} else {
		ec <- NewSimpleEvent("illegal-type", e)
	}
}

// EOF
