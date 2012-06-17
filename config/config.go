// Tideland Common Go Library - Configuration
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package config

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

//--------------------
// CONST
//--------------------

// timeFormat is used for the conversion between time and string.
const timeFormat = "2006-01-02T15:04:05.000000000Z07:00"

//--------------------
// CONFIGURATION PROVIDER
//--------------------

// ConfigurationProvider defines the interface to the backend
// of a configuration.
type ConfigurationProvider interface {
	// Get retrieves a raw value from the configuration provider.
	Get(key string) (value string, err error)
	// Set stores a raw value at the configuration provider.
	Set(key, value string) (old string, ok bool, err error)
	// Remove deletes a key from the configuration provider.
	Remove(key string) error
}

// NewMapConfigurationProvider creates a new configuration provider
// based on a simple Go map.
func NewMapConfigurationProvider() ConfigurationProvider {
	return &MapConfigurationProvider{
		data: make(map[string]string),
	}
}

// MapConfigurationProvider stores the configuration
// values in a simple map.
type MapConfigurationProvider struct {
	mutex sync.RWMutex
	data  map[string]string
}

// Get retrieves a raw value from the configuration provider.
func (p *MapConfigurationProvider) Get(key string) (value string, err error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	value, ok := p.data[key]
	if !ok {
		return "", InvalidKeyError{key}
	}
	return value, nil
}

// Set stores a value at the provider and returns an old value if exists.
func (p *MapConfigurationProvider) Set(key, value string) (old string, ok bool, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	old, ok = p.data[key]
	p.data[key] = value
	return old, ok, nil
}

// Remove deletes a key from the configuration provider.
func (p *MapConfigurationProvider) Remove(key string) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	delete(p.data, key)
	return nil
}

//--------------------
// CONFIGURATION
//--------------------

// Configuration maps keys to values for configuration purposes.
type Configuration struct {
	provider ConfigurationProvider
}

// New returns a new empty configuration.
func New(provider ConfigurationProvider) *Configuration {
	return &Configuration{provider}
}

// SetFromMap sets the configuration with map data.
func (c *Configuration) SetFromMap(m map[string]interface{}) error {
	for key, value := range m {
		_, err := c.Set(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetFromSlice sets the configuration with slice data. 
// It contains alternating keys and values.
func (c *Configuration) SetFromSlice(s []string) error {
	key := ""
	for i := 0; i < len(s); i++ {
		if i%2 == 0 {
			key = s[i]
		} else {
			_, err := c.Set(key, s[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Set sets a value in the configuration.
func (c *Configuration) Set(key string, value interface{}) (old string, err error) {
	var sv string
	switch v := value.(type) {
	case string:
		sv = v
	case time.Time:
		sv = v.Format(timeFormat)
	case fmt.Stringer:
		sv = v.String()
	default:
		sv = fmt.Sprintf("%v", v)
	}
	old, _, err = c.provider.Set(key, sv)
	return old, err
}

// Get returns a string value without type conversion. 
func (c *Configuration) Get(key string) (string, error) {
	return c.provider.Get(key)
}

// GetDefault returns a string value without type conversion,
// if key doesn't exist the default.
func (c *Configuration) GetDefault(key, d string) (string, error) {
	value, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return "", err
	}
	return value, nil
}

// GetBool returns a value as bool.
func (c *Configuration) GetBool(key string) (bool, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, InvalidTypeError{"bool", raw, err}
	}
	return b, nil
}

// GetBoolDefault returns a value as bool, if key doesn't exist the default.
func (c *Configuration) GetBoolDefault(key string, d bool) (bool, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return false, err
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, InvalidTypeError{"bool", raw, err}
	}
	return b, nil
}

// GetInt returns a value as int.
func (c *Configuration) GetInt(key string) (int, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(raw)
	if err != nil {
		return 0, InvalidTypeError{"int", raw, err}
	}
	return i, nil
}

// GetIntDefault returns a value as int, if key doesn't exist the default.
func (c *Configuration) GetIntDefault(key string, d int) (int, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return 0, err
	}
	i, err := strconv.Atoi(raw)
	if err != nil {
		return 0, InvalidTypeError{"int", raw, err}
	}
	return i, nil
}

// GetInt64 returns a value as int64.
func (c *Configuration) GetInt64(key string) (int64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return 0, err
	}
	i64, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, InvalidTypeError{"int64", raw, err}
	}
	return i64, nil
}

// GetInt64Default returns a value as int64, if key doesn't exist the default.
func (c *Configuration) GetInt64Default(key string, d int64) (int64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return 0, err
	}
	i64, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, InvalidTypeError{"int64", raw, err}
	}
	return i64, nil
}

// GetUint64 returns a value as uint64.
func (c *Configuration) GetUint64(key string) (uint64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return 0, err
	}
	ui64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, InvalidTypeError{"uint64", raw, err}
	}
	return ui64, nil
}

// GetUint64Default returns a value as uint64, if key doesn't exist the default.
func (c *Configuration) GetUint64Default(key string, d uint64) (uint64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return 0, err
	}
	ui64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, InvalidTypeError{"uint64", raw, err}
	}
	return ui64, nil
}

// GetFloat64 returns a value as float64.
func (c *Configuration) GetFloat64(key string) (float64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return 0.0, err
	}
	f64, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0.0, InvalidTypeError{"float64", raw, err}
	}
	return f64, nil
}

// GetFloat64Default returns a value as float64, if key doesn't exist the default.
func (c *Configuration) GetFloat64Default(key string, d float64) (float64, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return 0.0, err
	}
	f64, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0.0, InvalidTypeError{"float64", raw, err}
	}
	return f64, nil
}

// GetTime returns a value as time.
func (c *Configuration) GetTime(key string) (time.Time, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(timeFormat, raw)
	if err != nil {
		return time.Time{}, InvalidTypeError{"time", raw, err}
	}
	return t, nil
}

// GetTimeDefault returns a value as time, if key doesn't exist the default.
func (c *Configuration) GetTimeDefault(key string, d time.Time) (time.Time, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return time.Time{}, err
	}
	t, err := time.Parse(timeFormat, raw)
	if err != nil {
		return time.Time{}, InvalidTypeError{"time", raw, err}
	}
	return t, nil
}

// GetDuration returns a value as duration.
func (c *Configuration) GetDuration(key string) (time.Duration, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, InvalidTypeError{"duration", raw, err}
	}
	return d, nil
}

// GetDurationDefault returns a value as duration, if key doesn't exist the default.
func (c *Configuration) GetDurationDefault(key string, d time.Duration) (time.Duration, error) {
	raw, err := c.provider.Get(key)
	if err != nil {
		if IsInvalidKeyError(err) {
			return d, nil
		}
		return 0, err
	}
	td, err := time.ParseDuration(raw)
	if err != nil {
		return 0, InvalidTypeError{"duration", raw, err}
	}
	return td, nil
}

// Remove deletes a key.
func (c *Configuration) Remove(key string) error {
	return c.provider.Remove(key)
}

//--------------------
// ERRORS
//--------------------

// InvalidKeyError is returned if the key is invalid.
type InvalidKeyError struct {
	Key string
}

// Error returns the error in a readable form.
func (e InvalidKeyError) Error() string {
	return fmt.Sprintf("key %q does not exist", e.Key)
}

// IsInvalidKeyError checks if an error is an invalid key error.
func IsInvalidKeyError(err error) bool {
	_, ok := err.(InvalidKeyError)
	return ok
}

// Invalid type error is returned if the data can't be converted
// to the expected type.
type InvalidTypeError struct {
	ExpectedType string
	Value        string
	Err          error
}

// Error returns the error in a readable form.
func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type %q for %q (%v)", e.ExpectedType, e.Value, e.Err)
}

// IsInvalidTypeError checks if an error is an invalid type error.
func IsInvalidTypeError(err error) bool {
	_, ok := err.(InvalidTypeError)
	return ok
}

// EOF
