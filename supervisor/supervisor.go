// Tideland Common Go Library - Supervisor
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package supervisor

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/applog"
	"fmt"
	"time"
)

//--------------------
// MESSAGE
//--------------------

// message is for sending informations to superisors and children.
type message struct {
	code     int
	id       string
	sup      supervisable
	reason   interface{}
	response chan *message
}

func (m *message) err() error {
	if r, ok := m.reason.(error); ok {
		return r
	}
	return fmt.Errorf("reason: %v", m.reason)
}

const (
	msgStart = iota
	msgTerminate
	msgStop
	msgError
)

func newStartMsg(id string, sup supervisable) *message {
	return &message{
		code:     msgStart,
		id:       id,
		sup:      sup,
		response: make(chan *message),
	}
}

func newTerminateMsg(id string, reason interface{}) *message {
	return &message{
		code:     msgTerminate,
		id:       id,
		reason:   reason,
		response: make(chan *message),
	}
}

func newStopMsg(id string, reason interface{}) *message {
	return &message{
		code:     msgStop,
		id:       id,
		reason:   reason,
		response: make(chan *message),
	}
}

func newErrorMsg(id string, reason interface{}) *message {
	return &message{
		code:   msgError,
		id:     id,
		reason: reason,
	}
}

//--------------------
// HANDLE
//--------------------

// Handle contains the needed informations for a communication between
// supervisor and child.
type Handle struct {
	id         string
	supervisor *Supervisor
	terminate  chan bool
}

// Id return the id of the child.
func (h *Handle) Id() string {
	return h.id
}

// Terminate return a channel signaling that the goroutine should terminate.
func (h *Handle) Terminate() <-chan bool {
	return h.terminate
}

// IsTerminated returns true if a termination is signalled. It is intended
// for goroutines not using a select loop internally.
func (h *Handle) IsTerminated() bool {
	select {
	case <-h.terminate:
		return true
	default:
	}
	return false
}

// String returns the hierarchical id of the child.
func (h *Handle) String() string {
	return fmt.Sprintf("%s/%s", h.supervisor, h.id)
}

//--------------------
// SUPERVISABLE
//--------------------

// supervisable is the interface for all supervisable types
type supervisable interface {
	handle() *Handle
	setHandle(h *Handle)
	start()
	stop()
}

// status represents the status of a supervisable.
type status string

const (
	stReady   = "ready"
	stRunning = "running"
	stError   = "error"
)

//--------------------
// SUPERVISABLE FUNCTION
//--------------------

// SupervisedFunc is the signature of the goroutine 
// function that's supervised.
type SupervisedFunc func(h *Handle) error

type supervisableFunc struct {
	h      *Handle
	sfunc  SupervisedFunc
	status status
}

// handle returns the child handle.
func (sf *supervisableFunc) handle() *Handle {
	return sf.h
}

// setHandle supplies the child with the needed informations.
func (sf *supervisableFunc) setHandle(h *Handle) {
	sf.h = h
}

// wrapper is responsible for error and panic handling of the goroutine.
func (sf *supervisableFunc) wrapper() {
	var err error
	defer func() {
		var msg *message
		r := recover()
		switch {
		case r != nil:
			sf.status = stError
			msg = newErrorMsg(sf.h.id, r)
		case err != nil:
			sf.status = stError
			msg = newErrorMsg(sf.h.id, err)
		default:
			sf.status = stReady
		}
		if msg != nil {
			sf.h.supervisor.messages <- msg
		}
	}()
	err = sf.sfunc(sf.h)
}

// start runs the goroutine with the needed wrapping for error 
// and panic handling.
func (sf *supervisableFunc) start() {
	if sf.status == stReady {
		sf.status = stRunning
		go sf.wrapper()
	}
}

// stop signals the termination to the goroutine.
func (sf *supervisableFunc) stop() {
	if sf.status == stRunning {
		sf.h.terminate <- true
	}
	sf.status = stReady
}

//--------------------
// RESTART FREQUENCY
//--------------------

// restartFrequency is used to check the maximum number of
// restarts in a given period.
type restartFrequency struct {
	intensity int
	period    int64
	restarts  []int64
}

// newRestartFrequency returns an initialized restart frequency.
func newRestartFrequency(intensity int, period time.Duration) *restartFrequency {
	return &restartFrequency{
		intensity: intensity,
		period:    period.Nanoseconds(),
		restarts:  make([]int64, 0),
	}
}

// check stores restarts and checks their frequency. If the limit
// is exceeded an error is returned.
func (f *restartFrequency) check() error {
	// Check if enough values.
	if len(f.restarts) < f.intensity {
		f.restarts = append(f.restarts, time.Now().UnixNano())
		return nil
	}
	// Already intensity values, add new one and check frequency.
	copy(f.restarts, f.restarts[1:])
	f.restarts[f.intensity-1] = time.Now().UnixNano()
	p := f.restarts[f.intensity-1] - f.restarts[0]
	if p <= f.period {
		// Reset and return error.
		f.restarts = make([]int64, 0)
		return &TooMuchRestartsError{
			Restarts: f.intensity,
			Period:   time.Duration(p),
		}
	}
	return nil
}

//--------------------
// SUPERVISOR
//--------------------

// Strategy defines the restart strategy.
type Strategy int

const (
	OneForOne Strategy = iota // On termination only that child is restarted.
	OneForAll                 // On termination all children are restarted.
)

// Supervisor controls the execution and restart of a tree
// of supervisors and goroutines.
type Supervisor struct {
	id         string
	supervisor *Supervisor
	strategy   Strategy
	restarts   *restartFrequency
	messages   chan *message
	children   map[string]supervisable
	status     status
	terminate  chan bool
	err        error
}

// newSupervisor creates a new supervisor without backend loop.
func newSupervisor(id string, strategy Strategy, intensity int, period time.Duration) *Supervisor {
	return &Supervisor{
		id:        id,
		strategy:  strategy,
		restarts:  newRestartFrequency(intensity, period),
		messages:  make(chan *message),
		children:  make(map[string]supervisable),
		status:    stReady,
		terminate: make(chan bool),
	}
}

// NewSupervisor creates a new supervisor.
func NewSupervisor(id string, strategy Strategy, intensity int, period time.Duration) *Supervisor {
	sup := newSupervisor(id, strategy, intensity, period)
	sup.start()
	return sup
}

// Go starts the function as supervised goroutine with 
// the given id.
func (sup *Supervisor) Go(id string, sfunc SupervisedFunc) error {
	sf := &supervisableFunc{nil, sfunc, stReady}
	msg := newStartMsg(id, sf)
	sup.messages <- msg
	resp := <-msg.response
	if resp != nil {
		return resp.err()
	}
	return nil
}

// Supervisor creates a child supervisor with the given id.
func (sup *Supervisor) Supervisor(id string, strategy Strategy, intensity int, period time.Duration) (*Supervisor, error) {
	chsup := newSupervisor(id, strategy, intensity, period)
	msg := newStartMsg(id, chsup)
	sup.messages <- msg
	resp := <-msg.response
	if resp != nil {
		return nil, resp.err()
	}
	return chsup, nil
}

// Terminate tells a child to stop.
func (sup *Supervisor) Terminate(id string) error {
	if sup.status != stRunning {
		println(sup.status)
		return sup.Err()
	}
	msg := newTerminateMsg(id, nil)
	sup.messages <- msg
	resp := <-msg.response
	if resp != nil {
		return resp.err()
	}
	return nil
}

// Err returns the error status of the supervisor.
func (sup *Supervisor) Err() error {
	if sup.err == nil {
		return &StillRunningError{}
	}
	return sup.err
}

// Stop tells the supervisor to stop working.
func (sup *Supervisor) Stop() error {
	sup.stop()
	return sup.err
}

// String returns the hierarchical id of the supervisor.
func (sup *Supervisor) String() string {
	if sup.supervisor == nil {
		return fmt.Sprintf("/%s", sup.id)
	}
	return fmt.Sprintf("%s/%s", sup.supervisor, sup.id)
}

// handle returns the child handle.
func (sup *Supervisor) handle() *Handle {
	return &Handle{
		id:         sup.id,
		supervisor: sup.supervisor,
		terminate:  sup.terminate,
	}
}

// setHandle supplies the supervisor as child with the needed 
// informations.
func (sup *Supervisor) setHandle(h *Handle) {
	sup.id = h.id
	sup.supervisor = h.supervisor
	sup.terminate = h.terminate
}

// start runs the backend loop as goroutine.
func (sup *Supervisor) start() {
	if sup.status == stReady {
		sup.status = stRunning
		go sup.loop()
	}
}

// stop tells the supervisor to stop working.
func (sup *Supervisor) stop() {
	if sup.status == stRunning {
		sup.terminate <- true
	}
	sup.status = stReady
}

// loop is the backend loop of the supervisor.
func (sup *Supervisor) loop() {
	// Finalizing.
	defer sup.finish()
	// Start possible existing children after a restart.
	for _, child := range sup.children {
		child.start()
	}
	// Backend loop.
	for {
		select {
		case msg := <-sup.messages:
			switch msg.code {
			case msgStart:
				if sup.children[msg.id] != nil {
					msg.response <- newErrorMsg(sup.id, &InvalidIdError{true, msg.id})
					continue
				}
				cs := &Handle{
					id:         msg.id,
					supervisor: sup,
					terminate:  make(chan bool),
				}
				msg.sup.setHandle(cs)
				sup.children[msg.id] = msg.sup
				msg.sup.start()
				msg.response <- nil
			case msgTerminate:
				if sup.children[msg.id] == nil {
					msg.response <- newErrorMsg(sup.id, &InvalidIdError{false, msg.id})
					continue
				}
				sup.children[msg.id].stop()
				delete(sup.children, msg.id)
				msg.response <- nil
			case msgError:
				if msg.reason != nil {
					if err := sup.handleChildError(msg.id); err != nil {
						applog.Errorf("supervisor %q cannot handle error of child %q: %v", sup, msg.id, err)
						sup.err = err
						return
					}
				}
			}
		case <-sup.terminate:
			return
		}
	}
}

// finish does the cleanup when the supervisor terminates.
func (sup *Supervisor) finish() {
	// Clear message queue.
clean:
	for {
		select {
		case <-sup.messages:
		default:
			break clean
		}
	}
	// Allways stop the children.
	for _, child := range sup.children {
		child.stop()
	}
	// Check for error.
	if r := recover(); r != nil {
		sup.err = &TerminatedError{r}
		sup.status = stError
	} else {
		sup.status = stReady
	}
	// Notify parent supervisor if there's one.
	if sup.supervisor != nil {
		sup.supervisor.messages <- newErrorMsg(sup.id, sup.err)
	}
}

// handleChildError handles the error of a supervised child.
func (sup *Supervisor) handleChildError(id string) error {
	// Check restart frequency.
	if err := sup.restarts.check(); err != nil {
		return err
	}
	// Act depending on strategy.
	switch sup.strategy {
	case OneForOne:
		sup.children[id].stop()
		sup.children[id].start()
	case OneForAll:
		for _, child := range sup.children {
			child.stop()
		}
		for _, child := range sup.children {
			child.start()
		}
	}
	return nil
}

//--------------------
// ERRORS
//--------------------

// StillRunningError indicates a still running supervisor.
type StillRunningError struct{}

func (e *StillRunningError) Error() string {
	return fmt.Sprintf("supervisor is still running")
}

func IsStillRunningError(err error) bool {
	_, ok := err.(*StillRunningError)
	return ok
}

// InvalidIdError indicates the usage of an illegal child id.
type InvalidIdError struct {
	start bool
	Id    string
}

func (e *InvalidIdError) Error() string {
	if e.start {
		return fmt.Sprintf("child id %q is already in use", e.Id)
	}
	return fmt.Sprintf("child id %q is not in use", e.Id)
}

func IsInvalidIdError(e error) bool {
	_, ok := e.(*InvalidIdError)
	return ok
}

// TooMuchRestartsError shows that too much restarts happened in too short time.
type TooMuchRestartsError struct {
	Restarts int
	Period   time.Duration
}

func (e *TooMuchRestartsError) Error() string {
	return fmt.Sprintf("supervisor had %d restarts in %s", e.Restarts, e.Period)
}

func IsTooMuchRestartsError(e error) bool {
	_, ok := e.(*TooMuchRestartsError)
	return ok
}

// TerminatedError signals the termination of a supervisable
// due to the given reason.
type TerminatedError struct {
	Reason interface{}
}

func (e *TerminatedError) Error() string {
	return fmt.Sprintf("supervisor has terminated: %v", e.Reason)
}

func IsTerminatedError(err error) bool {
	_, ok := err.(*TerminatedError)
	return ok
}

// EOF
