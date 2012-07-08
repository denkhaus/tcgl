// Tideland Common Go Library - Configuration - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package config_test

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/config"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// TestSimple tests simple value retrieved as string.
func TestSimple(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", "quick brown fox")
	cfg.Set("bravo", true)
	cfg.Set("charlie", 4711)
	cfg.Set("delta", 47.11)

	// Successful gets.
	value, err := cfg.Get("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, "quick brown fox", "Right value 'alpha' returned.")
	value, err = cfg.Get("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, "true", "Right value 'bravo' returned.")
	value, err = cfg.Get("charlie")
	assert.Nil(err, "No error.")
	assert.Equal(value, "4711", "Right value 'charlie' returned.")
	value, err = cfg.Get("delta")
	assert.Nil(err, "No error.")
	assert.Equal(value, "47.11", "Right value 'delta' returned.")

	// Non-existing key.
	_, err = cfg.Get("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetDefault("non-existing-key", "default value")
	assert.Nil(err, "No error.")
	assert.Equal(value, "default value", "Right value 'non-existing-key' returned.")
}

// TestBool tests boolean values.
func TestBool(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", true)
	cfg.Set("bravo", "true")
	cfg.Set("charlie", "T")
	cfg.Set("delta", 1)
	cfg.Set("echo", false)
	cfg.Set("foxtrot", "false")
	cfg.Set("golf", "f")
	cfg.Set("hotel", 0)

	// Successful gets.
	value, err := cfg.GetBool("alpha")
	assert.Nil(err, "No error.")
	assert.True(value, "Right value 'alpha' returned.")
	value, err = cfg.GetBool("bravo")
	assert.Nil(err, "No error.")
	assert.True(value, "Right value 'bravo' returned.")
	value, err = cfg.GetBool("charlie")
	assert.Nil(err, "No error.")
	assert.True(value, "Right value 'charlie' returned.")
	value, err = cfg.GetBool("delta")
	assert.Nil(err, "No error.")
	assert.True(value, "Right value 'delta' returned.")
	value, err = cfg.GetBool("echo")
	assert.Nil(err, "No error.")
	assert.False(value, "Right value 'echo' returned.")
	value, err = cfg.GetBool("foxtrot")
	assert.Nil(err, "No error.")
	assert.False(value, "Right value 'foxtrot' returned.")
	value, err = cfg.GetBool("golf")
	assert.Nil(err, "No error.")
	assert.False(value, "Right value 'golf' returned.")
	value, err = cfg.GetBool("hotel")
	assert.Nil(err, "No error.")
	assert.False(value, "Right value 'hotel' returned.")

	// Illegal format.
	cfg.Set("illegal-format", 47.11)
	_, err = cfg.GetBool("illegal-format")
	assert.ErrorMatch(err, `invalid type "bool" for "47.11" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetBool("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetBoolDefault("non-existing-key", true)
	assert.Nil(err, "No error.")
	assert.True(value, "Right value 'non-existing-key' returned.")
}

// TestInt tests int values.
func TestInt(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", 4711)
	cfg.Set("bravo", -4711)
	cfg.Set("charlie", 0)

	// Successful gets.
	value, err := cfg.GetInt("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, 4711, "Right value 'alpha' returned.")
	value, err = cfg.GetInt("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, -4711, "Right value 'bravo' returned.")
	value, err = cfg.GetInt("charlie")
	assert.Nil(err, "No error.")
	assert.Equal(value, 0, "Right value 'charlie' returned.")

	// Illegal format.
	cfg.Set("illegal-format", 47.11)
	_, err = cfg.GetInt("illegal-format")
	assert.ErrorMatch(err, `invalid type "int" for "47.11" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetInt("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetIntDefault("non-existing-key", 4711)
	assert.Nil(err, "No error.")
	assert.Equal(value, 4711, "Right value 'non-existing-key' returned.")
}

// TestInt64 tests int64 values.
func TestInt64(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", 4711)
	cfg.Set("bravo", -4711)
	cfg.Set("charlie", 0)

	// Successful gets.
	value, err := cfg.GetInt64("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, int64(4711), "Right value 'alpha' returned.")
	value, err = cfg.GetInt64("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, int64(-4711), "Right value 'bravo' returned.")
	value, err = cfg.GetInt64("charlie")
	assert.Nil(err, "No error.")
	assert.Equal(value, int64(0), "Right value 'charlie' returned.")

	// Illegal format.
	cfg.Set("illegal-format", 47.11)
	_, err = cfg.GetInt64("illegal-format")
	assert.ErrorMatch(err, `invalid type "int64" for "47.11" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetInt64("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetInt64Default("non-existing-key", 4711)
	assert.Nil(err, "No error.")
	assert.Equal(value, int64(4711), "Right value 'non-existing-key' returned.")
}

// TestUint64 tests uint64 values.
func TestUint64(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", 4711)
	cfg.Set("bravo", 0)

	// Successful gets.
	value, err := cfg.GetUint64("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, uint64(4711), "Right value 'alpha' returned.")
	value, err = cfg.GetUint64("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, uint64(0), "Right value 'bravo' returned.")

	// Illegal format.
	cfg.Set("illegal-format", -4711)
	_, err = cfg.GetUint64("illegal-format")
	assert.ErrorMatch(err, `invalid type "uint64" for "-4711" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetUint64("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetUint64Default("non-existing-key", 4711)
	assert.Nil(err, "No error.")
	assert.Equal(value, uint64(4711), "Right value 'non-existing-key' returned.")
}

// TestFloat64 tests float64 values.
func TestFloat64(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", 4711)
	cfg.Set("bravo", -47.11)
	cfg.Set("charlie", 0.0)

	// Successful gets.
	value, err := cfg.GetFloat64("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, float64(4711), "Right value 'alpha' returned.")
	value, err = cfg.GetFloat64("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, float64(-47.11), "Right value 'bravo' returned.")
	value, err = cfg.GetFloat64("charlie")
	assert.Nil(err, "No error.")
	assert.Equal(value, float64(0), "Right value 'charlie' returned.")

	// Illegal format.
	cfg.Set("illegal-format", true)
	_, err = cfg.GetFloat64("illegal-format")
	assert.ErrorMatch(err, `invalid type "float64" for "true" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetFloat64("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetFloat64Default("non-existing-key", 47.11)
	assert.Nil(err, "No error.")
	assert.Equal(value, float64(47.11), "Right value 'non-existing-key' returned.")
}

// TestTime tests time.Time values.
func TestTime(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)
	now := time.Now()
	later := now.Add(12 * time.Hour)
	earlier := now.Add(-12 * time.Hour)

	cfg.Set("alpha", now)
	cfg.Set("bravo", later)

	// Successful gets.
	value, err := cfg.GetTime("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, now, "Right value 'alpha' returned.")
	value, err = cfg.GetTime("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, later, "Right value 'bravo' returned.")

	// Illegal format.
	cfg.Set("illegal-format", true)
	_, err = cfg.GetTime("illegal-format")
	assert.ErrorMatch(err, `invalid type "time" for "true" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetTime("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetTimeDefault("non-existing-key", earlier)
	assert.Nil(err, "No error.")
	assert.Equal(value, earlier, "Right value 'non-existing-key' returned.")
}

// TestDuration tests time.Duration values.
func TestDuration(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", "0")
	cfg.Set("bravo", 5*time.Second)
	cfg.Set("charlie", "4711ms")

	// Successful gets.
	value, err := cfg.GetDuration("alpha")
	assert.Nil(err, "No error.")
	assert.Equal(value, 0*time.Second, "Right value 'alpha' returned.")
	value, err = cfg.GetDuration("bravo")
	assert.Nil(err, "No error.")
	assert.Equal(value, 5*time.Second, "Right value 'bravo' returned.")
	value, err = cfg.GetDuration("charlie")
	assert.Nil(err, "No error.")
	assert.Equal(value, 4711*time.Millisecond, "Right value 'charlie' returned.")

	// Illegal format.
	cfg.Set("illegal-format", true)
	_, err = cfg.GetDuration("illegal-format")
	assert.ErrorMatch(err, `invalid type "duration" for "true" (.*)`, "Right error returned.")

	// Non-existing key.
	_, err = cfg.GetDuration("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")

	// Non-existing key with default.
	value, err = cfg.GetDurationDefault("non-existing-key", 5*time.Hour)
	assert.Nil(err, "No error.")
	assert.Equal(value, 5*time.Hour, "Right value 'non-existing-key' returned.")
}

// TestRemove tests the removing of configuration values.
func TestRemove(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	provider := config.NewMapConfigurationProvider()
	cfg := config.New(provider)

	cfg.Set("alpha", "quick brown fox")
	cfg.Set("bravo", true)
	cfg.Set("charlie", 4711)
	cfg.Set("delta", 47.11)

	cfg.Remove("bravo")
	_, err := cfg.GetBool("bravo")
	assert.ErrorMatch(err, `key "bravo" does not exist`, "Right error returned.")

	cfg.Remove("non-existing-key")
	_, err = cfg.Get("non-existing-key")
	assert.ErrorMatch(err, `key "non-existing-key" does not exist`, "Right error returned.")
}

// EOF
