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
	"io"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Utilities - Release 2012-01-30"

//--------------------
// DEBUGGING
//--------------------

// Debugf prints a debug information to the log with file and line.
func Debugf(format string, args ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	info := fmt.Sprintf(format, args...)
	logger := NewDefaultLogger("cgl")

	logger.Debugf("%s:%s:%d %v", fileName, funcName, line, info)
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

//--------------------
// LOGGER
//--------------------

// Log levels to control the logging output.
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelCritical
)

// logLevel controls the global log level used by the logger.
var logLevel = LogLevelDebug

// LogLevel returns the global log level and can be used in
// own implementations of the logger interface.
func LogLevel() int {
	return logLevel
}

// SetLogLevel sets the global log level used by the simple
// logger.
func SetLogLevel(level int) {
	logLevel = level
}

// Logger is the interface for different logger implementations.
type Logger interface {
	// Debugf logs a message at debug level.
	Debugf(format string, args ...interface{})
	// Infof logs a message at info level.
	Infof(format string, args ...interface{})
	// Warningf logs a message at warning level.
	Warningf(format string, args ...interface{})
	// Errorf logs a message at error level.
	Errorf(format string, args ...interface{})
	// Criticalf logs a message at critical level.
	Criticalf(format string, args ...interface{})
}

// simpleLogger is a logger implementation using the log package.
type simpleLogger struct {
	logger *log.Logger
}

// NewSimpleLogger creates a logger using the log package.
func NewSimpleLogger(out io.Writer, prefix string, flag int) Logger {
	return &simpleLogger{log.New(out, "["+prefix+"] ", flag)}
}

// NewDefaultLogger create a simple logger on stdout with
// printig of date and time.
func NewDefaultLogger(prefix string) Logger {
	return NewSimpleLogger(os.Stdout, prefix, log.Ldate|log.Ltime)
}

// Debugf logs a message at debug level.
func (sl simpleLogger) Debugf(format string, args ...interface{}) {
	if logLevel <= LogLevelDebug {
		sl.logger.Printf("[debug] "+format, args...)
	}
}

// Infof logs a message at info level.
func (sl simpleLogger) Infof(format string, args ...interface{}) {
	if logLevel <= LogLevelInfo {
		sl.logger.Printf("[info] "+format, args...)
	}
}

// Warningf logs a message at warning level.
func (sl simpleLogger) Warningf(format string, args ...interface{}) {
	if logLevel <= LogLevelWarning {
		sl.logger.Printf("[warning] "+format, args...)
	}
}

// Errorf logs a message at error level.
func (sl simpleLogger) Errorf(format string, args ...interface{}) {
	if logLevel <= LogLevelError {
		sl.logger.Printf("[error] "+format, args...)
	}
}

// Criticalf logs a message at critical level.
func (sl simpleLogger) Criticalf(format string, args ...interface{}) {
	sl.logger.Printf("[critical] "+format, args...)
}

// EOF
