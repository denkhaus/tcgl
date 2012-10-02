// Tideland Common Go Library - Supervisor - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package supervisor_test

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/supervisor"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

//--------------------
// VARS
//--------------------

var (
	shortWait = 100 * time.Millisecond
)

//--------------------
// TESTS
//--------------------

// TestChildren tests the retrieval of the children ids.
func TestChildren(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("children", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	child := func(h *supervisor.Handle) error { return selectChild(h, st) }

	sup.Go("alpha", child)
	sup.Go("beta", child)
	sup.Go("gamma", child)

	children := sup.Children()
	sort.Strings(children)
	assert.Equal(children, []string{"alpha", "beta", "gamma"}, "all children")

	sup.Terminate("beta")
	children = sup.Children()
	sort.Strings(children)
	assert.Equal(children, []string{"alpha", "gamma"}, "children w/o 'beta'")

	err := sup.Stop()
	assert.Nil(err, "stopping of 'children'")
}

// TestIllegalTerminate tests the termination of an illegal child.
func TestIllegalTerminate(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("terminate", supervisor.OneForOne, 5, time.Second)

	err := sup.Terminate("non-existing")
	assert.ErrorMatch(err, `child id ".*" is not in use`, "termination of 'non-existing'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'terminate'")
}

// TestFuncSelect tests the termination of a select child on demand.
func TestFuncSelect(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("select", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	child := func(h *supervisor.Handle) error { return selectChild(h, st) }

	sup.Go("alpha", child)

	time.Sleep(100 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(st.count("alpha"), 1, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'select'")
}

// TestFuncMethod tests the termination of a method child on demand.
func TestFuncMethod(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("method", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	child := func(h *supervisor.Handle) error { return methodChild(h, st) }

	sup.Go("alpha", child)

	time.Sleep(100 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(st.count("alpha"), 1, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'method'")
}

// TestFuncPanic tests the panic of a child.
func TestFuncPanic(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("panic", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	child := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }

	sup.Go("alpha", child)

	time.Sleep(325 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(st.count("alpha"), 4, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'panic'")
}

// TestFuncError tests the error of a child.
func TestFuncError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("error", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	child := func(h *supervisor.Handle) error { return errorChild(h, st, shortWait) }

	sup.Go("alpha", child)

	time.Sleep(325 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(st.count("alpha"), 4, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'error'")
}

// TestFuncsOneForOne tests multiple childs restarting one for one.
func TestFuncsOneForOne(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("one4one", supervisor.OneForOne, 25, time.Second)
	st := newStarts()
	childA := func(h *supervisor.Handle) error { return selectChild(h, st) }
	childB := func(h *supervisor.Handle) error { return methodChild(h, st) }
	childC := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }
	childD := func(h *supervisor.Handle) error { return errorChild(h, st, shortWait) }

	sup.Go("alpha", childA)
	sup.Go("beta", childB)
	sup.Go("gamma", childC)
	sup.Go("delta", childD)

	time.Sleep(time.Second)

	err := sup.Stop()
	assert.Nil(err, "stopping of 'one4one'")
	assert.Equal(st.count("alpha"), 1, "starts of 'alpha'")
	assert.Equal(st.count("beta"), 1, "starts of 'beta'")
	assert.Equal(st.count("gamma"), 10, "starts of 'gamma'")
	assert.Equal(st.count("delta"), 10, "starts of 'delta'")
}

// TestFuncsOneForAll tests multiple childs restarting one for all.
func TestFuncsOneForAll(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("one4all", supervisor.OneForAll, 25, time.Second)
	st := newStarts()
	childA := func(h *supervisor.Handle) error { return selectChild(h, st) }
	childB := func(h *supervisor.Handle) error { return methodChild(h, st) }
	childC := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }

	sup.Go("alpha", childA)
	sup.Go("beta", childB)
	sup.Go("gamma", childC)

	time.Sleep(time.Second)

	err := sup.Stop()
	assert.Nil(err, "stopping of 'one4all'")
	assert.Equal(st.count("alpha"), 10, "starts of 'alpha'")
	assert.Equal(st.count("beta"), 10, "starts of 'beta'")
	assert.Equal(st.count("gamma"), 10, "starts of 'gamma'")
}

// TestStampede tests a panic with strategy one for all and a large number of children.
func TestStampede(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("stampede", supervisor.OneForAll, 5, time.Second)
	st := newStarts()
	childA := func(h *supervisor.Handle) error { return methodChild(h, st) }
	childB := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }
	count := 1000

	for i := 0; i < count; i++ {
		id := fmt.Sprintf("alpha-%d", i)
		sup.Go(id, childA)
	}
	sup.Go("beta", childB)

	time.Sleep(2 * time.Second)

	err := sup.Stop()
	assert.Nil(err, "stopping of 'stampede'")
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("alpha-%d", i)
		assert.True(st.count(id) >= 1, "starts of child '"+id+"'")
	}
}

// TestChildSupervisor tests a supervisor as a child.
func TestChildSupervisor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("parent", supervisor.OneForAll, 3, time.Second)
	chsup, _ := sup.Supervisor("child", supervisor.OneForOne, 5, time.Second)
	st := newStarts()
	childA := func(h *supervisor.Handle) error { return methodChild(h, st) }
	childB := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }

	sup.Go("alpha", childA)
	sup.Go("beta", childA)

	chsup.Go("gamma", childA)
	chsup.Go("delta", childA)
	chsup.Go("epsilon", childB)

	time.Sleep(2 * time.Second)

	err := sup.Stop()
	assert.Nil(err, "stopping parent of 'child-supervisor'")
	assert.Equal(st.count("alpha"), 4, "starts of 'alpha'")
	assert.Equal(st.count("beta"), 4, "starts of 'beta'")
	assert.Equal(st.count("gamma"), 4, "starts of 'gamma'")
	assert.Equal(st.count("delta"), 4, "starts of 'delta'")
	assert.True(st.count("epsilon") > 1, "starts of 'epsilon'")
}

// TestSupervisorTree tests a tree of supervisors.
func TestSupervisorTree(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("parent", supervisor.OneForAll, 3, 5*time.Second)
	chsup, _ := sup.Supervisor("child", supervisor.OneForAll, 3, time.Second)
	gchsup, _ := chsup.Supervisor("grandchild", supervisor.OneForOne, 3, time.Second)
	st := newStarts()
	childA := func(h *supervisor.Handle) error { return methodChild(h, st) }
	childB := func(h *supervisor.Handle) error { return panicChild(h, st, shortWait) }

	sup.Go("alpha", childA)
	chsup.Go("beta", childA)
	gchsup.Go("gamma", childB)

	time.Sleep(10 * time.Second)

	err := sup.Stop()
	assert.ErrorMatch(err, `supervisor had .* restarts in .*`, "stopping of 'supervisor-tree'")
	assert.Equal(st.count("alpha"), 4, "starts of 'alpha'")
	assert.Equal(st.count("beta"), 16, "starts of 'beta'")
	assert.Equal(st.count("gamma"), 64, "starts of 'gamma'")
}

//--------------------
// HELPER
//--------------------

// starts collects goroutine starts.
type starts struct {
	mutex   sync.Mutex
	counter map[string]int
}

func newStarts() *starts {
	return &starts{
		counter: make(map[string]int),
	}
}

func (s *starts) incr(h *supervisor.Handle) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	id := h.Id()
	s.counter[id]++
}

func (s *starts) count(id string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.counter[id]
}

// selectChild works in a select loop until terminated.
func selectChild(h *supervisor.Handle, s *starts) error {
	s.incr(h)
	for {
		select {
		case <-h.Terminate():
			return nil
		case <-time.After(10 * time.Millisecond):
		}
	}
	return nil
}

// methodChild loops until the handle method signals termination.
func methodChild(h *supervisor.Handle, s *starts) error {
	s.incr(h)
	for !h.IsTerminated() {
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// panicChild produces a panic after a given time.
func panicChild(h *supervisor.Handle, s *starts, t time.Duration) error {
	s.incr(h)
	select {
	case <-h.Terminate():
		return nil
	case <-time.After(t):
		panic("panic!")
	}
	return nil
}

// errorChild returns an error after a given time.
func errorChild(h *supervisor.Handle, s *starts, t time.Duration) error {
	s.incr(h)
	select {
	case <-h.Terminate():
		return nil
	case <-time.After(t):
		return fmt.Errorf("error!")
	}
	return nil
}

// EOF
