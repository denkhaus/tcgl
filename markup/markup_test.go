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
	"cgl.tideland.biz/asserts"
	"bytes"
	"strings"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test creating.
func TestSmlCreating(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	root := createSmlStructure()
	assert.Equal(root.Tag(), "root", "Root tag has to be 'root'.")
	assert.NotEmpty(root, "Root tag is not empty.")
}

// Test SML writer processing.
func TestSmlWriterProcessing(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	root := createSmlStructure()
	bufA := bytes.NewBufferString("")
	bufB := bytes.NewBufferString("")
	sppA := NewSmlWriterProcessor(bufA, true)
	sppB := NewSmlWriterProcessor(bufB, false)

	root.ProcessWith(sppA)
	root.ProcessWith(sppB)

	assert.NotEmpty(bufA, "Buffer A should not be empty.")
	assert.NotEmpty(bufB, "Buffer B should not be empty.")
}

// Test positive reading.
func TestSmlPositiveReading(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sml := "Before!   {foo {bar:1:first Yadda ^{Test^} 1}  {inbetween}  {bar:2:last Yadda {Test ^^} 2}}   After!"
	reader := NewSmlReader(strings.NewReader(sml))
	root, err := reader.RootTagNode()
	assert.Nil(err, "Expected no reader error.")
	assert.Equal(root.Tag(), "foo", "Root tag is 'foo'.")
	assert.NotEmpty(root, "Root tag is not empty.")
}

// Test negative reading.
func TestSmlNegativeReading(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sml := "{Foo {bar:1 Yadda {test} {} 1} {bar:2 Yadda 2}}"
	reader := NewSmlReader(strings.NewReader(sml))
	_, err := reader.RootTagNode()
	assert.ErrorMatch(err, "invalid rune.*", "Invalid rune should be found.")
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
