// Tideland Common Go Library - Simple Markup Language - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package markup_test

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/markup"
	"strings"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test creating.
func TestSMLCreating(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	root := createSMLStructure()
	assert.Equal(root.Tag(), "root", "Root tag has to be 'root'.")
	assert.NotEmpty(root, "Root tag is not empty.")
}

// Test SML writer processing.
func TestSMLWriterProcessing(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	root := createSMLStructure()
	bufA := bytes.NewBufferString("")
	bufB := bytes.NewBufferString("")

	markup.WriteSML(root, bufA, true)
	markup.WriteSML(root, bufB, false)

	println("===== WITH INDENT =====")
	println(bufA.String())
	println("===== WITHOUT INDENT =====")
	println(bufB.String())
	println("===== DONE =====")

	assert.NotEmpty(bufA, "Buffer A should not be empty.")
	assert.NotEmpty(bufB, "Buffer B should not be empty.")
}

// Test positive reading.
func TestSMLPositiveReading(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sml := "Before!   {foo {bar:1:first Yadda ^{Test^} 1} {! Raw: }} { ! ^^^ !}  {inbetween}  {bar:2:last Yadda {Test ^^} 2}}   After!"
	builder := markup.NewNodeBuilder()
	err := markup.ReadSML(strings.NewReader(sml), builder)
	assert.Nil(err, "Expected no reader error.")
	root := builder.Root()
	assert.Equal(root.Tag(), "foo", "Root tag is 'foo'.")
	assert.NotEmpty(root, "Root tag is not empty.")

	buf := bytes.NewBufferString("")
	markup.WriteSML(root, buf, true)

	println("===== PARSED SML =====")
	println(buf.String())
	println("===== DONE =====")
}

// Test negative reading.
func TestSMLNegativeReading(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	sml := "{Foo {bar:1 Yadda {test} {} 1} {bar:2 Yadda 2}}"
	builder := markup.NewNodeBuilder()
	err := markup.ReadSML(strings.NewReader(sml), builder)
	assert.ErrorMatch(err, "invalid rune.*", "Invalid rune should be found.")
}

//--------------------
// HELPERS
//--------------------

// Create a SML structure.
func createSMLStructure() *markup.TagNode {
	root := markup.NewTagNode("root")
	root.AppendTextNode("Text A")
	root.AppendTextNode("Text B")
	root.AppendTaggedTextNode("comment", "A first comment.")
	subA := root.AppendTagNode("sub-a:1st:important").(*markup.TagNode)
	subA.AppendTextNode("Text A.A")
	root.AppendTaggedTextNode("comment", "A second comment.")
	subB := root.AppendTagNode("sub-b:2nd").(*markup.TagNode)
	subB.AppendTextNode("Text B.A")
	subB.AppendTaggedTextNode("text", "Any text with the special characters {, }, and ^.")
	subC := root.AppendTagNode("sub-c").(*markup.TagNode)
	subC.AppendTextNode("Before raw.")
	subC.AppendRawNode("func Test(i int) { println(i) }")
	subC.AppendTextNode("After raw.")
	return root
}

// EOF
