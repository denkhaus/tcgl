// Tideland Common Go Library - Utilities
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package util

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"reflect"
)

//--------------------
// METHOD DISPATCHING
//--------------------

// Dispatch a string to a method of a type.
func Dispatch(variable interface{}, name string, args ...interface{}) (interface{}, error) {
	numArgs := len(args)
	value := reflect.ValueOf(variable)
	valueType := value.Type()
	numMethods := valueType.NumMethod()
	// Search mathching method and call it.
	for i := 0; i < numMethods; i++ {
		method := valueType.Method(i)
		if (method.PkgPath == "") && (method.Type.NumIn() == numArgs+1) {
			if method.Name == name {
				// Prepare all args with variable and args.
				callArgs := make([]reflect.Value, numArgs+1)
				callArgs[0] = value
				for i, a := range args {
					callArgs[i+1] = reflect.ValueOf(a)
				}
				// Make the function call.
				results := method.Func.Call(callArgs)
				// Transfer results into slice of interfaces.
				l := len(results)
				var retVal interface{}
				if l == 1 {
					retVal = results[0].Interface()
				} else if l > 1 {
					tmpRetVal := make([]interface{}, l)
					for i, v := range results {
						tmpRetVal[i] = v.Interface()
					}
					retVal = tmpRetVal
				}
				return retVal, nil
			}
		}
	}
	return nil, fmt.Errorf("method %q with %d arguments not found", name, len(args))
}

//--------------------
// LAZY EVALUATOR BUILDERS
//--------------------

// Function to evaluate.
type EvalFunc func(interface{}) (interface{}, interface{})

// Generic builder for lazy evaluators.
func BuildLazyEvaluator(evalFunc EvalFunc, initState interface{}) func() interface{} {
	retValChan := make(chan interface{})
	loopFunc := func() {
		var actState interface{} = initState
		var retVal interface{}

		for {
			retVal, actState = evalFunc(actState)
			retValChan <- retVal
		}
	}
	retFunc := func() interface{} {
		return <-retValChan
	}
	go loopFunc()
	return retFunc
}

// Builder for lazy evaluators with ints as result.
func BuildLazyIntEvaluator(evalFunc EvalFunc, initState interface{}) func() int {
	ef := BuildLazyEvaluator(evalFunc, initState)
	return func() int {
		return ef().(int)
	}
}

// EOF
