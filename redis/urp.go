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
	"cgl.tideland.biz/applog"
	"cgl.tideland.biz/identifier"
	"cgl.tideland.biz/monitoring"
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

// envCommand is the envelope for almost all commands.
type envCommand struct {
	rs       *ResultSet
	multi    bool
	command  string
	args     []interface{}
	doneChan chan bool
}

// envSubscription is the envelope for subscriptions.
type envSubscription struct {
	in        bool
	channels  []string
	countChan chan int
}

// envData is the envelope for data read from the database.
type envData struct {
	length int
	data   []byte
	err    error
}

// String returns the data in a more human readable way.
func (ed *envData) String() string {
	return fmt.Sprintf("DATA(%v / %s / %v)", ed.length, ed.data, ed.err)
}

// envPublishedData is the envelope for published data.
type envPublishedData struct {
	data [][]byte
	err  error
}

//--------------------
// UNIFIED REQUEST PROTOCOL
//--------------------

// unifiedRequestProtocol implements the Redis unified request protocol URP.
type unifiedRequestProtocol struct {
	database          *Database
	conn              net.Conn
	writer            *bufio.Writer
	reader            *bufio.Reader
	err               error
	commandChan       chan *envCommand
	subscriptionChan  chan *envSubscription
	dataChan          chan *envData
	publishedDataChan chan *envPublishedData
	stopChan          chan bool
}

// newUnifiedRequestProtocol creates a new protocol.
func newUnifiedRequestProtocol(db *Database) (*unifiedRequestProtocol, error) {
	// Establish the connection.
	conn, err := net.DialTimeout("tcp", db.configuration.Address, db.configuration.Timeout)
	if err != nil {
		return nil, &ConnectionError{err}
	}
	// Create the URP.
	urp := &unifiedRequestProtocol{
		database:          db,
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
	urp.command(rs, false, "select", db.configuration.Database)
	if !rs.IsOK() {
		// Connection or database is not ok, so reset.
		urp.stop()
		return nil, rs.Error()
	}
	// Authenticate if needed.
	if db.configuration.Auth != "" {
		rs = newResultSet("auth")
		urp.command(rs, false, "auth", db.configuration.Auth)
		if !rs.IsOK() {
			// Authentication is not ok, so reset.
			urp.stop()
			return nil, rs.Error()
		}
	}
	return urp, nil
}

// command performs a Redis command.
func (urp *unifiedRequestProtocol) command(rs *ResultSet, multi bool, command string, args ...interface{}) {
	m := monitoring.BeginMeasuring(identifier.Identifier("redis", "command", command))
	doneChan := make(chan bool)
	urp.commandChan <- &envCommand{rs, multi, command, args, doneChan}
	<-doneChan
	m.EndMeasuring()
}

// subscribe subscribes to one or more channels.
func (urp *unifiedRequestProtocol) subscribe(channels ...string) int {
	countChan := make(chan int)
	urp.subscriptionChan <- &envSubscription{true, channels, countChan}
	return <-countChan
}

// unsubscribe unsubscribes from one or more channels.
func (urp *unifiedRequestProtocol) unsubscribe(channels ...string) int {
	countChan := make(chan int)
	urp.subscriptionChan <- &envSubscription{false, channels, countChan}
	return <-countChan
}

// stop tells the protocol to end its work.
func (urp *unifiedRequestProtocol) stop() {
	urp.stopChan <- true
}

// receiver is the goroutine for the receiving of the results in the background.
func (urp *unifiedRequestProtocol) receiver() {
	var ed *envData
	for {
		b, err := urp.reader.ReadBytes('\n')
		if err != nil {
			urp.dataChan <- &envData{0, nil, &ConnectionError{err}}
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
						urp.dataChan <- &envData{0, nil, &ConnectionError{err}}
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

// backend is the backend goroutine for the protocol.
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

// handleCommand executes a command and returns the reply.
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
	// Format the command for the log entry.
	formatCommand := func() string {
		var log string
		if ec.multi {
			log = "multi "
		}
		log += "command " + ec.command
		for _, arg := range ec.args {
			log = fmt.Sprintf("%s %v", log, arg)
		}
		return log
	}
	// Positive commands only if wanted, errors always.
	if ec.rs.IsOK() {
		if urp.database.configuration.LogCommands {
			applog.Infof("%s OK", formatCommand())
		}
	} else {
		applog.Errorf("%s ERROR %v", formatCommand(), ec.rs.err)
	}
}

// handleSubscription exucutes subscribe and unsubscribe commands.
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
	v, _ := lastResultValue.Int64()
	es.countChan <- int(v)
}

// handlePublishing handles the publishing of data to a channel.
func (urp *unifiedRequestProtocol) handlePublishing(ed *envData) {
	start := time.Now()
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
	case ed.length == 0:
		// No result.
		urp.publishedDataChan <- &envPublishedData{[][]byte{}, nil}
	case ed.length == -1:
		// Timeout.
		urp.publishedDataChan <- &envPublishedData{nil, &TimeoutError{time.Now().Sub(start)}}
	default:
		// Invalid reply.
		urp.publishedDataChan <- &envPublishedData{nil, &InvalidReplyError{ed.length, ed.data, ed.err}}
	}
}

// writeRequest send the request to the server.
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
	// Write the number, the command and its arguments.
	if err := urp.writeDataNumber(dataNum); err != nil {
		return err
	}
	if err := urp.writeData([]byte(cmd)); err != nil {
		return err
	}
	for _, arg := range args {
		if err := urp.writeArgument(arg); err != nil {
			return err
		}
	}
	return nil
}

// writeDataNumber sends the number of data elements to the server.
func (urp *unifiedRequestProtocol) writeDataNumber(dataLen int) (err error) {
	if err = urp.write([]byte(fmt.Sprintf("*%d\r\n", dataLen))); err != nil {
		return
	}
	if err = urp.flush(); err != nil {
		return
	}
	return nil
}

// writeData sends a data element to the server.
func (urp *unifiedRequestProtocol) writeData(data []byte) (err error) {
	// Write the len of the data.
	if err = urp.write([]byte(fmt.Sprintf("$%d\r\n", len(data)))); err != nil {
		return
	}
	// Write the data.
	if err = urp.write(data); err != nil {
		return
	}
	if err = urp.write([]byte{'\r', '\n'}); err != nil {
		return
	}
	if err = urp.flush(); err != nil {
		return
	}
	return nil
}

// writeArgument sends an argument to the server.
func (urp *unifiedRequestProtocol) writeArgument(arg interface{}) (err error) {
	// Little helper for converting and writing.
	convertAndWrite := func(a interface{}) (ierr error) {
		data := valueToBytes(a)
		if ierr := urp.writeData(data); ierr != nil {
			return ierr
		}
		return nil
	}
	// Another helper for writing a hash.
	writeHash := func(h Hash) error {
		for k, v := range h {
			if ierr := convertAndWrite(k); ierr != nil {
				return err
			}
			if ierr := convertAndWrite(v); ierr != nil {
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

// write is an error handling wrapper for the writers write method.
func (urp *unifiedRequestProtocol) write(raw []byte) error {
	if _, err := urp.writer.Write(raw); err != nil {
		urp.err = &ConnectionError{err}
		return urp.err
	}
	return nil
}

// flush is an error handling wrapper for the writers flush method.
func (urp *unifiedRequestProtocol) flush() error {
	if err := urp.writer.Flush(); err != nil {
		urp.err = &ConnectionError{err}
		return urp.err
	}
	return nil
}

// receiveReply gets the reply from the server.
func (urp *unifiedRequestProtocol) receiveReply(rs *ResultSet, multi bool) {
	start := time.Now()
	ed := <-urp.dataChan
	switch {
	case ed.err != nil:
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
	case ed.length == 0:
		// No result.
		rs.values = []Value{}
		rs.err = nil
	case ed.length == -1:
		// Timeout.
		rs.err = &TimeoutError{time.Now().Sub(start)}
	default:
		// Invalid reply.
		rs.err = &InvalidReplyError{ed.length, ed.data, ed.err}
	}
	urp.err = rs.err
}

// prepareChannels converts the channels from strings to interfaces which is
// needed for proper writing. It also checks if one of the channels contains a
// pattern.
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
