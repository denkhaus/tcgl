// Tideland Common Go Library - Map/Reduce
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package mapreduce

//--------------------
// IMPORTS
//--------------------

import (
	"hash/adler32"
	"cgl.tideland.biz/sort"
)

//--------------------
// KEY/VALUE TYPES
//--------------------

// KeyValue is a pair of a string key and any data as value.
type KeyValue struct {
	Key   string
	Value interface{}
}

// KeyValueChan is a channel for the transfer of key/value pairs.
type KeyValueChan chan *KeyValue

// KeyValueChans is a slice of key/value channels.
type KeyValueChans []KeyValueChan

// KeyValueLessFunc is the less function type needed for sorting.
type KeyValueLessFunc func(*KeyValue, *KeyValue) bool

// KeyValues is an ordered slice of keys and values, keys can occur multiple times.
type KeyValues struct {
	Data     []*KeyValue
	LessFunc KeyValueLessFunc
}

// NewKeyValues create a new empty key/value slice with size 0, capacity c and
// a less func for sorting.
func NewKeyValues(c int, kvlf KeyValueLessFunc) *KeyValues {
	return &KeyValues{make([]*KeyValue, 0, c), kvlf}
}

// Add adds a key/value to the slice.
func (kv *KeyValues) Add(key string, value interface{}) {
	kv.Data = append(kv.Data, &KeyValue{key, value})
}

// Append appends the data of other key/values to this key/values.
func (kv *KeyValues) Append(akv KeyValues) {
	kv.Data = append(kv.Data, akv.Data...)
}

// AppendChan appends the key/values of a chan to this key/values.
func (kv *KeyValues) AppendChan(kvChan KeyValueChan) {
	for dkv := range kvChan {
		l := len(kv.Data)

		if l == cap(kv.Data) {
			tmp := make([]*KeyValue, l, l+1024)

			copy(tmp, kv.Data)

			kv.Data = tmp
		}

		kv.Data = kv.Data[0 : l+1]
		kv.Data[l] = dkv
	}
}

// KeyValueChan returns a key/value channel and feeds it with the data.
func (kv *KeyValues) KeyValueChan() KeyValueChan {
	kvChan := make(KeyValueChan)

	go func() {
		for _, dkv := range kv.Data {
			kvChan <- dkv
		}

		close(kvChan)
	}()

	return kvChan
}

// Len returns the number of keys and values.
func (kv *KeyValues) Len() int {
	return len(kv.Data)
}

// Less returns wether the key with index i should sort before the
// key with index j.
func (kv *KeyValues) Less(i, j int) bool {
	return kv.LessFunc(kv.Data[i], kv.Data[j])
}

// Swap swaps the key/value pairs with indexes i and j.
func (kv *KeyValues) Swap(i, j int) {
	kv.Data[i], kv.Data[j] = kv.Data[j], kv.Data[i]
}

// Sort sorts the key/values based on the less func.
func (kv *KeyValues) Sort() {
	sort.Sort(kv)
}

// KeyLessFunc compares the keys of two key/value
// pairs. It returns true if the key of a is less
// the key of b.
func KeyLessFunc(a *KeyValue, b *KeyValue) bool {
	return a.Key < b.Key
}

//--------------------
// MAP/REDUCE
//--------------------

// Map a key/value pair, emit to the channel.
type MapFunc func(*KeyValue, KeyValueChan)

// Reduce the key/values of the first channel, emit to the second channel.
type ReduceFunc func(KeyValueChan, KeyValueChan)

// Channel for closing signals.
type SigChan chan bool

// Close given channel after a number of signals.
func closeSignalChannel(kvc KeyValueChan, size int) SigChan {
	sigChan := make(SigChan)

	go func() {
		ctr := 0

		for {
			<-sigChan

			ctr++

			if ctr == size {
				close(kvc)

				return
			}
		}
	}()

	return sigChan
}

// Perform the reducing.
func performReducing(mapEmitChan KeyValueChan, reduceFunc ReduceFunc, reduceSize int, reduceEmitChan KeyValueChan) {
	// Start a closer for the reduce emit chan.

	sigChan := closeSignalChannel(reduceEmitChan, reduceSize)

	// Start reduce funcs.

	reduceChans := make(KeyValueChans, reduceSize)

	for i := 0; i < reduceSize; i++ {
		reduceChans[i] = make(KeyValueChan)

		go func(inChan KeyValueChan) {
			reduceFunc(inChan, reduceEmitChan)

			sigChan <- true
		}(reduceChans[i])
	}

	// Read map emitted data.

	for kv := range mapEmitChan {
		hash := adler32.Checksum([]byte(kv.Key))
		idx := hash % uint32(reduceSize)

		reduceChans[idx] <- kv
	}

	// Close reduce channels.

	for _, reduceChan := range reduceChans {
		close(reduceChan)
	}
}

// Perform the mapping.
func performMapping(mapInChan KeyValueChan, mapFunc MapFunc, mapSize int, mapEmitChan KeyValueChan) {
	// Start a closer for the map emit chan.

	sigChan := closeSignalChannel(mapEmitChan, mapSize)

	// Start mapping goroutines.

	mapChans := make(KeyValueChans, mapSize)

	for i := 0; i < mapSize; i++ {
		mapChans[i] = make(KeyValueChan)

		go func(inChan KeyValueChan) {
			for kv := range inChan {
				mapFunc(kv, mapEmitChan)
			}

			sigChan <- true
		}(mapChans[i])
	}

	// Dispatch input data to map channels.

	idx := 0

	for kv := range mapInChan {
		mapChans[idx%mapSize] <- kv

		idx++
	}

	// Close mapping channels channel.

	for i := 0; i < mapSize; i++ {
		close(mapChans[i])
	}
}

// MapReduce applies a map and a reduce function to keys and values in parallel.
func MapReduce(inChan KeyValueChan, mapFunc MapFunc, mapSize int, reduceFunc ReduceFunc, reduceSize int) KeyValueChan {
	mapEmitChan := make(KeyValueChan)
	reduceEmitChan := make(KeyValueChan)

	// Perform operations.

	go performReducing(mapEmitChan, reduceFunc, reduceSize, reduceEmitChan)
	go performMapping(inChan, mapFunc, mapSize, mapEmitChan)

	return reduceEmitChan
}

// SortedMapReduce performes a map/reduce and sorts the result.
func SortedMapReduce(inChan KeyValueChan, mapFunc MapFunc, mapSize int, reduceFunc ReduceFunc, reduceSize int, lessFunc KeyValueLessFunc) KeyValueChan {
	kvChan := MapReduce(inChan, mapFunc, mapSize, reduceFunc, reduceSize)
	kv := NewKeyValues(1024, lessFunc)

	kv.AppendChan(kvChan)
	kv.Sort()

	return kv.KeyValueChan()
}

// EOF
