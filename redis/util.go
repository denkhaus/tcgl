// Tideland Common Go Library - Redis - Utilities
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
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
)

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
