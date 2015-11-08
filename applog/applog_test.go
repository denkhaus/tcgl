// Tideland Common Go Library - Application Log - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package applog_test

//--------------------
// IMPORTS
//--------------------

import (
	"log"
	"os"
	"testing"

	"github.com/denkhaus/tcgl/applog"
	"github.com/denkhaus/tcgl/asserts"
)

//--------------------
// TESTS
//--------------------

// Test log level.
func TestLevel(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	applog.SetLevel(applog.LevelDebug)
	assert.Equal(applog.Level(), applog.LevelDebug, "Level debug.")
	applog.SetLevel(applog.LevelCritical)
	assert.Equal(applog.Level(), applog.LevelCritical, "Level critical.")
	applog.SetLevel(applog.LevelDebug)
	assert.Equal(applog.Level(), applog.LevelDebug, "Level debug.")
}

// Test debugging.
func TestDebug(t *testing.T) {
	applog.Debugf("Hello, I'm debugging %v!", "here")
	applog.SetLevel(applog.LevelError)
	applog.Debugf("Should not be shown!")
}

// Test log at all levels.
func TestAllLevels(t *testing.T) {
	applog.SetLevel(applog.LevelDebug)

	applog.Debugf("Debug.")
	applog.Infof("Info.")
	applog.Warningf("Warning.")
	applog.Errorf("Error.")
	applog.Criticalf("Critical.")
}

// Test logging from level warning and above.
func TestWarningAndAbove(t *testing.T) {
	applog.SetLevel(applog.LevelWarning)

	applog.Debugf("Debug.")
	applog.Infof("Info.")
	applog.Warningf("Warning.")
	applog.Errorf("Error.")
	applog.Criticalf("Critical.")
}

// Test logging with the go logger.
func TestGoLogger(t *testing.T) {
	log.SetOutput(os.Stdout)

	applog.SetLevel(applog.LevelDebug)
	applog.SetLogger(applog.GoLogger{})

	applog.Debugf("Debug.")
	applog.Infof("Info.")
	applog.Warningf("Warning.")
	applog.Errorf("Error.")
	applog.Criticalf("Critical.")
}

// Test logging with an own logger.
func TestOwnLogger(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	ownLogger := &testLogger{[]string{}}

	applog.SetLevel(applog.LevelDebug)
	applog.SetLogger(ownLogger)

	applog.Debugf("Debug.")
	applog.Infof("Info.")
	applog.Warningf("Warning.")
	applog.Errorf("Error.")
	applog.Criticalf("Critical.")

	assert.Length(ownLogger.logs, 5, "Everything logged.")
}

//--------------------
// LOGGER
//--------------------

type testLogger struct {
	logs []string
}

func (tl *testLogger) Debug(info, msg string) {
	tl.logs = append(tl.logs, info+" "+msg)
}

func (tl *testLogger) Info(info, msg string) {
	tl.logs = append(tl.logs, info+" "+msg)
}
func (tl *testLogger) Warning(info, msg string) {
	tl.logs = append(tl.logs, info+" "+msg)
}
func (tl *testLogger) Error(info, msg string) {
	tl.logs = append(tl.logs, info+" "+msg)
}
func (tl *testLogger) Critical(info, msg string) {
	tl.logs = append(tl.logs, info+" "+msg)
}

// EOF
