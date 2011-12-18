// Tideland Common Go Library - Identifier - Unit Tests - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package identifier

//--------------------
// IMPORTS
//--------------------

import (
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test the UUID.
func TestUuid(t *testing.T) {
	uuids := make(map[string]bool)

	t.Logf("Start generating UUIDs ...")

	for i := 0; i < 1000000; i++ {
		uuid := NewUUID().String()

		if uuids[uuid] {
			t.Fatalf("UUID collision")
		}

		uuids[uuid] = true
	}

	t.Logf("Done generating UUIDs!")
}

// Test the creation of an identifier.
func TestIdentifier(t *testing.T) {
	// Type as identifier.
	var tai TypeToSplitForIdentifier

	idp := TypeAsIdentifierPart(tai)

	if idp != "type-to-split-for-identifier" {
		t.Errorf("Identifier part for TypeTpSplitForIdentifier is wrong, returned '%v'!", idp)
	}

	idp = TypeAsIdentifierPart(NewUUID())

	if idp != "u-u-i-d" {
		t.Errorf("Identifier part for UUID is wrong, returned '%v'!", idp)
	}

	// Identifier.
	id := Identifier("One", 2, "three four")

	if id != "one:2:three-four" {
		t.Errorf("First identifier is wrong! Id: %v", id)
	}

	id = Identifier(2011, 6, 22, "One, two, or  three things.")

	if id != "2011:6:22:one-two-or-three-things" {
		t.Errorf("Second identifier is wrong! Id: %v", id)
	}

	id = SepIdentifier("+", 1, "oNe", 2, "TWO", "3", "ÄÖÜ")

	if id != "1+one+2+two+3+äöü" {
		t.Errorf("Third identifier is wrong! Id: %v", id)
	}

	id = LimitedSepIdentifier("+", true, "     ", 1, "oNe", 2, "TWO", "3", "ÄÖÜ", "Four", "+#-:,")

	if id != "1+one+2+two+3+four" {
		t.Errorf("Fourth identifier is wrong! Id: %v", id)
	}
}

//--------------------
// HELPER
//--------------------

// Type as part of an identifier.
type TypeToSplitForIdentifier bool

// EOF
