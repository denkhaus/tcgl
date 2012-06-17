// Tideland Common Go Library - Cells
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
	"code.google.com/p/tcgl/applog"
	"code.google.com/p/tcgl/config"
	"code.google.com/p/tcgl/identifier"
	"code.google.com/p/tcgl/monitoring"
	"fmt"
	"runtime"
	"sync"
	"time"
)

//--------------------
// EVENT
//--------------------

// Event is anything that has a topic and a payload. Data to and
// between cells is passed as event.
type Event interface {
	// Topic returns the topic of the simple event.
	Topic() string
	// Payload returns the payload of the simple event.
	Payload() interface{}
	// Context returns the context of a set of event processings.
	Context() *Context
	// SetContext set the context of a set of event processings.
	SetContext(c *Context)
}

// simpleEvent can be used if no own event implementation is
// wanted or needed.
type simpleEvent struct {
	topic   string
	payload interface{}
	context *Context
}

// NewSimpleEvent creates a simple event.
func NewSimpleEvent(t string, p interface{}) Event {
	return &simpleEvent{t, p, nil}
}

// Topic returns the topic of the simple event.
func (se simpleEvent) Topic() string {
	return se.topic
}

// Payload returns the payload of the simple event.
func (se simpleEvent) Payload() interface{} {
	return se.payload
}

// Context returns the context of a set of event processings.
func (se simpleEvent) Context() *Context {
	return se.context
}

// SetContext set the context of a set of event processings.
func (se *simpleEvent) SetContext(c *Context) {
	se.context = c
}

//--------------------
// BEHAVIOR
//--------------------

// Behavior is the interface that has to be implemented
// of those behaviors which can be plugged into an environment
// for the real event processing.
type Behavior interface {
	// Init the deployed behavior inside an environment.
	Init(env *Environment, id Id) error
	// ProcessEvent processes an event and can emit own events.
	ProcessEvent(e Event, emitter EventEmitter)
	// Recover from an error during the processing of event e.
	Recover(r interface{}, e Event)
	// Stop the behavior.
	Stop()
}

// PoolableBehavior is the interface for behaviors which want
// to be pooled.
type PoolableBehavior interface {
	PoolConfig() (poolSize int, stateful bool)
}

// BehaviorFactory is a function that creates a behavior instance.
type BehaviorFactory func() Behavior

// BehaviorFactoryMap is a map of ids to behavior factories.
type BehaviorFactoryMap map[Id]BehaviorFactory

// poolBehavior manages a pool of behaviors and distributes the
// received events round robin.
type poolBehavior struct {
	cellPool chan *cell
}

// newPoolBehavior creates a new pool behavior with the passed size and
// the already created first behavior instance. It then creates the rest 
// of the behavior instances.
func newPoolBehavior(env *Environment, id Id, poolSize int, stateful bool, b Behavior, bf BehaviorFactory) (Behavior, error) {
	pb := &poolBehavior{make(chan *cell, poolSize)}
	c, err := newCell(env, id, b)
	if err != nil {
		return nil, err
	}
	pb.cellPool <- c
	for i := 1; i < poolSize; i++ {
		if stateful {
			// Stateful, so multiple instances.
			c, err = newCell(env, id, bf())
		} else {
			// Not stateful, the pool is sharing only one instance.
			c, err = newCell(env, id, b)
		}
		if err != nil {
			return nil, err
		}
		pb.cellPool <- c
	}
	return pb, nil
}

// Init the behavior by creating the cells for the buffer.
func (b *poolBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *poolBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	c := <-b.cellPool
	c.processEvent(e)
	b.cellPool <- c
}

// Recover from an error.
func (b *poolBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior, which means to stop all pooled cells.
func (b *poolBehavior) Stop() {
	for i := 0; i < len(b.cellPool); i++ {
		c := <-b.cellPool
		c.stop()
	}
	close(b.cellPool)
}

//--------------------
// ENVIRONMENT
//--------------------

// SubscriptionMap is a map of emitter ids to subscribed ids.
type SubscriptionMap map[Id][]Id

// Environment defines a common set of cells.
type Environment struct {
	mutex         sync.RWMutex
	id            Id
	configuration *config.Configuration
	cells         cellMap
	tickers       map[Id]*ticker
}

// NewEnvironment creates a new environment.
func NewEnvironment(id Id) *Environment {
	env := &Environment{
		id:      id,
		cells:   make(cellMap),
		tickers: make(map[Id]*ticker),
	}
	runtime.SetFinalizer(env, (*Environment).Shutdown)
	return env
}

// SetConfiguration sets the configuration of the environment.
func (env *Environment) SetConfiguration(configuration *config.Configuration) {
	env.configuration = configuration
}

// Configuration returns the configuration of the environment.
func (env *Environment) Configuration() *config.Configuration {
	return env.configuration
}

// AddCell adds a cell with a given id and its behavior factory.
func (env *Environment) AddCell(id Id, bf BehaviorFactory) (Behavior, error) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	return env.startCell(id, bf)
}

// AddCell adds a number of cells with a given ids and their behavior factories.
func (env *Environment) AddCells(bfm BehaviorFactoryMap) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	for id, bf := range bfm {
		if _, err := env.startCell(id, bf); err != nil {
			return err
		}
	}
	return nil
}

// startCell starts the cell with the behavior returned by the behavior factory.
func (env *Environment) startCell(id Id, bf BehaviorFactory) (Behavior, error) {
	if _, ok := env.cells[id]; ok {
		return nil, CellAlreadyExistsError{id}
	}
	// Check poolability.
	behavior := bf()
	if pb, ok := behavior.(PoolableBehavior); ok {
		var err error
		poolSize, stateful := pb.PoolConfig()
		behavior, err = newPoolBehavior(env, id, poolSize, stateful, behavior, bf)
		if err != nil {
			return nil, err
		}
	}
	// Create cell.
	c, err := newCell(env, id, behavior)
	if err != nil {
		return nil, err
	}
	env.cells[id] = c
	return behavior, nil
}

// RemoveCell removes the cell with the given id.
func (env *Environment) RemoveCell(id Id) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	if c, ok := env.cells[id]; ok {
		delete(env.cells, id)
		c.stop()
	}
}

// HasCell returns true if the cell with the given id exists.
func (env *Environment) HasCell(id Id) bool {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	_, ok := env.cells[id]
	return ok
}

// CellBehavior returns the behavior with the id or nil (take care).
func (env *Environment) CellBehavior(id Id) (Behavior, error) {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	if c, ok := env.cells[id]; ok {
		return c.behavior, nil
	}
	return nil, CellDoesNotExistError{id}
}

// Subscribe assigns cells as receivers of the emitted 
// events of the first cell.
func (env *Environment) Subscribe(emitterId Id, subscriberIds ...Id) error {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	return env.subscribe(emitterId, subscriberIds...)
}

// SubscribeAll assigns all subscribers to the emitters of the map.
func (env *Environment) SubscribeAll(sm SubscriptionMap) error {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	for emitterId, subscriberIds := range sm {
		if err := env.subscribe(emitterId, subscriberIds...); err != nil {
			return err
		}
	}
	return nil
}

// subscribe performs the subscription in a read-locked environment state.
func (env *Environment) subscribe(emitterId Id, subscriberIds ...Id) error {
	if c, ok := env.cells[emitterId]; ok {
		subscriberCells, err := env.cells.subset(subscriberIds...)
		if err != nil {
			return err
		}
		return c.changeSubscriptions(true, subscriberCells)
	}
	return CellDoesNotExistError{emitterId}
}

// Unsubscribe removes the assignment of emitting und subscribed cells. 
func (env *Environment) Unsubscribe(emitterId Id, unsubscriberIds ...Id) error {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	if c, ok := env.cells[emitterId]; ok {
		unsubscriberCells, err := env.cells.subset(unsubscriberIds...)
		if err != nil {
			return err
		}
		return c.changeSubscriptions(false, unsubscriberCells)
	}
	return CellDoesNotExistError{emitterId}
}

// Emit emits an event to the cell with a given id and returns its
// (possibly new created) context.
func (env *Environment) Emit(id Id, e Event) (ctx *Context, err error) {
	defer func() {
		if err != nil {
			applog.Errorf("can't emit topic %q to %q: %v", e.Topic(), id, err)
		}
	}()
	sleep := 5
	for {
		env.mutex.RLock()
		c, ok := env.cells[id]
		env.mutex.RUnlock()
		if ok {
			if e.Context() == nil {
				e.SetContext(newContext())
			}
			e.Context().incrActivity()
			if err := c.processEvent(e); err != nil {
				return nil, err
			}
			return e.Context(), nil
		}
		// Wait an increasing time befor retry, max 5 seconds.
		if sleep <= 5000 {
			time.Sleep(time.Duration(sleep) * time.Millisecond)
			sleep *= 10
		}
	}
	return nil, CellDoesNotExistError{id}
}

// EmitSimple is a convenience method emitting a simple event in one call.
func (env *Environment) EmitSimple(id Id, t string, p interface{}) (*Context, error) {
	return env.Emit(id, NewSimpleEvent(t, p))
}

// AddTicker adds a new ticker for periodical ticker events with the given
// id to the emitId.
func (env *Environment) AddTicker(id, emitId Id, period time.Duration) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	if _, ok := env.tickers[id]; ok {
		return fmt.Errorf("ticker with id %q already added", id)
	}
	env.tickers[id] = startTicker(env, id, emitId, period)
	return nil
}

// RemoveTicker removes a periodical ticker event.
func (env *Environment) RemoveTicker(id Id) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	if ticker, ok := env.tickers[id]; ok {
		ticker.stop()
		delete(env.tickers, id)
		return nil
	}
	return fmt.Errorf("ticker with id %q does not exist", id)
}

// Shutdown manages the proper finalization of an environment.
func (env *Environment) Shutdown() error {
	// Stop all tickers.
	for _, ticker := range env.tickers {
		ticker.stop()
	}
	env.tickers = nil
	// Stop all cells and delete registry.
	for _, c := range env.cells {
		c.stop()
	}
	env.cells = nil
	runtime.SetFinalizer(env, nil)
	return nil
}

//--------------------
// EVENT EMITTER
//--------------------

// EventEmitter is any type that can be used for emitting events.
type EventEmitter interface {
	// Emit emits an event to a number of cells depending on the implementation.
	Emit(e Event)
	// EmitSimple emits convieniently a simple event.
	EmitSimple(topic string, payload interface{})
}

// cellEventEmitter implements EventEmitter for the processing
// of an event in a cell.
type cellEventEmitter struct {
	cells   cellMap
	context *Context
}

// Emit emits an event to the subscribers of a cell. It passes
// the context to that event.
func (cee *cellEventEmitter) Emit(e Event) {
	e.SetContext(cee.context)
	e.Context().incrActivity()
	erroneousSubscriberIds := []Id{}
	for id, sc := range cee.cells {
		if err := sc.processEvent(e); err != nil {
			erroneousSubscriberIds = append(erroneousSubscriberIds, id)
		}
	}
	for _, id := range erroneousSubscriberIds {
		delete(cee.cells, id)
	}
}

// EmitSimple emits convieniently a simple event to the subscribers
// of a cell. It passes the context to that event.
func (cee *cellEventEmitter) EmitSimple(topic string, payload interface{}) {
	cee.Emit(NewSimpleEvent(topic, payload))
}

//--------------------
// CELL
//--------------------

// cell for event processing.
type cell struct {
	env         *Environment
	id          Id
	behavior    Behavior
	subscribers cellMap
	queue       *cellMessageQueue
	measuringId string
}

// newCell create a new cell around a behavior.
func newCell(env *Environment, id Id, b Behavior) (*cell, error) {
	c := &cell{
		env:         env,
		id:          id,
		behavior:    b,
		subscribers: make(cellMap),
		queue:       newCellMessageQueue(),
		measuringId: identifier.Identifier("cells", env.id, "cell", identifier.TypeAsIdentifierPart(b)),
	}
	// Init behavior.
	if err := b.Init(env, id); err != nil {
		return nil, CellInitError{id, err}
	}
	go c.processLoop()
	monitoring.IncrVariable(identifier.Identifier("cells", c.env.id, "total-cells"))
	monitoring.IncrVariable(c.measuringId)
	return c, nil
}

// stop terminates the cell.
func (c *cell) stop() {
	c.queue.push(nil, nil, false)
}

// changeSubscriptions tells the cell to change subscribers.
func (c *cell) changeSubscriptions(add bool, cells cellMap) error {
	return c.queue.push(nil, cells, add)
}

// processEvent tells the cell to handle an event.
func (c *cell) processEvent(e Event) error {
	return c.queue.push(e, nil, false)
}

// processLoop is the backend for the processing of events.
func (c *cell) processLoop() {
	for {
		message := c.queue.pull()
		switch {
		case message.event != nil:
			// Process the event.
			c.process(message.event)
		case message.cells != nil:
			// Change the subscriptions.
			for id, sc := range message.cells {
				if message.add {
					c.subscribers[id] = sc
				} else {
					delete(c.subscribers, id)
				}
			}
		case message.event == nil && message.cells == nil:
			// Stop the cell.
			c.queue.close()
			break
		}
	}
	monitoring.DecrVariable(c.measuringId)
	monitoring.DecrVariable(identifier.Identifier("cells", c.env.id, "total-cells"))
	c.behavior.Stop()
}

// process encapsulates event processing including error 
// recovery and measuring.
func (c *cell) process(e Event) {
	// Error recovering.
	defer func() {
		if r := recover(); r != nil {
			if e != nil {
				applog.Errorf("cell %q has error '%v' with event '%+v'", c.id, r, EventString(e))

			} else {
				applog.Errorf("cell %q has error '%v'", c.id, r)
			}
			c.behavior.Recover(r, e)
		}
	}()
	defer e.Context().decrActivity()
	// Handle the event inside a measuring.
	measuring := monitoring.BeginMeasuring(c.measuringId)
	emitter := &cellEventEmitter{c.subscribers, e.Context()}
	c.behavior.ProcessEvent(e, emitter)
	measuring.EndMeasuring()
}

// EOF
