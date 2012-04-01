// Tideland Common Go Library - Redis
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
	"code.google.com/p/tcgl/identifier"
	"code.google.com/p/tcgl/monitoring"
	"fmt"
	"sync"
	"time"
)

//--------------------
// CONFIGURATION
//--------------------

// Configuration of a database client.
type Configuration struct {
	Address     string
	Timeout     time.Duration
	Database    int
	Auth        string
	PoolSize    int
	LogCommands bool
}

// String returns the configured address and
// database as string.
func (c *Configuration) String() string {
	return fmt.Sprintf("%s/%d", c.Address, c.Database)
}

//--------------------
// REDIS DATABASE
//--------------------

// RedisDatabase manages the access to one database.
type RedisDatabase struct {
	mutex         sync.Mutex
	configuration *Configuration
	pool          chan *unifiedRequestProtocol
	poolUsage     int
	dbClosed      bool
}

// NewRedisDatabase create a new accessor.
func NewRedisDatabase(c Configuration) *RedisDatabase {
	checkConfiguration(&c)
	// Create the database client instance.
	rd := &RedisDatabase{
		configuration: &c,
		pool:          make(chan *unifiedRequestProtocol, c.PoolSize),
	}
	// Init pool with nils.
	for i := 0; i < c.PoolSize; i++ {
		rd.pool <- nil
	}
	return rd
}

// Command performs a Redis command.
func (rd *RedisDatabase) Command(cmd string, args ...interface{}) *ResultSet {
	rs := newResultSet(cmd)
	if rd.dbClosed {
		rs.err = &DatabaseClosedError{rd}
		return rs
	}

	urp, err := rd.pullURP()
	defer rd.pushURP(urp)

	if err != nil {
		rs.err = err
		return rs
	}
	urp.command(rs, false, cmd, args...)
	return rs
}

// AsyncCommand performs a Redis command asynchronously.
func (rd *RedisDatabase) AsyncCommand(cmd string, args ...interface{}) *Future {
	fut := newFuture()
	go func() {
		fut.setResultSet(rd.Command(cmd, args...))
	}()
	return fut
}

// MultiCommand executes a function for the performing
// of multiple commands in one call.
func (rd *RedisDatabase) MultiCommand(f func(*MultiCommand)) *ResultSet {
	// Create result set.
	rs := newResultSet("multi")
	rs.resultSets = []*ResultSet{}

	urp, err := rd.pullURP()
	defer rd.pushURP(urp)

	if err != nil {
		rs.err = err
		return rs
	}

	mc := newMultiCommand(rs, urp)
	mc.process(f)
	return rs
}

// AsyncMultiCommand executes a function for the performing
// of multiple commands in one call asynchronously.
func (rd *RedisDatabase) AsyncMultiCommand(f func(*MultiCommand)) *Future {
	fut := newFuture()
	go func() {
		fut.setResultSet(rd.MultiCommand(f))
	}()
	return fut
}

// Subscribe to one or more channels.
func (rd *RedisDatabase) Subscribe(channel ...string) (*Subscription, error) {
	// URP handling.
	urp, err := newUnifiedRequestProtocol(rd)
	if err != nil {
		return nil, err
	}
	// Now return new subscription.
	return newSubscription(urp, channel...), nil
}

// Publish a message to a channel.
func (rd *RedisDatabase) Publish(channel string, message interface{}) (int, error) {
	rs := rd.Command("publish", channel, message)
	if !rs.IsOK() {
		return 0, rs.Error()
	}
	v, err := rs.Value().Int64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// pullURP retrieves a unified request protocol managing the
// communication with Redis out of the pool.
func (rd *RedisDatabase) pullURP() (*unifiedRequestProtocol, error) {
	rd.mutex.Lock()
	defer rd.mutex.Unlock()

	urp := <-rd.pool
	if urp == nil {
		// Lazy creation of a new URP.
		var err error
		urp, err = newUnifiedRequestProtocol(rd)
		if err != nil {
			return nil, err
		}
	}
	rd.poolUsage++
	monitoring.SetVariable(identifier.Identifier("redis", "pool", "usage"), int64(rd.poolUsage))
	return urp, nil
}

// pushURP returns a unified request protocol back to the pool.
func (rd *RedisDatabase) pushURP(urp *unifiedRequestProtocol) {
	rd.mutex.Lock()
	defer rd.mutex.Unlock()

	rd.pool <- urp
	if urp != nil {
		rd.poolUsage--
	}
	monitoring.SetVariable(identifier.Identifier("redis", "pool", "usage"), int64(rd.poolUsage))
}

//--------------------
// MULTI COMMAND
//--------------------

// MultiCommand enables the user to perform multiple commands
// in one call.
type MultiCommand struct {
	urp       *unifiedRequestProtocol
	rs        *ResultSet
	discarded bool
}

// newMultiCommand creates a new multi command helper.
func newMultiCommand(rs *ResultSet, urp *unifiedRequestProtocol) *MultiCommand {
	return &MultiCommand{
		urp: urp,
		rs:  rs,
	}
}

// process executes the multi command function.
func (mc *MultiCommand) process(f func(*MultiCommand)) {
	// Send the multi command.
	mc.urp.command(mc.rs, false, "multi")
	if mc.rs.IsOK() {
		// Execute multi command function.
		f(mc)
		mc.urp.command(mc.rs, true, "exec")
	}
}

// Command performs a command inside the transaction. It will
// be queued.
func (mc *MultiCommand) Command(cmd string, args ...interface{}) {
	rs := newResultSet(cmd)
	mc.rs.resultSets = append(mc.rs.resultSets, rs)
	mc.urp.command(rs, false, cmd, args...)
}

// Discard throws all so far queued commands away.
func (mc *MultiCommand) Discard() {
	// Send the discard command and empty result sets.
	mc.urp.command(mc.rs, false, "discard")
	mc.rs.resultSets = []*ResultSet{}
	// Now send the new multi command.
	mc.urp.command(mc.rs, false, "multi")
}

//--------------------
// HELPERS
//--------------------

// checkConfiguration ensures that unset configuration
// parameters get default values.
func checkConfiguration(c *Configuration) {
	if c.Address == "" {
		// Default is localhost and default port.
		c.Address = "127.0.0.1:6379"
	}
	if c.Timeout <= 0 {
		// Timeout for connection dialing is 5 seconds.
		c.Timeout = 5 * time.Second
	}
	if c.Database < 0 {
		// Shouldn't happen.
		c.Database = 0
	}
	if c.PoolSize <= 0 {
		// Default is 10.
		c.PoolSize = 10
	}
}

// EOF
