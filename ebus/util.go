// Tideland Common Go Library - Event Bus - Utilities
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
	"cgl.tideland.biz/identifier"
)

//--------------------
// IDENTIFIER
//--------------------

// Id is used to identify context values and cells.
type Id string

// NewId generates an identifier based on the given parts.
func NewId(parts ...interface{}) Id {
	return Id(identifier.Identifier(parts...))
}

// EOF
