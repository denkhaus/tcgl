// Tideland Common Go Library - Monitoring
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"os"
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Monitoring - Release 2012-01-29"

//--------------------
// CONSTANTS
//--------------------

const (
	etmTLine  = "+------------------------------------------+-----------+-----------+-----------+-----------+---------------+-----------+\n"
	etmHeader = "| Name                                     | Count     | Min Dur   | Max Dur   | Avg Dur   | Total Dur     | Op/Sec    |\n"
	etmFormat = "| %-40s | %9d | %9.3f | %9.3f | %9.3f | %13.3f | %9d |\n"
	etmFooter = "| All times in milliseconds.                                                                                           |\n"
	etmELine  = "+----------------------------------------------------------------------------------------------------------------------+\n"

	ssiTLine  = "+------------------------------------------+-----------+---------------+---------------+---------------+---------------+\n"
	ssiHeader = "| Name                                     | Count     | Act Value     | Min Value     | Max Value     | Avg Value     |\n"
	ssiFormat = "| %-40s | %9d | %13d | %13d | %13d | %13d |\n"

	dsrTLine  = "+------------------------------------------+---------------------------------------------------------------------------+\n"
	dsrHeader = "| Name                                     | Value                                                                     |\n"
	dsrFormat = "| %-40s | %-73s |\n"
)

const (
	cmdMeasuringPointsMap = iota
	cmdMeasuringPointsDo
	cmdStaySetVariablesMap
	cmdStaySetVariablesDo
	cmdDynamicStatusRetrieversMap
	cmdDynamicStatusRetrieversDo
)

//--------------------
// MONITORING
//--------------------

// Command encapsulated the data for any command.
type command struct {
	opCode   int
	args     interface{}
	respChan chan interface{}
}

// The system monitor type.
type systemMonitor struct {
	etmData                   map[string]*MeasuringPoint
	ssiData                   map[string]*StaySetVariable
	dsrData                   map[string]DynamicStatusRetriever
	measuringChan             chan *Measuring
	valueChan                 chan *value
	retrieverRegistrationChan chan *retrieverRegistration
	commandChan               chan *command
}

// monitor is the one global monitor instance.
var monitor *systemMonitor

// init creates the global monitor.
func init() {
	monitor = &systemMonitor{
		etmData:                   make(map[string]*MeasuringPoint),
		ssiData:                   make(map[string]*StaySetVariable),
		dsrData:                   make(map[string]DynamicStatusRetriever),
		measuringChan:             make(chan *Measuring, 1000),
		valueChan:                 make(chan *value, 1000),
		retrieverRegistrationChan: make(chan *retrieverRegistration, 10),
		commandChan:               make(chan *command),
	}
	go backend()
}

// BeginMeasuring starts a new measuring with a given id.
// All measurings with the same id will be aggregated.
func BeginMeasuring(id string) *Measuring {
	return &Measuring{id, time.Now(), time.Now()}
}

// Measure the execution of a function.
func Measure(id string, f func()) {
	m := BeginMeasuring(id)
	f()
	m.EndMeasuring()
}

// MeasuringPointsMap performs the function f for all measuring points
// and returns a slice with the return values of the function that are
// not nil.
func MeasuringPointsMap(f func(*MeasuringPoint) interface{}) []interface{} {
	cmd := &command{cmdMeasuringPointsMap, f, make(chan interface{})}
	monitor.commandChan <- cmd
	resp := <-cmd.respChan
	return resp.([]interface{})
}

// MeasuringPointsDo performs the function f for 
// all measuring points.
func MeasuringPointsDo(f func(*MeasuringPoint)) {
	cmd := &command{cmdMeasuringPointsDo, f, nil}
	monitor.commandChan <- cmd
}

// MeasuringPointsWrite prints the measuring points for which
// the passed function returns true to the passed writer.
func MeasuringPointsWrite(w io.Writer, ff func(*MeasuringPoint) bool) {
	pf := func(d time.Duration) float64 { return float64(d) / 1000000.0 }
	// Header.
	fmt.Fprint(w, etmTLine)
	fmt.Fprint(w, etmHeader)
	fmt.Fprint(w, etmTLine)
	// Body.
	lines := MeasuringPointsMap(func(mp *MeasuringPoint) interface{} {
		if ff(mp) {
			ops := 1e9 / mp.AvgDuration
			return fmt.Sprintf(etmFormat, mp.Id, mp.Count, pf(mp.MinDuration), pf(mp.MaxDuration),
				pf(mp.AvgDuration), pf(mp.TtlDuration), ops)
		}
		return nil
	})
	for _, line := range lines {
		fmt.Fprint(w, line)
	}
	// Footer.
	fmt.Fprint(w, etmTLine)
	fmt.Fprint(w, etmFooter)
	fmt.Fprint(w, etmELine)
}

// MeasuringPointsPrintAll prints all measuring points
// to STDOUT.
func MeasuringPointsPrintAll() {
	MeasuringPointsWrite(os.Stdout, func(mp *MeasuringPoint) bool { return true })
}

// SetValue sets a value of a stay-set variable.
func SetValue(id string, v int64) {
	monitor.valueChan <- &value{id, v}
}

// StaySetVariablesMap performs the function f for all variables
// and returns a slice with the return values of the function that are
// not nil.
func StaySetVariablesMap(f func(*StaySetVariable) interface{}) []interface{} {
	cmd := &command{cmdStaySetVariablesMap, f, make(chan interface{})}
	monitor.commandChan <- cmd
	resp := <-cmd.respChan
	return resp.([]interface{})
}

// StaySetVariablesDo performs the function f for all
// variables.
func StaySetVariablesDo(f func(*StaySetVariable)) {
	cmd := &command{cmdStaySetVariablesDo, f, nil}
	monitor.commandChan <- cmd
}

// StaySetVariablesWrite prints the stay-set variables for which
// the passed function returns true to the passed writer.
func StaySetVariablesWrite(w io.Writer, ff func(*StaySetVariable) bool) {
	// Header.
	fmt.Fprint(w, ssiTLine)
	fmt.Fprint(w, ssiHeader)
	fmt.Fprint(w, ssiTLine)
	// Body.
	lines := StaySetVariablesMap(func(ssv *StaySetVariable) interface{} {
		if ff(ssv) {
			return fmt.Sprintf(ssiFormat, ssv.Id, ssv.Count, ssv.ActValue, ssv.MinValue, ssv.MaxValue, ssv.AvgValue)
		}

		return nil
	})
	for _, line := range lines {
		fmt.Fprint(w, line)
	}
	// Footer.
	fmt.Fprint(w, ssiTLine)
}

// StaySetVariablesPrintAll prints all stay-set variables
// to STDOUT.
func StaySetVariablesPrintAll() {
	StaySetVariablesWrite(os.Stdout, func(ssv *StaySetVariable) bool { return true })
}

// Register registers a new dynamic status retriever function.
func Register(id string, rf DynamicStatusRetriever) {
	monitor.retrieverRegistrationChan <- &retrieverRegistration{id, rf}
}

// Unregister unregisters a dynamic status retriever function.
func Unregister(id string) {
	monitor.retrieverRegistrationChan <- &retrieverRegistration{id, nil}
}

// DynamicStatusValuesMap performs the function f for all status values
// and returns a slice with the return values of the function that are
// not nil.
func DynamicStatusValuesMap(f func(string, string) interface{}) []interface{} {
	cmd := &command{cmdDynamicStatusRetrieversMap, f, make(chan interface{})}
	monitor.commandChan <- cmd
	resp := <-cmd.respChan
	return resp.([]interface{})
}

// DynamicStatusValuesDo performs the function f for all
// status values.
func DynamicStatusValuesDo(f func(string, string)) {
	cmd := &command{cmdDynamicStatusRetrieversDo, f, nil}
	monitor.commandChan <- cmd
}

// DynamicStatusValuesWrite prints the status values for which
// the passed function returns true to the passed writer.
func DynamicStatusValuesWrite(w io.Writer, ff func(string, string) bool) {
	// Header.
	fmt.Fprint(w, dsrTLine)
	fmt.Fprint(w, dsrHeader)
	fmt.Fprint(w, dsrTLine)
	// Body.
	lines := DynamicStatusValuesMap(func(id, dsv string) interface{} {
		if ff(id, dsv) {
			return fmt.Sprintf(dsrFormat, id, dsv)
		}

		return nil
	})
	for _, line := range lines {
		fmt.Fprint(w, line)
	}
	// Footer.
	fmt.Fprint(w, dsrTLine)
}

// DynamicStatusValuesPrintAll prints all status values to STDOUT.
func DynamicStatusValuesPrintAll() {
	DynamicStatusValuesWrite(os.Stdout, func(id, dsv string) bool { return true })
}

// Backend of the system monitor.
func backend() {
	for {
		select {
		case measuring := <-monitor.measuringChan:
			// Received a new measuring.
			if mp, ok := monitor.etmData[measuring.id]; ok {
				// Measuring point found.
				mp.update(measuring)
			} else {
				// New measuring point.
				monitor.etmData[measuring.id] = newMeasuringPoint(measuring)
			}
		case value := <-monitor.valueChan:
			// Received a new value.
			if ssv, ok := monitor.ssiData[value.id]; ok {
				// Variable found.
				ssv.update(value)
			} else {
				// New stay-set variable.
				monitor.ssiData[value.id] = newStaySetVariable(value)
			}
		case registration := <-monitor.retrieverRegistrationChan:
			// Received a new retriever for registration.
			if registration.dsr != nil {
				// Register a new retriever.
				monitor.dsrData[registration.id] = registration.dsr
			} else {
				// Deregister a retriever.
				delete(monitor.dsrData, registration.id)
			}
		case cmd := <-monitor.commandChan:
			// Receivedd a command to process.
			processCommand(cmd)
		}
	}
}

// Process a command.
func processCommand(cmd *command) {
	switch cmd.opCode {
	case cmdMeasuringPointsMap:
		// Map the measuring points.
		var resp []interface{}
		f := cmd.args.(func(*MeasuringPoint) interface{})
		for _, mp := range monitor.etmData {
			v := f(mp)
			if v != nil {
				resp = append(resp, v)
			}
		}
		cmd.respChan <- resp
	case cmdMeasuringPointsDo:
		// Iterate over the measurings.
		f := cmd.args.(func(*MeasuringPoint))
		for _, mp := range monitor.etmData {
			f(mp)
		}
	case cmdStaySetVariablesMap:
		// Map the stay-set variables.
		var resp []interface{}
		f := cmd.args.(func(*StaySetVariable) interface{})
		for _, ssv := range monitor.ssiData {
			v := f(ssv)
			if v != nil {
				resp = append(resp, v)
			}
		}
		cmd.respChan <- resp
	case cmdStaySetVariablesDo:
		// Iterate over the stay-set variables.
		f := cmd.args.(func(*StaySetVariable))
		for _, ssv := range monitor.ssiData {
			f(ssv)
		}
	case cmdDynamicStatusRetrieversMap:
		// Map the return values of the dynamic status
		// retriever functions.
		var resp []interface{}
		f := cmd.args.(func(string, string) interface{})
		for id, dsr := range monitor.dsrData {
			dsv := dsr()
			v := f(id, dsv)
			if v != nil {
				resp = append(resp, v)
			}
		}
		cmd.respChan <- resp
	case cmdDynamicStatusRetrieversDo:
		// Iterate over the return values of the
		// dynamic status retriever functions.
		f := cmd.args.(func(string, string))
		for id, dsr := range monitor.dsrData {
			dsv := dsr()
			f(id, dsv)
		}
	}
}

//--------------------
// ADDITIONAL MEASURING TYPES
//--------------------

// Measuring contains one measuring.
type Measuring struct {
	id        string
	startTime time.Time
	endTime   time.Time
}

// EndMEasuring ends a measuring and passes it to the 
// measuring server in the background.
func (m *Measuring) EndMeasuring() time.Duration {
	m.endTime = time.Now()
	monitor.measuringChan <- m
	return m.endTime.Sub(m.startTime)
}

// MeasuringPoint contains the cumulated measuring
// data of one measuring point.
type MeasuringPoint struct {
	Id          string
	Count       int64
	MinDuration time.Duration
	MaxDuration time.Duration
	TtlDuration time.Duration
	AvgDuration time.Duration
}

// Create a new measuring point out of a measuring.
func newMeasuringPoint(m *Measuring) *MeasuringPoint {
	duration := m.endTime.Sub(m.startTime)
	mp := &MeasuringPoint{
		Id:          m.id,
		Count:       1,
		MinDuration: duration,
		MaxDuration: duration,
		TtlDuration: duration,
		AvgDuration: duration,
	}
	return mp
}

// Update a measuring point with a measuring.
func (mp *MeasuringPoint) update(m *Measuring) {
	duration := m.endTime.Sub(m.startTime)
	mp.Count++
	if mp.MinDuration > duration {
		mp.MinDuration = duration
	}
	if mp.MaxDuration < duration {
		mp.MaxDuration = duration
	}
	mp.TtlDuration += duration
	mp.AvgDuration = time.Duration(mp.TtlDuration.Nanoseconds() / mp.Count)
}

// value stores a stay-set variable with a given id.
type value struct {
	id    string
	value int64
}

// StaySetVariable contains the cumulated values
// for one stay-set variable.
type StaySetVariable struct {
	Id       string
	Count    int64
	ActValue int64
	MinValue int64
	MaxValue int64
	AvgValue int64
	total    int64
}

// Create a new stay-set variable out of a value.
func newStaySetVariable(v *value) *StaySetVariable {
	ssv := &StaySetVariable{
		Id:       v.id,
		Count:    1,
		ActValue: v.value,
		MinValue: v.value,
		MaxValue: v.value,
		AvgValue: v.value,
	}
	return ssv
}

// Update a stay-set variable with a value.
func (ssv *StaySetVariable) update(v *value) {
	ssv.Count++
	ssv.ActValue = v.value
	ssv.total += v.value
	if ssv.MinValue > ssv.ActValue {
		ssv.MinValue = ssv.ActValue
	}
	if ssv.MaxValue < ssv.ActValue {
		ssv.MaxValue = ssv.ActValue
	}
	ssv.AvgValue = ssv.total / ssv.Count
}

// DynamicStatusRetriever is called by the server and
// returns a current status as string.
type DynamicStatusRetriever func() string

// New registration of a retriever function.
type retrieverRegistration struct {
	id  string
	dsr DynamicStatusRetriever
}

// EOF