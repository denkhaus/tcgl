// Tideland Common Go Library - Identifier - Unit Tests - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package identifier

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/asserts"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test the UUID.
func TestUuid(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Asserts.
	uuid := NewUUID()
	uuidStr := uuid.String()
	assert.Equal(len(uuid), 16, "UUID length has to be 16.")
	assert.Match(uuidStr, "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "UUID has to match.")
	uuids := make(map[string]bool)
	for i := 0; i < 1000000; i++ {
		uuid = NewUUID()
		uuidStr = uuid.String()
		assert.False(uuids[uuidStr], "UUID collision should not happen.")
		uuids[uuidStr] = true
	}
}

// Test the creation of identifiers based on types.
func TestTypeAsIdentifierPart(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Type as identifier.
	var tai TypeToSplitForIdentifier

	id := TypeAsIdentifierPart(tai)
	assert.Equal(id, "type-to-split-for-identifier", "Wrong TypeAsIdentifierPart() result!")

	id = TypeAsIdentifierPart(NewUUID())
	assert.Equal(id, "u-u-i-d", "Wrong TypeAsIdentifierPart() result!")
}

// Test the creation of identifiers based on parts.
func TestIdentifier(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Identifier.
	id := Identifier("One", 2, "three four")
	assert.Equal(id, "one:2:three-four", "Wrong Identifier() result!")

	id = Identifier(2011, 6, 22, "One, two, or  three things.")
	assert.Equal(id, "2011:6:22:one-two-or-three-things", "Wrong Identifier() result!")
}

// Test the creation of identifiers based on parts with defined seperators.
func TestSepIdentifier(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	id := SepIdentifier("+", 1, "oNe", 2, "TWO", "3", "ÄÖÜ")
	assert.Equal(id, "1+one+2+two+3+äöü", "Wrong SepIdentifier() result!")

	id = LimitedSepIdentifier("+", true, "     ", 1, "oNe", 2, "TWO", "3", "ÄÖÜ", "Four", "+#-:,")
	assert.Equal(id, "1+one+2+two+3+four", "Wrong LimitedSepIdentifier() result!")
}

//--------------------
// HELPER
//--------------------

// Type as part of an identifier.
type TypeToSplitForIdentifier bool

// EOF
