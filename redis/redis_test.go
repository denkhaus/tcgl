// Tideland Common Go Library - Redis - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"code.google.com/p/tcgl/asserts"
	"code.google.com/p/tcgl/monitoring"
	"testing"
	"time"
)

//--------------------
// HELPER
//--------------------

// hashableTestType is a simple type implementing the
// Hashable interface.
type hashableTestType struct {
	a string
	b int64
	c bool
	d float64
}

// GetHash returns the fields as hash.
func (htt *hashableTestType) GetHash() Hash {
	h := NewHash()

	h.Set("hashable:field:a", htt.a)
	h.Set("hashable:field:b", htt.b)
	h.Set("hashable:field:c", htt.c)
	h.Set("hashable:field:d", htt.d)

	return h
}

// SetHash sets the fields from a hash.
func (htt *hashableTestType) SetHash(h Hash) {
	htt.a, _ = h.String("hashable:field:a")
	htt.b, _ = h.Int64("hashable:field:b")
	htt.c, _ = h.Bool("hashable:field:c")
	htt.d, _ = h.Float64("hashable:field:d")
}

//--------------------
// TESTS
//--------------------

func TestConnection(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	// Connection commands.
	assert.Equal(rd.Command("echo", "Hello, World!").ValueAsString(), "Hello, World!", "Echo of a string.")
	assert.Equal(rd.Command("ping").ValueAsString(), "PONG", "Playing ping pong.")
}

func TestSimpleSingleValue(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rs := rd.Command("del", "single-value")
	assert.True(rs.IsOK(), "'del' is ok.")
	rs = rd.Command("set", "single-value", "Hello, World!")
	assert.True(rs.IsOK(), "'set' is ok.")
	rs = rd.Command("get", "single-value")
	assert.True(rs.IsOK(), "'get' is ok.")
	assert.False(rs.IsMulti(), "'get' returned no multi-result-set.")
	assert.Equal(rs.Command(), "get", "Command has been 'get'.")
	assert.Equal(rs.ValueCount(), 1, "'get' returned one value.")
	assert.Equal(rs.Value().String(), "Hello, World!", "'get' returned the right value.")
}

func TestSimpleMultipleValues(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "multiple-value:1")
	rd.Command("del", "multiple-value:2")
	rd.Command("del", "multiple-value:3")
	rd.Command("del", "multiple-value:4")
	rd.Command("del", "multiple-value:5")

	rd.Command("set", "multiple-value:1", "one")
	rd.Command("set", "multiple-value:2", "two")
	rd.Command("set", "multiple-value:3", "three")
	rd.Command("set", "multiple-value:4", "four")
	rd.Command("set", "multiple-value:5", "five")

	rs := rd.Command("mget", "multiple-value:1", "multiple-value:2", "multiple-value:3", "multiple-value:4", "multiple-value:5")
	assert.True(rs.IsOK(), "'mget' is ok.")
	assert.False(rs.IsMulti(), "'mget' returned no multi-result-set.")
	assert.Equal(rs.Command(), "mget", "Command has been 'mget'.")
	assert.Equal(rs.ValueCount(), 5, "'mget' returned five value.")
	assert.Equal(rs.ValueAt(0).String(), "one", "'mget' returned the right first value.")	
	assert.Equal(rs.ValueAt(1).String(), "two", "'mget' returned the right second value.")	
	assert.Equal(rs.ValueAt(2).String(), "three", "'mget' returned the right third value.")	
	assert.Equal(rs.ValueAt(3).String(), "four", "'mget' returned the right fourth value.")	
	assert.Equal(rs.ValueAt(4).String(), "five", "'mget' returned the right fifth value.")	
}

func TestHash(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "hash:manual")
	rd.Command("del", "hash:hashable")

	// Manual hash usage.
	rd.Command("hset", "hash:manual", "field:1", "one")
	rd.Command("hset", "hash:manual", "field:2", "two")

	rs := rd.Command("hget", "hash:manual", "field:1")
	assert.True(rs.IsOK(), "'hget' is ok.")
	assert.Equal(rs.ValueCount(), 1, "'hget' returned one value.")
	assert.Equal(rs.Value().String(), "one", "'hget' returned the right value.")
	rs = rd.Command("hgetall", "hash:manual")
	assert.True(rs.IsOK(), "'hgetall' is ok.")
	assert.Equal(rs.ValueCount(), 4, "'hgetall' returned four values (fields and values).")
	assert.Equal(rs.ValueAt(0).String(), "field:1", "'hgetall' returned the right first value.")	
	assert.Equal(rs.ValueAt(1).String(), "one", "'hgetall' returned the right second value.")	
	assert.Equal(rs.ValueAt(2).String(), "field:2", "'hgetall' returned the right third value.")	
	assert.Equal(rs.ValueAt(3).String(), "two", "'hgetall' returned the right fourth value.")

	// Use the Hash type and the Hashable interface.
	h := rd.Command("hgetall", "hash:manual").Hash()
	assert.Equal(h.Len(), 2, "Manual hash has the right len.")
	v, err := h.String("field:1")
	assert.Nil(err, "Reading 'field:1' is ok.")
	assert.Equal(v, "one", "Hash field 'field:1' has right value.")
	v, err = h.String("field:2")
	assert.Nil(err, "Reading 'field:2' is ok.")
	assert.Equal(v, "two", "Hash field 'field:2' has right value.")
	v, err = h.String("field:not-existing")
	assert.ErrorMatch(err, `redis: invalid key "field:not-existing"`, "Right error for not-existing field.")

	htIn := hashableTestType{"foo \"bar\" yadda", 4711, true, 8.15}
	rd.Command("hmset", "hash:hashable", htIn.GetHash())
	rd.Command("hincrby", "hash:hashable", "hashable:field:b", 289)

	htOut := hashableTestType{}
	htOut.SetHash(rd.Command("hgetall", "hash:hashable").Hash())
	assert.Equal(htOut.a, "foo \"bar\" yadda", "Hash field 'a' is ok.")
	assert.Equal(htOut.b, int64(5000), "Hash field 'b' is ok.")
	assert.True(htOut.c, "Hash field 'c' is ok.")
	assert.Equal(htOut.d, 8.15, "Hash field 'd' is ok.")
}

func TestFuture(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "future")

 	fut := rd.AsyncCommand("rpush", "future", "one", "two", "three", "four", "five")
 	rs := fut.ResultSet()
 	assert.True(rs.IsOK(), "Future result is ok.")
 	v, err := rs.ValueAsInt()
	assert.Nil(err, "Future value is an integer.")
	assert.Equal(v, 5, "The returned future value is 5.")
}

func TestStringMap(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "string:map")

 	mapIn := map[string]string{
 		"A": "1",
 		"B": "2",
 		"C": "3",
 		"D": "4",
 		"E": "5",
 	}
 	rs := rd.Command("set", "string:map", mapIn)
 	assert.True(rs.IsOK(), "Setting a string map is ok.")

 	rs = rd.Command("get", "string:map")
 	assert.True(rs.IsOK(), "Getting a string map is ok.")
 	mapOut := rs.Value().StringMap()
 	assert.Length(mapOut, 5, "Length of the retrieved string map is ok.")
 	assert.Equal(mapOut, mapIn, "Retrieval string map is ok.")
}

func TestStringSlice(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "string:slice")

 	sliceIn := []string{"A", "B", "C", "D", "E"}
 	rs := rd.Command("set", "string:slice", sliceIn)
 	assert.True(rs.IsOK(), "Setting a string slice is ok.")

 	rs = rd.Command("get", "string:slice")
 	assert.True(rs.IsOK(), "Getting a string slice is ok.")
 	sliceOut := rs.Value().StringSlice()
 	assert.Length(sliceOut, 5, "Length of the retrieved string slice is ok.")
 	assert.Equal(sliceOut, sliceIn, "Retrieval string slice is ok.")
}

func TestMultiCommand(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "multi-command:1")
	rd.Command("del", "multi-command:2")
	rd.Command("del", "multi-command:3")
	rd.Command("del", "multi-command:4")
	rd.Command("del", "multi-command:5")

 	rs := rd.MultiCommand(func(mc *MultiCommand) {
 		mc.Command("set", "multi-command:1", "1")
 		mc.Command("set", "multi-command:1", "2")
 		mc.Discard()
  		mc.Command("set", "multi-command:1", "one")
 		mc.Command("set", "multi-command:2", "two")
 		mc.Command("set", "multi-command:3", "three")
 		mc.Command("set", "multi-command:4", "four")
 		mc.Command("set", "multi-command:5", "five")

  		mc.Command("get", "multi-command:3")
 	})
 	assert.True(rs.IsOK(), "Executing the multi-command has been ok.")
 	assert.Equal(rs.ResultSetCount(), 6, "Multi-command returned 6 result sets.")
 	assert.Equal(rs.ResultSetAt(5).ValueAsString(), "three", "Sixth result set contained right value 'three'.")
}

func TestBlockingPop(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "queue")

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			rs := rd.Command("lpush", "queue", i)
			assert.True(rs.IsOK(), "'lpush' into queue has been ok.")
		}
	}()

	for i := 0; i < 10; i++ {
		rs := rd.Command("brpop", "queue", 5)
		assert.True(rs.IsOK(), "'brpop' from queue has been ok.")
		assert.Equal(rs.ValueAt(0).String(), "queue", "Right 'queue' has been returned.")
		v, err := rs.ValueAt(1).Int()
		assert.Nil(err, "No error retrieving the integer value.")
		assert.Equal(v, i, "Popped value has been right.")
	}

	rs := rd.Command("brpop", "queue", 1)
	assert.ErrorMatch(rs.Error(), "redis: timeout after .*", "'brpop' timed out.")
	assert.Assignable(rs.Error(), &TimeoutError{}, "Error has correct type.")
}

func TestPubSub(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{})

	sub, err := rd.Subscribe("pubsub:1", "pubsub:2", "pubsub:3")
	assert.Nil(err, "No error when subscribing.")
	sub.Subscribe("pubsub:pattern:*")

	go func() {
		time.Sleep(50 * time.Millisecond)
		rd.Publish("pubsub:1", "foo")
		rd.Publish("pubsub:2", "bar")
		rd.Publish("pubsub:3", "baz")
		rd.Publish("pubsub:pattern:yadda", "yadda")
	}()

	// Check value receiving.
	value := <-sub.Values()
	assert.Equal(value.Channel, "pubsub:1", "First value channel has been ok.")
	assert.Equal(value.Value.String(), "foo", "First value has been ok.")

	value = <-sub.Values()
	assert.Equal(value.Channel, "pubsub:2", "Second value channel has been ok.")
	assert.Equal(value.Value.String(), "bar", "Second value has been ok.")

	value = <-sub.Values()
	assert.Equal(value.Channel, "pubsub:3", "Third value channel has been ok.")
	assert.Equal(value.Value.String(), "baz", "Third value has been ok.")

	value = <-sub.Values()
	assert.Equal(value.Channel, "pubsub:pattern:yadda", "Fourth value channel has been ok.")
	assert.Equal(value.ChannelPattern, "pubsub:pattern:*", "Fourth value channel pattern has been ok.")
	assert.Equal(value.Value.String(), "yadda", "Fourth value has been ok.")

	// Check no more receiving.
	select {
	case value = <-sub.Values():
		assert.Nil(value, "Nothing expected here.")
	case <-time.After(200 * time.Millisecond):
		assert.True(true, "Timeout like expected.")
	}

	// Check unsubscribing.
	sub.Unsubscribe("pubsub:3")

	go func() {
		time.Sleep(50 * time.Millisecond)
		rd.Publish("pubsub:3", "foobar")
	}()

	select {
	case value = <-sub.Values():
		assert.Nil(value, "Nothing expected here.")
	case <-time.After(200 * time.Millisecond):
		assert.True(true, "Timeout like expected.")
	}

	// Check subscription stopping.
	sub.Stop()

	select {
	case _, ok := <-sub.Values():
		assert.False(ok, "Expected signalling of closed channel.")
	case <-time.After(200 * time.Millisecond):
		assert.False(true, "Timeout not expected here.")
	}
}

// Test illegal databases.
func TestIllegalDatabases(t *testing.T) {
	if testing.Short() {
		return
	}

	// Test illegal database number.
	assert := asserts.NewTestingAsserts(t, true)
	rd := NewRedisDatabase(Configuration{Database: 4711})

	rs := rd.Command("ping")
	assert.ErrorMatch(rs.Error(), "redis: invalid DB index", "Error message for invalid DB index is ok.")

	// Test illegal network address.
	rd = NewRedisDatabase(Configuration{Address: "192.168.100.100:12345"})

	rs = rd.Command("ping")
	assert.ErrorMatch(rs.Error(), "redis: connection has a an error: dial tcp .*: i/o timeout", "Error message for invalid address is ok.")
	assert.Assignable(rs.Error(), &ConnectionError{}, "Error has correct type.")
}

// Test measuring (pure output).
func TestMeasuring(t *testing.T) {
	monitoring.MeasuringPointsPrintAll()
	monitoring.StaySetVariablesPrintAll()
}

// EOF
