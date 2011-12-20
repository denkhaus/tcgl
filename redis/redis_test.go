// Tideland Common Go Library - Redis - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"rand"
	"testing"
	"time"
	"tcgl.googlecode.com/hg/monitoring"
	"tcgl.googlecode.com/hg/util"
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
	htt.a = h.String("hashable:field:a")
	htt.b = h.Int64("hashable:field:b")
	htt.c = h.Bool("hashable:field:c")
	htt.d = h.Float64("hashable:field:d")
}

//--------------------
// TESTS
//--------------------

// Test connection commands.
func TestConnection(t *testing.T) {
	util.Debug("Connection ...")

	rd := NewRedisDatabase(Configuration{})

	// Connection commands.
	if rd.Command("echo", "Hello, World!").ValueAsString() != "Hello, World!" {
		t.Errorf("Invalid echo result, expected 'Hello, World!'!")
	}

	if rd.Command("ping").ValueAsString() != "PONG" {
		t.Errorf("Can't ping the server!")
	}
}

// Test simple value commands.
func TestSimpleValue(t *testing.T) {
	util.Debug("Simple value ...")

	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "simple:string")
	rd.Command("del", "simple:int")
	rd.Command("del", "simple:float64")
	rd.Command("del", "simple:bytes")
	rd.Command("del", "simple:bits")
	rd.Command("del", "simple:nx")
	rd.Command("del", "simple:range")

	// Simple value commands.
	rd.Command("set", "simple:string", "Hello,")
	rd.Command("append", "simple:string", " World")

	if rd.Command("get", "simple:string").ValueAsString() != "Hello, World" {
		t.Errorf("Invalid append result for simple:string, expected 'Hello, World'!")
	}

	// ---
	rd.Command("set", "simple:int", 10)

	if v := rd.Command("incr", "simple:int").ValueAsInt(); v != 11 {
		t.Errorf("Got '%v', expected '11'!", v)
	}

	// ---
	rd.Command("set", "simple:float64", 47.11)

	if v := rd.Command("get", "simple:float64").Value().Float64(); v != float64(47.11) {
		t.Errorf("Got '%v', expected '47.11' as float64!", v)
	}

	// ---
	bytesIn := make([]byte, 128)

	for i := 0; i < 128; i++ {
		bytesIn[i] = byte(rand.Intn(255))
	}

	rd.Command("set", "simple:bytes", bytesIn)

	bytesOut := rd.Command("get", "simple:bytes").Value().Bytes()

	for i, b := range bytesOut {
		if bytesIn[i] != b {
			t.Errorf("Got '%v', expected '%v'!", b, bytesIn[i])
		}
	}

	// ---
	rd.Command("setbit", "simple:bits", 0, 1)
	rd.Command("setbit", "simple:bits", 1, 0)

	if !rd.Command("getbit", "simple:bits", 0).ValueAsBool() || rd.Command("getbit", "simple:bits", 1).ValueAsBool() {
		t.Errorf("Bit setting or getting went wrong!")
		t.Errorf("Got '%v' for bit 0.", rd.Command("getbit", "simple:bits", 0).ValueAsBool())
		t.Errorf("Got '%v' for bit 1.", rd.Command("getbit", "simple:bits", 1).ValueAsBool())
	}

	// ---
	if rd.Command("get", "non:existing:key").IsOK() {
		t.Errorf("Expected a failure for non-existing key!")
	}

	if rd.Command("exists", "non:existing:key").ValueAsBool() {
		t.Errorf("Expected 'false' for non-existing key!")
	}

	// ---
	if !rd.Command("setnx", "simple:nx", "Test").ValueAsBool() {
		t.Errorf("Setting of not-existing key failed!")
	}

	if rd.Command("setnx", "simple:nx", "Test").ValueAsBool() {
		t.Errorf("Overwriting of existing key!")
	}

	// ---
	if rd.Command("setrange", "simple:range", 10, "Test").ValueAsInt() != 14 {
		t.Errorf("Range setting has the wrong length!")
	}

	if rd.Command("getrange", "simple:range", 11, 12).ValueAsString() != "es" {
		t.Errorf("Range getting has the wrong value!")
	}
}

// Test multi-key commands.
func TestMultiple(t *testing.T) {
	util.Debug("Multiple ...")

	rd := NewRedisDatabase(Configuration{})

	// Set values first.
	rd.Command("set", "multiple:a", "a")
	rd.Command("set", "multiple:b", "b")
	rd.Command("set", "multiple:c", "c")

	if v := rd.Command("mget", "multiple:a", "multiple:b", "multiple:c"); v == nil {
		t.Errorf("Multiple get has an error!")
	}

	h := &hashableTestType{"multi", 4711, true, 3.141}

	rd.Command("mset", h)

	if v := rd.Command("mget", "hashable:field:a", "hashable:field:c").Values(); len(v) != 2 {
		t.Errorf("Multiple get after hashable has wrong len '%v'!", len(v))
	}
}

// Test hash accessing.
func TestHash(t *testing.T) {
	util.Debug("Hash ...")

	rd := NewRedisDatabase(Configuration{})

	// Single field values.
	rd.Command("hset", "hash:bool", "true:1", 1)
	rd.Command("hset", "hash:bool", "true:2", true)
	rd.Command("hset", "hash:bool", "true:3", "T")
	rd.Command("hset", "hash:bool", "false:1", 0)
	rd.Command("hset", "hash:bool", "false:2", false)
	rd.Command("hset", "hash:bool", "false:3", "FALSE")

	if !rd.Command("hget", "hash:bool", "true:1").ValueAsBool() {
		t.Errorf("hash:bool true:1 is not true!")
	}

	if rd.Command("hget", "hash:bool", "false:2").ValueAsBool() {
		t.Errorf("hash:bool false:2 is not false, it's '%v'!", rd.Command("hget", "hash:bool", "false:2").ValueAsString())
	}

	// ---
	ha := rd.Command("hgetall", "hash:bool").Hash()

	if ha.Len() != 6 {
		t.Errorf("Hash size of hash:bool is not 6, it's '%v'!", ha.Len())
	}

	if ha.Bool("false:3") {
		t.Errorf("hash:bool false:3 is not false, it's '%v'!", ha.String("false:3"))
	}

	// ---
	hb := hashableTestType{"foo \"bar\" yadda", 4711, true, 8.15}

	rd.Command("hmset", "hashable", hb.GetHash())
	rd.Command("hincrby", "hashable", "hashable:field:b", 289)

	hb = hashableTestType{}

	hb.SetHash(rd.Command("hgetall", "hashable").Hash())

	if hb.a != "foo \"bar\" yadda" || hb.b != 5000 || !hb.c || hb.d != 8.15 {
		t.Errorf("At leas one of the hashable fields is wrong! '%v'", hb)
	}
}

// Test list commands.
func TestList(t *testing.T) {
	util.Debug("List ...")

	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "list:a")
	rd.Command("del", "list:b")
	rd.Command("del", "list:c")

	// Push values into list.
	rd.Command("rpush", "list:a", "one")
	rd.Command("rpush", "list:a", "two")
	rd.Command("rpush", "list:a", "three")
	rd.Command("rpush", "list:a", "four")
	rd.Command("rpush", "list:a", "five")
	rd.Command("rpush", "list:a", "six")
	rd.Command("rpush", "list:a", "seven")
	rd.Command("rpush", "list:a", "eight")
	rd.Command("rpush", "list:a", "nine")

	if l := rd.Command("llen", "list:a").ValueAsInt(); l != 9 {
		t.Errorf("Length of list:a is not 9, it's '%v'!", l)
	}

	if v := rd.Command("lpop", "list:a").ValueAsString(); v != "one" {
		t.Errorf("Left pop in list:a did not returned 'one', it returned '%v'!", v)
	}

	// ---
	if vs := rd.Command("lrange", "list:a", 3, 6).Values(); vs != nil && vs[0].String() != "five" {
		t.Errorf("Range '%v' is wrong!", vs)
	}

	rd.Command("ltrim", "list:a", 0, 3)

	if l := rd.Command("llen", "list:a").ValueAsInt(); l != 4 {
		t.Errorf("Trimming the list:a didn't worked!")
	}

	// ---
	rd.Command("rpoplpush", "list:a", "list:b")

	if rd.Command("lindex", "list:b", 4711).IsOK() {
		t.Errorf("Oops, expected error for invalid list index!")
	}

	if v := rd.Command("lindex", "list:b", 0).ValueAsString(); v != "five" {
		t.Errorf("Right pop left push didn't worked, value has the value '%v'!", v)
	}

	// ---
	rd.Command("rpush", "list:c", 1)
	rd.Command("rpush", "list:c", 2)
	rd.Command("rpush", "list:c", 3)
	rd.Command("rpush", "list:c", 4)
	rd.Command("rpush", "list:c", 5)

	if v := rd.Command("lpop", "list:c").ValueAsInt(); v != 1 {
		t.Errorf("Value of list:c has the wrong value '%v'!", v)
	}
}

// Test set commands.
func TestSet(t *testing.T) {
	util.Debug("Set ...")

	rd := NewRedisDatabase(Configuration{})

	rd.Command("del", "set:a")
	rd.Command("del", "set:b")

	// Add values to the set.
	rd.Command("sadd", "set:a", 1)
	rd.Command("sadd", "set:a", 2)
	rd.Command("sadd", "set:a", 3)
	rd.Command("sadd", "set:a", 4)
	rd.Command("sadd", "set:a", 5)
	rd.Command("sadd", "set:a", 4)
	rd.Command("sadd", "set:a", 3)

	if c := rd.Command("scard", "set:a").ValueAsInt(); c != 5 {
		t.Errorf("Set cardinality is not 5, it's '%v'!", c)
	}

	if rd.Command("sismember", "set:a", Pack(4)).ValueAsBool() {
		t.Errorf("4 is not as expected member of the set!")
	}
}

// Test asynchronous commands.
func TestFuture(t *testing.T) {
	util.Debug("Future ...")

	rd := NewRedisDatabase(Configuration{})
	fut := rd.AsyncCommand("keys", "*")
	rs := fut.ResultSet()

	if !rs.IsOK() {
		t.Errorf("Future result set is not ok! RS: %v", rs)
	}

	if rs.ValueCount() == 0 {
		t.Errorf("Wrong number of values!")
	}
}

// Test complex commands.
func TestComplex(t *testing.T) {
	util.Debug("Complex ...")

	rd := NewRedisDatabase(Configuration{})
	rsA := rd.Command("info")

	if rsA.Value().StringMap()["arch_bits"] != "64" {
		t.Errorf("Invalid result, expected '64'!")
	}

	sliceIn := []string{"A", "B", "C", "D", "E"}

	rd.Command("set", "complex:slice", sliceIn)

	rsB := rd.Command("get", "complex:slice")
	sliceOut := rsB.Value().StringSlice()

	for i, s := range sliceOut {
		if sliceIn[i] != s {
			t.Errorf("Got '%v', expected '%v'!", s, sliceIn[i])
		}
	}

	mapIn := map[string]string{
		"A": "1",
		"B": "2",
		"C": "3",
		"D": "4",
		"E": "5",
	}

	rd.Command("set", "complex:map", mapIn)

	rsC := rd.Command("get", "complex:map")
	mapOut := rsC.Value().StringMap()

	for k, v := range mapOut {
		if mapIn[k] != v {
			t.Errorf("Got '%v', expected '%v'!", v, mapIn[k])
		}
	}
}

// Test multi-value commands.
func TestMulti(t *testing.T) {
	util.Debug("Multi ...")

	rd := NewRedisDatabase(Configuration{})

	rd.Command("sadd", "multi:set", "one")
	rd.Command("sadd", "multi:set", "two")
	rd.Command("sadd", "multi:set", "three")

	rsA := rd.Command("smembers", "multi:set")

	if rsA.ValueCount() != 3 {
		t.Errorf("Got '%v', expected '3'!", rsA.ValueCount())
	}

	for i := 0; i < 100; i++ {
		rd.Command("rpush", "multi:list", i)
	}
}

// Test mass data.
func TestMass(t *testing.T) {
	util.Debug("Mass ...")

	rd := NewRedisDatabase(Configuration{})

	for i := 0; i < 1000; i++ {
		k := fmt.Sprintf("mass:set:%v", i)

		rd.Command("set", k, i)
	}
}

// Test long run to allow database kill.
func TestLongRun(t *testing.T) {
	util.Debug("Long run for database kill ...")

	rd := NewRedisDatabase(Configuration{PoolSize: 5})

	for i := 1; i < 120; i++ {
		util.Debug("Long run iteration #%v ...", i)

		if !rd.Command("set", "long:run", i).IsOK() {
			t.Errorf("Long run not ok!")

			return
		}

		if time.Sleep(1e9) != nil {
			t.Errorf("Error during sleep!")

			return
		}
	}
}

// Test transactions.
func TestTransactions(t *testing.T) {
	util.Debug("Transactions ...")

	rd := NewRedisDatabase(Configuration{})
	rsA := rd.MultiCommand(func(mc *MultiCommand) {
		mc.Command("set", "tx:a:string", "Hello, World!")
		mc.Command("get", "tx:a:string")
	})

	if rsA.ResultSetAt(1).ValueAsString() != "Hello, World!" {
		t.Errorf("Got '%v', expected 'Hello, World!'!", rsA.ResultSetAt(1).ValueAsString())
	}

	rsB := rd.MultiCommand(func(mc *MultiCommand) {
		mc.Command("set", "tx:b:string", "Hello, World!")
		mc.Command("get", "tx:b:string")
		mc.Discard()
		mc.Command("set", "tx:c:string", "Hello, Redis!")
		mc.Command("get", "tx:c:string")
	})

	if rsB.ResultSetAt(1).ValueAsString() != "Hello, Redis!" {
		t.Errorf("Got '%v', expected 'Hello, Redis!'!", rsB.ResultSetAt(1).ValueAsString())
	}
}

// Test pop.
func TestPop(t *testing.T) {
	util.Debug("Pop ...")

	fooPush := func(rd *RedisDatabase) {
		time.Sleep(1e9)
		rd.Command("lpush", "pop:first", "foo")
	}

	// Set A: no database timeout.
	rdA := NewRedisDatabase(Configuration{})

	go fooPush(rdA)

	rsAA := rdA.Command("blpop", "pop:first", 5)

	if kv := rsAA.KeyValue(); kv.Value.String() != "foo" {
		t.Errorf("Got '%v', expected 'foo'!", kv.Value.String())
	}

	rsAB := rdA.Command("blpop", "pop:first", 1)

	if rsAB.Error().String() != "rdc: timeout" {
		t.Errorf("Got '%v', expected 'rdc: timeout'!", rsAB.Error().String())
	}

	// Set B: database with timeout.
	rdB := NewRedisDatabase(Configuration{})

	rsBA := rdB.Command("blpop", "pop:first", 5)

	if rsBA.Error() == nil {
		t.Errorf("Expected an error!")
	}
}

// Test subscribe.
func TestSubscribe(t *testing.T) {
	util.Debug("Subscribe ...")

	rd := NewRedisDatabase(Configuration{})
	sub, err := rd.Subscribe("subscribe:one", "subscribe:two")

	if err != nil {
		t.Errorf("Can't subscribe, error is '%v'!", err)

		return
	}

	go func() {
		for sv := range sub.SubscriptionValueChan {
			if sv == nil {
				util.Debug("Received nil!")
			} else {
				util.Debug("Published '%v' Channel '%v' Pattern '%v'", sv, sv.Channel, sv.ChannelPattern)
			}
		}

		util.Debug("Subscription stopped!")
	}()

	if rd.Publish("subscribe:one", "1 Alpha") != 1 {
		t.Errorf("First publishing has illegal receiver count!")
	}

	rd.Publish("subscribe:one", "1 Beta")
	rd.Publish("subscribe:one", "1 Gamma")
	rd.Publish("subscribe:two", "2 Alpha")
	rd.Publish("subscribe:two", "2 Beta")

	time.Sleep(1e8)

	if cnt := sub.Unsubscribe("subscribe:two"); cnt != 1 {
		t.Errorf("First unsubscribe has illegal channel (pattern) count '%v', expected '1'!", cnt)
	}

	if cnt := sub.Unsubscribe("subscribe:one"); cnt != 0 {
		t.Errorf("Second unsubscribe has illegal channel (pattern) count '%v', expected '0'!", cnt)
	}

	if cnt := rd.Publish("subscribe:two", "2 Gamma"); cnt != 0 {
		t.Errorf("Last publishing has illegal receiver count '%v', expected '0'!", cnt)
	}

	sub.Subscribe("subscribe:*")

	rd.Publish("subscribe:one", "Pattern 1")
	rd.Publish("subscribe:two", "Pattern 2")

	time.Sleep(1e8)

	sub.Stop()
}

// Test illegal databases.
func TestIllegalDatabases(t *testing.T) {
	util.Debug("Illegal database ...")

	if testing.Short() {
		return
	}

	rdA := NewRedisDatabase(Configuration{Database: 4711})
	rsA := rdA.Command("ping")

	if rsA.Error().String() != "rdc: invalid DB index" {
		t.Errorf("Expected 'rdc: invalid DB index', got '%v'!", rsA.Error().String())
	}

	util.Debug("Testing illegal address ...")

	rdB := NewRedisDatabase(Configuration{Address: "192.168.100.100:12345"})
	rsB := rdB.Command("ping")

	if rsB.IsOK() {
		t.Errorf("Expected an error!'")
	}
}

// Test measuring (pure output).
func TestMeasuring(t *testing.T) {
	monitoring.Monitor().MeasuringPointsPrintAll()
	time.Sleep(1e9)
}

// EOF
