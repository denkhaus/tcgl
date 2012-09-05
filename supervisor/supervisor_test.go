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
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

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
	results := starts{}
	child := func(h *supervisor.Handle) error { return selectChild(h, results) }

	sup.Go("alpha", child)

	time.Sleep(100 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(results["alpha"], 1, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'select'")
}

// TestFuncMethod tests the termination of a method child on demand.
func TestFuncMethod(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("method", supervisor.OneForOne, 5, time.Second)
	results := starts{}
	child := func(h *supervisor.Handle) error { return methodChild(h, results) }

	sup.Go("alpha", child)

	time.Sleep(100 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(results["alpha"], 1, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'method'")
}

// TestFuncPanic tests the panic of a child.
func TestFuncPanic(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("panic", supervisor.OneForOne, 5, time.Second)
	results := starts{}
	child := func(h *supervisor.Handle) error { return panicChild(h, results) }

	sup.Go("alpha", child)

	time.Sleep(125 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(results["alpha"], 3, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'panic'")
}

// TestFuncError tests the error of a child.
func TestFuncError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("error", supervisor.OneForOne, 5, time.Second)
	results := starts{}
	child := func(h *supervisor.Handle) error { return errorChild(h, results) }

	sup.Go("alpha", child)

	time.Sleep(125 * time.Millisecond)

	err := sup.Terminate("alpha")
	assert.Nil(err, "termination of 'alpha'")
	assert.Equal(results["alpha"], 2, "starts of 'alpha'")

	err = sup.Stop()
	assert.Nil(err, "stopping of 'error'")
}

// TestFuncsOneForOne tests multiple childs restarting one for one.
func TestFuncsOneForOne(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("one4one", supervisor.OneForOne, 10, time.Second)
	results := starts{}
	childA := func(h *supervisor.Handle) error { return selectChild(h, results) }
	childB := func(h *supervisor.Handle) error { return methodChild(h, results) }
	childC := func(h *supervisor.Handle) error { return panicChild(h, results) }
	childD := func(h *supervisor.Handle) error { return errorChild(h, results) }

	sup.Go("alpha", childA)
	sup.Go("beta", childB)
	sup.Go("gamma", childC)
	sup.Go("delta", childD)

	time.Sleep(125 * time.Millisecond)

	err := sup.Stop()
	assert.Nil(err, "stopping of 'one4one'")
	assert.Equal(results["alpha"], 1, "starts of 'alpha'")
	assert.Equal(results["beta"], 1, "starts of 'beta'")
	assert.Equal(results["gamma"], 3, "starts of 'gamma'")
	assert.Equal(results["delta"], 2, "starts of 'delta'")
}

// TestFuncsOneForAll tests multiple childs restarting one for all.
func TestFuncsOneForAll(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("one4all", supervisor.OneForAll, 10, time.Second)
	results := starts{}
	childA := func(h *supervisor.Handle) error { return selectChild(h, results) }
	childB := func(h *supervisor.Handle) error { return methodChild(h, results) }
	childC := func(h *supervisor.Handle) error { return panicChild(h, results) }

	sup.Go("alpha", childA)
	sup.Go("beta", childB)
	sup.Go("gamma", childC)

	time.Sleep(125 * time.Millisecond)

	err := sup.Stop()
	assert.Nil(err, "stopping of 'one4all'")
	assert.Equal(results["alpha"], 3, "starts of 'alpha'")
	assert.Equal(results["beta"], 3, "starts of 'beta'")
	assert.Equal(results["gamma"], 3, "starts of 'gamma'")
}

// TestChildSupervisor tests a supervisor as a child.
func TestChildSupervisor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("parent", supervisor.OneForAll, 3, time.Second)
	chsup, _ := sup.Supervisor("child", supervisor.OneForOne, 5, 250*time.Millisecond)
	results := starts{}
	childA := func(h *supervisor.Handle) error { return methodChild(h, results) }
	childB := func(h *supervisor.Handle) error { return panicChild(h, results) }

	sup.Go("alpha", childA)
	sup.Go("beta", childA)

	chsup.Go("gamma", childA)
	chsup.Go("delta", childA)
	chsup.Go("epsilon", childB)

	time.Sleep(2 * time.Second)

	err := sup.Stop()
	assert.ErrorMatch(err, `supervisor had .* restarts in .*`, "stopping of 'child-supervisor'")
	assert.Equal(results["alpha"], 4, "starts of 'alpha'")
	assert.Equal(results["beta"], 4, "starts of 'beta'")
	assert.Equal(results["gamma"], 4, "starts of 'gamma'")
	assert.Equal(results["delta"], 4, "starts of 'delta'")
	assert.Equal(results["epsilon"], 24, "starts of 'epsilon'")
}

// TestSupervisorTree tests a tree of supervisors.
func TestSupervisorTree(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sup := supervisor.NewSupervisor("parent", supervisor.OneForAll, 2, time.Second)
	chsup, _ := sup.Supervisor("child", supervisor.OneForAll, 3, time.Second)
	gchsup, _ := chsup.Supervisor("grandchild", supervisor.OneForOne, 3, 250*time.Millisecond)
	results := starts{}
	childA := func(h *supervisor.Handle) error { return methodChild(h, results) }
	childB := func(h *supervisor.Handle) error { return panicChild(h, results) }

	sup.Go("alpha", childA)
	chsup.Go("beta", childA)
	gchsup.Go("gamma", childB)

	time.Sleep(3 * time.Second)

	err := sup.Stop()
	assert.ErrorMatch(err, `supervisor had .* restarts in .*`, "stopping of 'supervisor-tree'")
	assert.Equal(results["alpha"], 3, "starts of 'alpha'")
	assert.Equal(results["beta"], 12, "starts of 'beta'")
	assert.Equal(results["gamma"], 48, "starts of 'gamma'")
}

//--------------------
// HELPER
//--------------------

// starts collects goroutine starts.
type starts map[string]int

func (s starts) incr(h *supervisor.Handle) {
	id := h.Id()
	s[id]++
}

// selectChild works in a select loop until terminated.
func selectChild(h *supervisor.Handle, s starts) error {
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
func methodChild(h *supervisor.Handle, s starts) error {
	s.incr(h)
	for !h.IsTerminated() {
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// panicChild produces a panic after 5 iterations.
func panicChild(h *supervisor.Handle, s starts) error {
	s.incr(h)
	counter := 0
	for {
		counter++
		select {
		case <-h.Terminate():
			return nil
		case <-time.After(10 * time.Millisecond):
		}
		if counter == 5 {
			panic("panic!")
		}
	}
	return nil
}

// errorChild returns an error after 10 iterations.
func errorChild(h *supervisor.Handle, s starts) error {
	s.incr(h)
	counter := 0
	for {
		counter++
		select {
		case <-h.Terminate():
			return nil
		case <-time.After(10 * time.Millisecond):
		}
		if counter == 10 {
			return fmt.Errorf("error!")
		}
	}
	return nil
}

// EOF
