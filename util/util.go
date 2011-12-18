// Tideland Common Go Library - Utilities
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package util

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"runtime"
	"strings"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Utilities - Release 2011-09-22"

//--------------------
// DEBUGGING
//--------------------

// Debug prints a debug information to the log with file and line.
func Debug(format string, args ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	info := fmt.Sprintf(format, args...)

	log.Printf("[cgl] debug %s:%s:%d %v", fileName, funcName, line, info)
}

//--------------------
// METHOD DISPATCHING
//--------------------

// Dispatch a string to a method of a type.
func Dispatch(variable interface{}, name string, args ...interface{}) ([]interface{}, bool) {
	numArgs := len(args)
	value := reflect.ValueOf(variable)
	valueType := value.Type()
	numMethods := valueType.NumMethod()

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

				allResults := make([]interface{}, len(results))

				for i, v := range results {
					allResults[i] = v.Interface()
				}

				return allResults, true
			}
		}
	}

	return nil, false
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
