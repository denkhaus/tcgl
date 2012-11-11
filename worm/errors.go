// Tideland Common Go Library - Write once read multiple - Errors
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package worm

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
)

//--------------------
// ERRORS
//--------------------

// InvalidKeyError is returned when a key in a dictionary doesn't exist.
type InvalidKeyError struct {
	Key string
}

// Error returns the error in a readable form.
func (e *InvalidKeyError) Error() string {
	return fmt.Sprintf("invalid key %q for the dictionary", e.Key)
}

// IsInvalidKeyError tests the error type.
func IsInvalidKeyError(err error) bool {
	_, ok := err.(*InvalidKeyError)
	return ok
}

// InvalidTypeError is returned when a retrieved value has
// an invalid type.
type InvalidTypeError struct {
	Key          string
	ExpectedType string
}

// Error returns the error in a readable form.
func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type %q expected for key %q", e.ExpectedType, e.Key)
}

// IsInvalidTypeError tests the error type.
func IsInvalidTypeError(err error) bool {
	_, ok := err.(*InvalidTypeError)
	return ok
}

// EOF
