// Tideland Common Go Library - Redis - Utilities
//
// Copyright (C) 2009-2013 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
	"time"
)

//--------------------
// ERROR TYPES
//--------------------

// ConnectionError is returned when the stile connections show an error.
type ConnectionError struct {
	Err error
}

// Error returns the error in a readable form.
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("redis: connection has a an error: %v", e.Err)
}

// IsConnectionError check if the passed error is a connection error.
func IsConnectionError(err error) bool {
	_, ok := err.(*ConnectionError)
	return ok
}

// TimeoutError is returned when Redis signals a timeout.
type TimeoutError struct {
	ElapsedTime time.Duration
}

// Error returns the error in a readable form.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("redis: timeout after %s", e.ElapsedTime)
}

// IsTimeoutError check if the passed error is a timeout error.
func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// InvalidReplyError is returned when the client recieves an
// invalid answer.
type InvalidReplyError struct {
	Length int
	Data   []byte
	Err    error
}

// Error returns the error in a readable form.
func (e *InvalidReplyError) Error() string {
	return fmt.Sprintf("redis: invalid reply (length: %d / data: %v / error: %v)", e.Length, e.Data, e.Err)
}

// IsInvalidReplyError check if the passed error is an invalid reply error.
func IsInvalidReplyError(err error) bool {
	_, ok := err.(*InvalidReplyError)
	return ok
}

// InvalidTerminationError is returned when a command terminates
// in an unspeciefied illegal way.
type InvalidTerminationError struct{}

// Error returns the error in a readable form.
func (e *InvalidTerminationError) Error() string {
	return "redis: invalid command termination"
}

// IsInvalidTerminationError check if the passed error is an invalid termmination error.
func IsInvalidTerminationError(err error) bool {
	_, ok := err.(*InvalidTerminationError)
	return ok
}

// InvalidTypeError is returned when the client recieves an
// invalid answer.
type InvalidTypeError struct {
	ExpectedType string
	Value        string
	Err          error
}

// Error returns the error in a readable form.
func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("redis: invalid type %q for %q (%v)", e.ExpectedType, e.Value, e.Err)
}

// IsInvalidTypeError check if the passed error is an invalid type error.
func IsInvalidTypeError(err error) bool {
	_, ok := err.(*InvalidTypeError)
	return ok
}

// InvalidKeyError is returned when a key or hash field
// doesn't exist.
type InvalidKeyError struct {
	Key string
}

// Error returns the error in a readable form.
func (e *InvalidKeyError) Error() string {
	return fmt.Sprintf("redis: invalid key %q", e.Key)
}

// IsInvalidKeyError check if the passed error is an invalid key error.
func IsInvalidKeyError(err error) bool {
	_, ok := err.(*InvalidKeyError)
	return ok
}

// InvalidIndexError is returned when an illegal index for addressing is used.
type InvalidIndexError struct {
	Length int
	Index  int
}

// Error returns the error in a readable form.
func (e *InvalidIndexError) Error() string {
	return fmt.Sprintf("redis: invalid index %d, length is %d", e.Index, e.Length)
}

// IsInvalidIndexError check if the passed error is an invalid index error.
func IsInvalidIndexError(err error) bool {
	_, ok := err.(*InvalidIndexError)
	return ok
}

// DatabaseClosedError is returned when Redis is used in a closed state.
type DatabaseClosedError struct {
	database *Database
}

// Error returns the error in a readable form.
func (e *DatabaseClosedError) Error() string {
	return fmt.Sprintf("redis: database %q is closed", e.database.configuration)
}

// IsDatabaseClosedError check if the passed error is a database closed error.
func IsDatabaseClosedError(err error) bool {
	_, ok := err.(*DatabaseClosedError)
	return ok
}

//--------------------
// INTERFACES
//--------------------

// Hashable represents types for Redis hashes.
type Hashable interface {
	GetHash() Hash
	SetHash(h Hash)
}

//--------------------
// PACKING
//--------------------

// Pack a value for use with lists or sets.
func Pack(v interface{}) []byte {
	vb := valueToBytes(v)
	pb := make([]byte, len(vb)+2)
	pb = append(pb, '[')
	pb = append(pb, vb...)
	pb = append(pb, ']')
	return pb
}

//--------------------
// USEFUL HELPERS
//--------------------

// valueToBytes converts a value into a byte slice.
func valueToBytes(v interface{}) []byte {
	var bs []byte
	switch vt := v.(type) {
	case string:
		bs = []byte(vt)
	case []byte:
		bs = vt
	case []string:
		bs = []byte(strings.Join(vt, "\r\n"))
	case map[string]string:
		tmp := make([]string, len(vt))
		i := 0
		for vtk, vtv := range vt {
			tmp[i] = fmt.Sprintf("%v:%v", vtk, vtv)
			i++
		}
		bs = []byte(strings.Join(tmp, "\r\n"))
	default:
		bs = []byte(fmt.Sprintf("%v", vt))
	}
	return bs
}

// argsToInterfaces converts different argument values into a slice of interfaces.
func argsToInterfaces(args ...interface{}) []interface{} {
	is := make([]interface{}, 0)
	for _, a := range args {
		// Switch based on the argument types.
		switch ta := a.(type) {
		case []string:
			for _, s := range ta {
				is = append(is, s)
			}
		default:
			is = append(is, ta)
		}
	}
	return is
}

// EOF
