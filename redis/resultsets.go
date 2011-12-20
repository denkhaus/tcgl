// Tideland Common Go Library - Redis - Result Sets
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

//--------------------
// VALUE
//--------------------

// Value is simply a byte slice.
type Value []byte

// String returns the value as string (alternative to type conversion).
func (v Value) String() string {
	return string([]byte(v))
}

// Bool return the value as bool.
func (v Value) Bool() bool {
	if b, err := strconv.Atob(v.String()); err == nil {
		return b
	}

	return false
}

// Int returns the value as int.
func (v Value) Int() int {
	if i, err := strconv.Atoi(v.String()); err == nil {
		return i
	}

	return 0
}

// Int64 returns the value as int64.
func (v Value) Int64() int64 {
	if i, err := strconv.Atoi64(v.String()); err == nil {
		return i
	}

	return 0
}

// Uint64 returns the value as uint64.
func (v Value) Uint64() uint64 {
	if i, err := strconv.Atoui64(v.String()); err == nil {
		return i
	}

	return 0
}

// Float64 returns the value as float64.
func (v Value) Float64() float64 {
	if f, err := strconv.Atof64(v.String()); err == nil {
		return f
	}

	return 0.0
}

// Bytes returns the value as byte slice.
func (v Value) Bytes() []byte {
	return []byte(v)
}

// StringSlice returns the value as slice of strings when seperated by CRLF.
func (v Value) StringSlice() []string {
	return strings.Split(v.String(), "\r\n")
}

// StringMap returns the value as a map of strings when seperated by CRLF
// and colons between key and value.
func (v Value) StringMap() map[string]string {
	tmp := v.StringSlice()
	m := make(map[string]string, len(tmp))

	for _, s := range tmp {
		kv := strings.Split(s, ":")

		if len(kv) > 1 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return m
}

// Unpack removes the braces of a list value.
func (v Value) Unpack() Value {
	if len(v) > 2 && v[0] == '[' && v[len(v)-1] == ']' {
		return Value(v[1 : len(v)-1])
	}

	return v
}

//--------------------
// SPECIAL VALUES
//--------------------

// ScoredValue contains a value with its score from a sorted set.
type ScoredValue struct {
	Value Value
	Score int
}

// KeyValue combines a key and a value for blocked lists.
type KeyValue struct {
	Key   string
	Value Value
}

//--------------------
// HASH
//--------------------

// Hash maps multiple fields of a hash to the
// according result values.
type Hash map[string]Value

// NewHash creates a new empty hash.
func NewHash() Hash {
	return make(Hash)
}

// Len returns the number of elements in the hash.
func (h Hash) Len() int {
	return len(h)
}

// Set sets a key to the given value.
func (h Hash) Set(k string, v interface{}) {
	h[k] = Value(valueToBytes(v))
}

// String returns the value of a key as string.
func (h Hash) String(k string) string {
	if v, ok := h[k]; ok {
		return v.String()
	}

	return ""
}

// Bool returns the value of a key as bool.
func (h Hash) Bool(k string) bool {
	if v, ok := h[k]; ok {
		return v.Bool()
	}

	return false
}

// Int returns the value of a key as int.
func (h Hash) Int(k string) int {
	if v, ok := h[k]; ok {
		return v.Int()
	}

	return 0
}

// Int64 returns the value of a key as int64.
func (h Hash) Int64(k string) int64 {
	if v, ok := h[k]; ok {
		return v.Int64()
	}

	return 0
}

// Uint64 returns the value of a key as uint64.
func (h Hash) Uint64(k string) uint64 {
	if v, ok := h[k]; ok {
		return v.Uint64()
	}

	return 0
}

// Float64 returns the value of a key as float64.
func (h Hash) Float64(k string) float64 {
	if v, ok := h[k]; ok {
		return v.Float64()
	}

	return 0.0
}

// Bytes returns the value of a key as byte slice.
func (h Hash) Bytes(k string) []byte {
	if v, ok := h[k]; ok {
		return v.Bytes()
	}

	return []byte{}
}

// StringSlice returns the value of a key as string slice.
func (h Hash) StringSlice(k string) []string {
	if v, ok := h[k]; ok {
		return v.StringSlice()
	}

	return []string{}
}

// StringMap returns the value of a key as string map.
func (h Hash) StringMap(k string) map[string]string {
	if v, ok := h[k]; ok {
		return v.StringMap()
	}

	return map[string]string{}
}

//--------------------
// RESULT SET
//--------------------

// ResultSet is the returned struct of commands.
type ResultSet struct {
	cmd        string
	values     []Value
	resultSets []*ResultSet
	error      os.Error
}

// newResultSet creates a result set.
func newResultSet(cmd string) *ResultSet {
	return &ResultSet{
		cmd:   cmd,
		error: os.NewError("illegal terminated command"),
	}
}

// IsOK return true if the result is ok.
func (rs *ResultSet) IsOK() bool {
	if rs.error != nil {
		return false
	}

	return true
}

// IsMulti returns true if the result set contains
// multiple result sets.
func (rs *ResultSet) IsMulti() bool {
	return rs.resultSets != nil
}

// Command returns the executed command.
func (rs *ResultSet) Command() string {
	return rs.cmd
}

// ValueCount returns the number of returned values.
func (rs *ResultSet) ValueCount() int {
	if rs.values == nil {
		return 0
	}

	return len(rs.values)
}

// ValueAt returns a wanted value by its index.
func (rs *ResultSet) ValueAt(idx int) Value {
	if idx < 0 || idx >= len(rs.values) {
		return Value([]byte{})
	}

	return rs.values[idx]
}

// Value returns the first value.
func (rs *ResultSet) Value() Value {
	return rs.ValueAt(0)
}

//UnpackedValue returns the first value unpacked.
func (rs *ResultSet) UnpackedValue() Value {
	return rs.ValueAt(0).Unpack()
}

// Values returns all values as slice.
func (rs *ResultSet) Values() []Value {
	if rs.values == nil {
		return nil
	}

	vs := make([]Value, len(rs.values))

	copy(vs, rs.values)

	return vs
}

// UnpackedValues returns all values unpacked as slice.
func (rs *ResultSet) UnpackedValues() []Value {
	vs := rs.Values()

	for i, v := range vs {
		vs[i] = v.Unpack()
	}

	return vs
}

// ValueAsInt returns the first value as bool.
func (rs *ResultSet) ValueAsBool() bool {
	return rs.Value().Bool()
}

// ValueAsInt returns the first value as int.
func (rs *ResultSet) ValueAsInt() int {
	return rs.Value().Int()
}

// ValueAsString returns the first value as string.
func (rs *ResultSet) ValueAsString() string {
	return rs.Value().String()
}

// KeyValue return the first value as key and the second as value.
func (rs *ResultSet) KeyValue() *KeyValue {
	return &KeyValue{
		Key:   rs.ValueAt(0).String(),
		Value: rs.ValueAt(1),
	}
}

// ValuesDo iterates over the result values and
// performs the passed function for each one.
func (rs *ResultSet) ValuesDo(f func(int, Value)) {
	for idx, v := range rs.values {
		f(idx, v)
	}
}

// ValuesMap iterates over the result values and
// performs the passed function for each one. The result
// is a slice of values returned by the functions.
func (rs *ResultSet) ValuesMap(f func(Value) interface{}) []interface{} {
	result := make([]interface{}, len(rs.values))

	for idx, v := range rs.values {
		result[idx] = f(v)
	}

	return result
}

// Hash returns the values of the result set as hash.
func (rs *ResultSet) Hash() Hash {
	var key string

	result := make(Hash)
	set := false

	for _, v := range rs.values {
		if set {
			// Write every second value.
			result.Set(key, v.Bytes())

			set = false
		} else {
			// First value is always a key.
			key = v.String()

			set = true
		}
	}

	return result
}

// SetHashable takes the values of the result set as hash
// and sets the passed hashable.
func (rs *ResultSet) SetHashable(h Hashable) {
	h.SetHash(rs.Hash())
}

// ResultSetCount returns the number of result sets
// inside the result set.
func (rs *ResultSet) ResultSetCount() int {
	if rs.resultSets == nil {
		return 0
	}

	return len(rs.resultSets)
}

// ResultSetAt returns a result set by its index.
func (rs *ResultSet) ResultSetAt(idx int) *ResultSet {
	if idx < 0 || idx >= len(rs.resultSets) {
		rs := newResultSet("none")

		rs.error = os.NewError("illegal result set index")

		return rs
	}

	return rs.resultSets[idx]
}

// ResultSetsDo iterates over the result sets and
// performs the passed function for each one.
func (rs *ResultSet) ResultSetsDo(f func(*ResultSet)) {
	for _, rs := range rs.resultSets {
		f(rs)
	}
}

// ResultSetsMap iterates over the result sets and
// performs the passed function for each one. The result
// is a slice of values returned by the functions.
func (rs *ResultSet) ResultSetsMap(f func(*ResultSet) interface{}) []interface{} {
	result := make([]interface{}, len(rs.resultSets))

	for idx, rs := range rs.resultSets {
		result[idx] = f(rs)
	}

	return result
}

// Error returns the error if the operation creating
// the result set failed.
func (rs *ResultSet) Error() os.Error {
	return rs.error
}

// String returns the result set as a string.
func (rs *ResultSet) String() string {
	r := fmt.Sprintf("C(%v) V(%v) E(%v)", rs.cmd, rs.values, rs.error)

	if rs.IsMulti() {
		rs.ResultSetsDo(func(each *ResultSet) {
			r += "\n- " + each.String()
		})
	}

	return r
}

//--------------------
// RESULT SET FUTURE
//--------------------

// Future just waits for a result set
// returned somewhere in the future.
type Future struct {
	rsChan chan *ResultSet
}

// newFuture creates the new future.
func newFuture() *Future {
	return &Future{make(chan *ResultSet, 1)}
}

// setResultSet sets the result set.
func (f *Future) setResultSet(rs *ResultSet) {
	f.rsChan <- rs
}

// ResultSet returns the result set in the moment it is available.
func (f *Future) ResultSet() (rs *ResultSet) {
	rs = <-f.rsChan

	f.rsChan <- rs

	return
}

// EOF