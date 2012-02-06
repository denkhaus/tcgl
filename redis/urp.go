// Tideland Common Go Library - Redis - Unified Request Protocol
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
	"bufio"
	"code.google.com/p/tcgl/identifier"
	"code.google.com/p/tcgl/monitoring"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

//--------------------
// MISC
//--------------------

// Envelope type for commands.
type envCommand struct {
	rs       *ResultSet
	multi    bool
	command  string
	args     []interface{}
	doneChan chan bool
}

// Envelope type for subscriptions.
type envSubscription struct {
	in        bool
	channels  []string
	countChan chan int
}

// Envelope type for read data.
type envData struct {
	length int
	data   []byte
	err    error
}

// Helper for debugging.
func (ed *envData) String() string {
	return fmt.Sprintf("ED(%v / %s / %v)", ed.length, ed.data, ed.err)
}

// Envelope type for published data.
type envPublishedData struct {
	data [][]byte
	err  error
}

//--------------------
// UNIFIED REQUEST PROTOCOL
//--------------------

// Redis unified request protocol type.
type unifiedRequestProtocol struct {
	database          *RedisDatabase
	conn              net.Conn
	writer            *bufio.Writer
	reader            *bufio.Reader
	commandChan       chan *envCommand
	subscriptionChan  chan *envSubscription
	dataChan          chan *envData
	publishedDataChan chan *envPublishedData
	stopChan          chan bool
}

// Create a new protocol.
func newUnifiedRequestProtocol(rd *RedisDatabase) (*unifiedRequestProtocol, error) {
	// Establish the connection.
	conn, err := net.DialTimeout("tcp", rd.configuration.Address, rd.configuration.Timeout*time.Millisecond)
	if err != nil {
		return nil, err
	}
	// Create the URP.
	urp := &unifiedRequestProtocol{
		database:          rd,
		conn:              conn,
		writer:            bufio.NewWriter(conn),
		reader:            bufio.NewReader(conn),
		commandChan:       make(chan *envCommand),
		subscriptionChan:  make(chan *envSubscription),
		dataChan:          make(chan *envData, 20),
		publishedDataChan: make(chan *envPublishedData, 5),
		stopChan:          make(chan bool),
	}
	// Start goroutines.
	go urp.receiver()
	go urp.backend()
	// Select database.
	rs := newResultSet("select")
	urp.command(rs, false, "select", rd.configuration.Database)
	if !rs.IsOK() {
		// Connection or database is not ok, so reset.
		urp.stop()
		return nil, rs.Error()
	}
	// Authenticate if needed.
	if rd.configuration.Auth != "" {
		rs = newResultSet("auth")
		urp.command(rs, false, "auth", rd.configuration.Auth)
		if !rs.IsOK() {
			// Authentication is not ok, so reset.
			urp.stop()
			return nil, rs.Error()
		}
	}
	return urp, nil
}

// Execute a command.
func (urp *unifiedRequestProtocol) command(rs *ResultSet, multi bool, command string, args ...interface{}) {
	m := monitoring.BeginMeasuring(identifier.Identifier("redis", "command", command))
	doneChan := make(chan bool)
	urp.commandChan <- &envCommand{rs, multi, command, args, doneChan}
	<-doneChan
	m.EndMeasuring()
}

// Send a subscription request.
func (urp *unifiedRequestProtocol) subscribe(channels ...string) int {
	countChan := make(chan int)
	urp.subscriptionChan <- &envSubscription{true, channels, countChan}
	return <-countChan
}

// Send an unsubscription request.
func (urp *unifiedRequestProtocol) unsubscribe(channels ...string) int {
	countChan := make(chan int)
	urp.subscriptionChan <- &envSubscription{false, channels, countChan}
	return <-countChan
}

// Stop the protocol.
func (urp *unifiedRequestProtocol) stop() {
	urp.stopChan <- true
}

// Goroutine for receiving data from the TCP connection.
func (urp *unifiedRequestProtocol) receiver() {
	var ed *envData
	for {
		b, err := urp.reader.ReadBytes('\n')
		if err != nil {
			urp.dataChan <- &envData{0, nil, err}
			return
		}
		// Analyze first bytes.
		switch b[0] {
		case '+':
			// Status reply.
			r := b[1 : len(b)-2]
			ed = &envData{len(r), r, nil}
		case '-':
			// Error reply.
			ed = &envData{0, nil, errors.New("redis: " + string(b[5:len(b)-2]))}
		case ':':
			// Integer reply.
			r := b[1 : len(b)-2]
			ed = &envData{len(r), r, nil}
		case '$':
			// Bulk reply, or key not found.
			i, _ := strconv.Atoi(string(b[1 : len(b)-2]))
			if i == -1 {
				// Key not found.
				ed = &envData{0, nil, errors.New("redis: key not found")}
			} else {
				// Reading the data.
				ir := i + 2
				br := make([]byte, ir)
				r := 0
				for r < ir {
					n, err := urp.reader.Read(br[r:])
					if err != nil {
						urp.dataChan <- &envData{0, nil, err}
						return
					}
					r += n
				}
				ed = &envData{i, br[0:i], nil}
			}
		case '*':
			// Multi-bulk reply. Just return the count
			// of the replies. The caller has to do the
			// individual calls.
			i, _ := strconv.Atoi(string(b[1 : len(b)-2]))
			ed = &envData{i, nil, nil}
		default:
			// Oops!
			ed = &envData{0, nil, errors.New("redis: invalid received data type")}
		}
		// Send result.
		urp.dataChan <- ed
	}
}

// Goroutine as backend for the protocol.
func (urp *unifiedRequestProtocol) backend() {
	// Prepare cleanup.
	defer func() {
		urp.conn.Close()

		urp.conn = nil
	}()
	// Receive commands and data.
	for {
		select {
		case ec := <-urp.commandChan:
			// Received a command.
			urp.handleCommand(ec)
		case es := <-urp.subscriptionChan:
			// Received a subscription.
			urp.handleSubscription(es)
		case ed := <-urp.dataChan:
			// Received data w/o command, so published data
			// after a subscription.
			urp.handlePublishing(ed)
		case <-urp.stopChan:
			// Stop processing.
			return
		}
	}
}

// Handle a sent command.
func (urp *unifiedRequestProtocol) handleCommand(ec *envCommand) {
	if err := urp.writeRequest(ec.command, ec.args); err == nil {
		// Receive and return reply.
		urp.receiveReply(ec.rs, ec.multi)
	} else {
		// Return error.
		ec.rs.err = err
	}
	urp.logCommand(ec)
	ec.doneChan <- true
}

// logCommand logs a command and its execution status.
func (urp *unifiedRequestProtocol) logCommand(ec *envCommand) {
	var log string
	if ec.multi {
		log = "multi "	
	}
	log += "command " + ec.command
	for _, arg := range ec.args {
		log = fmt.Sprintf("%s %v", log, arg)
	}
	if ec.rs.IsOK() {
		urp.database.logger.Infof("%s OK", log)
	} else {
		urp.database.logger.Warningf("%s ERROR %v", log, ec.rs.err)		
	}
}

// Handle a subscription.
func (urp *unifiedRequestProtocol) handleSubscription(es *envSubscription) {
	// Prepare command.
	var command string
	if es.in {
		command = "subscribe"
	} else {
		command = "unsubscribe"
	}
	cis, pattern := urp.prepareChannels(es.channels)
	if pattern {
		command = "p" + command
	}
	// Send the subscription request.
	rs := newResultSet(command)
	if err := urp.writeRequest(command, cis); err != nil {
		es.countChan <- 0
		return
	}
	// Receive the replies.
	channelLen := len(es.channels)
	rs.resultSets = make([]*ResultSet, channelLen)
	rs.err = nil
	for i := 0; i < channelLen; i++ {
		rs.resultSets[i] = newResultSet(command)
		urp.receiveReply(rs.resultSets[i], false)
	}
	// Get the number of subscribed channels.
	lastResultSet := rs.ResultSetAt(channelLen - 1)
	lastResultValue := lastResultSet.ValueAt(lastResultSet.ValueCount() - 1)
	es.countChan <- int(lastResultValue.Int64())
}

// Handle published data.
func (urp *unifiedRequestProtocol) handlePublishing(ed *envData) {
	// Continue according to the initial data.
	switch {
	case ed.err != nil:
		// Error.
		urp.publishedDataChan <- &envPublishedData{nil, ed.err}
	case ed.length > 0:
		// Multiple results as part of the one reply.
		values := make([][]byte, ed.length)
		for i := 0; i < ed.length; i++ {
			ed := <-urp.dataChan
			if ed.err != nil {
				urp.publishedDataChan <- &envPublishedData{nil, ed.err}
			}
			values[i] = ed.data
		}
		urp.publishedDataChan <- &envPublishedData{values, nil}
	case ed.length == -1:
		// Timeout.
		urp.publishedDataChan <- &envPublishedData{nil, errors.New("redis: timeout")}
	default:
		// Invalid reply.
		urp.publishedDataChan <- &envPublishedData{nil, errors.New("redis: invalid reply")}
	}
}

// Write a request.
func (urp *unifiedRequestProtocol) writeRequest(cmd string, args []interface{}) error {
	// Calculate number of data.
	dataNum := 1
	for _, arg := range args {
		switch typedArg := arg.(type) {
		case Hash:
			dataNum += len(typedArg) * 2
		case Hashable:
			dataNum += len(typedArg.GetHash()) * 2
		default:
			dataNum++
		}
	}
	// Write number of following data.
	if err := urp.writeDataNumber(dataNum); err != nil {
		return err
	}
	// Write command.
	if err := urp.writeData([]byte(cmd)); err != nil {
		return err
	}
	// Write arguments.
	for _, arg := range args {
		if err := urp.writeArgument(arg); err != nil {
			return err
		}
	}
	return nil
}

// Write the number of arguments.
func (urp *unifiedRequestProtocol) writeDataNumber(dataLen int) error {
	urp.writer.Write([]byte(fmt.Sprintf("*%d\r\n", dataLen)))
	return urp.writer.Flush()
}

// Write data.
func (urp *unifiedRequestProtocol) writeData(data []byte) error {
	// Write the len of the data.
	b := []byte(fmt.Sprintf("$%d\r\n", len(data)))
	if _, err := urp.writer.Write(b); err != nil {
		return err
	}
	// Write the data.
	if _, err := urp.writer.Write(data); err != nil {
		return err
	}
	urp.writer.Write([]byte{'\r', '\n'})
	return urp.writer.Flush()
}

// Write a request argument.
func (urp *unifiedRequestProtocol) writeArgument(arg interface{}) error {
	// Little helper for converting and writing.
	convertAndWrite := func(a interface{}) error {
		// Convert data.
		data := valueToBytes(a)
		// Now write data.
		if err := urp.writeData(data); err != nil {
			return err
		}
		return nil
	}
	// Another helper for writing a hash.
	writeHash := func(h Hash) error {
		for k, v := range h {
			if err := convertAndWrite(k); err != nil {
				return err
			}
			if err := convertAndWrite(v); err != nil {
				return err
			}
		}
		return nil
	}
	// Switch types.
	switch typedArg := arg.(type) {
	case Hash:
		if err := writeHash(typedArg); err != nil {
			return err
		}
	case Hashable:
		if err := writeHash(typedArg.GetHash()); err != nil {
			return err
		}
	default:
		if err := convertAndWrite(typedArg); err != nil {
			return err
		}
	}
	return nil
}

// Receive a reply.
func (urp *unifiedRequestProtocol) receiveReply(rs *ResultSet, multi bool) {
	// Read initial data.
	ed := <-urp.dataChan
	// Continue according to the initial data.
	switch {
	case ed.err != nil:
		// Error.
		rs.err = ed.err
	case ed.data != nil:
		// Single result.
		rs.values = []Value{Value(ed.data)}
		rs.err = nil
	case ed.length > 0:
		// Multiple result sets or results.
		rs.err = nil
		if multi {
			for i := 0; i < ed.length; i++ {
				urp.receiveReply(rs.resultSets[i], false)
			}
		} else {
			rs.values = make([]Value, ed.length)
			for i := 0; i < ed.length; i++ {
				ied := <-urp.dataChan
				if ied.err != nil {
					rs.values = nil
					rs.err = ied.err
					return
				}
				rs.values[i] = Value(ied.data)
			}
		}
	case ed.length == -1:
		// Timeout.
		rs.err = errors.New("redis: timeout")
	default:
		// Invalid reply.
		rs.err = errors.New("redis: invalid reply")
	}
}

// Prepare the channels.
func (urp *unifiedRequestProtocol) prepareChannels(channels []string) ([]interface{}, bool) {
	pattern := false
	cis := make([]interface{}, len(channels))
	for idx, channel := range channels {
		cis[idx] = channel
		if strings.IndexAny(channel, "*?[") != -1 {
			pattern = true
		}
	}
	return cis, pattern
}

// EOF
