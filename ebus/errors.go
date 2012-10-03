// Tideland Common Go Library - Event Bus - Errors
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
)

//--------------------
// ERRORS
//--------------------

// ContextValueNotFoundError is returned when a requested
// context value is not available.
type ContextValueNotFoundError struct {
	Key Id
}

func (e *ContextValueNotFoundError) Error() string {
	return fmt.Errorf("context value for %q cannot be found", e.Key)
}

// EOF
