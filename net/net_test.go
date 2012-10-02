// Tideland Common Go Library - Networking - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package net_test

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/asserts"
	"cgl.tideland.biz/net"
	"testing"
)

//--------------------
// TESTS
//--------------------

func TestStripTags(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	in := "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog</em>.</p>"
	out, err := net.StripTags(in, true, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox jumps over the lazy dog .", "Tags have been removed.")

	in = "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog.</p>"
	out, err = net.StripTags(in, true, false)
	assert.ErrorMatch(err, `XML syntax error on line 1.*`, "Error in document detected.")

	in = "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog.</p>"
	out, err = net.StripTags(in, false, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox jumps over the lazy dog.", "Tags have been removed.")

	in = "<p>The quick brown <b>fox &amp; goose</b> jump over the lazy &lt;em&gt;dog&lt;/em&gt;.</p>"
	out, err = net.StripTags(in, true, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox & goose jump over the lazy <em>dog</em>.", "Tags have been removed.")

	in = "<p>The quick brown <b>fox &amp;amp; goose</b> jump over the lazy &lt;em&gt;dog&lt;/em&gt;.</p>"
	out, err = net.StripTags(in, true, true)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox & goose jump over the lazy dog .", "Tags have been removed.")
}

// EOF
