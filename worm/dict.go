// Tideland Common Go Library - Write once read multiple - Dict
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
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
)

//--------------------
// DICT
//--------------------

// DictValues is a map for data exchange with a dict.
type DictValues map[string]interface{}

// Dict stores keys and values.
type Dict struct {
	values DictValues
}

// NewDict creates a new dictionary.
func NewDict(values DictValues) (Dict, error) {
	d := Dict{make(DictValues)}
	if values != nil {
		for key, value := range values {
			switch v := value.(type) {
			case string, bool, int, int64, uint64, float64, complex128:
				d.values[key] = v
			case []byte:
				d.values[key] = duplicate(v)
			case IntSet, StringSet:
				d.values[key] = v
			default:
				buf := new(bytes.Buffer)
				enc := gob.NewEncoder(buf)
				err := enc.Encode(value)
				if err != nil {
					// Return empty dictionary with error.
					return Dict{make(DictValues)}, err
				}
				d.values[key] = buf.Bytes()
			}
		}
	}
	return d, nil
}

// Len returns the number of values in the dictionary.
func (d Dict) Len() int {
	return len(d.values)
}

// Keys returns the keys of the dictionary.
func (d Dict) Keys() []string {
	keys := make([]string, len(d.values))
	ptr := 0
	for key := range d.values {
		keys[ptr] = key
		ptr++
	}
	sort.Strings(keys)
	return keys
}

// ContainsKeys tests if all the passed keys are in the dictionary.
func (d Dict) ContainsKeys(keys ...string) bool {
	for _, key := range keys {
		if _, ok := d.values[key]; !ok {
			return false
		}
	}
	return true
}

// Copy create a new dictionary and adds the values of the keys to it.
func (d Dict) Copy(keys ...string) Dict {
	nv := make(DictValues)
	for _, key := range keys {
		if value, ok := d.values[key]; ok {
			nv[key] = value
		}
	}
	nd, _ := NewDict(nv)
	return nd
}

// CopyAll creates a new dictionary and adds all values to it.
func (d Dict) CopyAll() Dict {
	nd, _ := NewDict(d.values)
	return nd
}

// Apply creates a new dictionary with all passed values and those
// of this dictionary which are not in the values.
func (d Dict) Apply(values DictValues) (Dict, error) {
	nd, err := NewDict(values)
	if err != nil {
		return nd, err
	}
	for key, value := range d.values {
		if nd.values[key] == nil {
			nd.values[key] = value
		}
	}
	return nd, nil
}

// Read reads the value of a key into value, types have to match.
func (d Dict) Read(key string, value interface{}) (err error) {
	var b []byte
	if b, err = d.Bytes(key); err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	return dec.Decode(value)
}

// Bytes returns the value of a key as byte slice.
func (d Dict) Bytes(key string) ([]byte, error) {
	if v, ok := d.values[key]; ok {
		if bs, ok := v.([]byte); ok {
			return duplicate(bs), nil
		}
		return nil, &InvalidTypeError{key, "[]byte"}
	}
	return nil, &InvalidKeyError{key}
}

// String returns the value of a key as string.
func (d Dict) String(key string) (string, error) {
	if v, ok := d.values[key]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
		return "", &InvalidTypeError{key, "string"}
	}
	return "", &InvalidKeyError{key}
}

// String returns the value/representation of a key as string, at least 
// the default value.
func (d Dict) StringDefault(key, def string) string {
	if v, ok := d.values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
}

// Bool returns the value of a key as bool.
func (d Dict) Bool(key string) (bool, error) {
	if v, ok := d.values[key]; ok {
		if b, ok := v.(bool); ok {
			return b, nil
		}
		return false, &InvalidTypeError{key, "bool"}
	}
	return false, &InvalidKeyError{key}
}

// Int returns the value of a key as int.
func (d Dict) Int(key string) (int, error) {
	if v, ok := d.values[key]; ok {
		if i, ok := v.(int); ok {
			return i, nil
		}
		return 0, &InvalidTypeError{key, "int"}
	}
	return 0, &InvalidKeyError{key}
}

// Int64 returns the value of a key as int64.
func (d Dict) Int64(key string) (int64, error) {
	if v, ok := d.values[key]; ok {
		if i, ok := v.(int64); ok {
			return i, nil
		}
		return 0, &InvalidTypeError{key, "int64"}
	}
	return 0, &InvalidKeyError{key}
}

// Uint64 returns the value of a key as uint64.
func (d Dict) Uint64(key string) (uint64, error) {
	if v, ok := d.values[key]; ok {
		if ui, ok := v.(uint64); ok {
			return ui, nil
		}
		return 0, &InvalidTypeError{key, "uint64"}
	}
	return 0, &InvalidKeyError{key}
}

// Float64 returns the value of a key as float64.
func (d Dict) Float64(key string) (float64, error) {
	if v, ok := d.values[key]; ok {
		if f, ok := v.(float64); ok {
			return f, nil
		}
		return 0.0, &InvalidTypeError{key, "float64"}
	}
	return 0.0, &InvalidKeyError{key}
}

// Complex128 returns the value of a key as complex128.
func (d Dict) Complex128(key string) (complex128, error) {
	if v, ok := d.values[key]; ok {
		if c, ok := v.(complex128); ok {
			return c, nil
		}
		return complex(0, 0), &InvalidTypeError{key, "complex128"}
	}
	return complex(0, 0), &InvalidKeyError{key}
}

// IntSet returns the value of a key as set of integers.
func (d Dict) IntSet(key string) (IntSet, error) {
	if v, ok := d.values[key]; ok {
		if i, ok := v.(IntSet); ok {
			return i, nil
		}
		return NewIntSet(Ints{}), &InvalidTypeError{key, "IntSet"}
	}
	return NewIntSet(Ints{}), &InvalidKeyError{key}
}

// StringSet returns the value of a key as set of strings.
func (d Dict) StringSet(key string) (StringSet, error) {
	if v, ok := d.values[key]; ok {
		if s, ok := v.(StringSet); ok {
			return s, nil
		}
		return NewStringSet(Strings{}), &InvalidTypeError{key, "IntSet"}
	}
	return NewStringSet(Strings{}), &InvalidKeyError{key}
}

//--------------------
// HELPERS
//--------------------

// duplicate creates a duplicate of the passed byte slice.
func duplicate(bs []byte) []byte {
	cbs := make([]byte, len(bs))
	copy(cbs, bs)
	return cbs
}

// EOF
