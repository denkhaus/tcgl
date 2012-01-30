// Tideland Common Go Library - Simple Markup Language - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package markup

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"strings"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test creating.
func TestSmlCreating(t *testing.T) {
	root := createSmlStructure()

	t.Logf("Root: %v", root)
}

// Test SML writer processing.
func TestSmlWriterProcessing(t *testing.T) {
	root := createSmlStructure()
	bufA := bytes.NewBufferString("")
	bufB := bytes.NewBufferString("")
	sppA := NewSmlWriterProcessor(bufA, true)
	sppB := NewSmlWriterProcessor(bufB, false)

	root.ProcessWith(sppA)
	root.ProcessWith(sppB)

	t.Logf("Print A: %v", bufA)
	t.Logf("Print B: %v", bufB)
}

// Test positive reading.
func TestSmlPositiveReading(t *testing.T) {
	sml := "Before!   {foo {bar:1:first Yadda ^{Test^} 1}  {inbetween}  {bar:2:last Yadda {Test ^^} 2}}   After!"
	reader := NewSmlReader(strings.NewReader(sml))

	root, err := reader.RootTagNode()

	if err == nil {
		t.Logf("Root:%v", root)
	} else {
		t.Errorf("Error: %v", err)
	}
}

// Test negative reading.
func TestSmlNegativeReading(t *testing.T) {
	sml := "{Foo {bar:1 Yadda {test} {} 1} {bar:2 Yadda 2}}"
	reader := NewSmlReader(strings.NewReader(sml))

	root, err := reader.RootTagNode()

	if err == nil {
		t.Errorf("Root: %v", root)
	} else {
		t.Logf("Error: %v", err)
	}
}

//--------------------
// HELPERS
//--------------------

// Create a SML structure.
func createSmlStructure() *TagNode {
	root := NewTagNode("root")

	root.AppendText("Text A")
	root.AppendText("Text B")

	root.AppendTaggedText("comment", "A first comment.")

	subA := root.AppendTag("sub-a:1st:important")

	subA.AppendText("Text A.A")

	root.AppendTaggedText("comment", "A second comment.")

	subB := root.AppendTag("sub-b:2nd")

	subB.AppendText("Text B.A")
	subB.AppendTaggedText("raw", "Any raw text with {, }, and ^.")

	return root
}

// EOF
