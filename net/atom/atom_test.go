// Tideland Common Go Library - Networking / Atom - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package atom_test

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"code.google.com/p/tcgl/applog"
	"code.google.com/p/tcgl/asserts"
	"code.google.com/p/tcgl/net/atom"
	"net/url"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test parsing and composing of date/times.
func TestParseComposeTime(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	nowOne := time.Now()
	nowStr := atom.ComposeTime(nowOne)

	applog.Infof("Now as string: %s", nowStr)

	year, month, day := nowOne.Date()
	hour, min, sec := nowOne.Clock()
	loc := nowOne.Location()
	nowCmp := time.Date(year, month, day, hour, min, sec, 0, loc)
	nowTwo, err := atom.ParseTime(nowStr)

	assert.Nil(err, "No error during time parsing.")
	assert.Equal(nowCmp, nowTwo, "Both times have to be equal.")
}

// Test encoding and decoding a doc.
func TestEncodeDecode(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	a1 := &atom.Feed{
		XMLNS:   atom.XMLNS,
		Id:      "http://tideland.biz/pkg/net/atom",
		Title:   &atom.Text{"Test Encode/Decode", "", "text"},
		Updated: atom.ComposeTime(time.Now()),
		Entries: []*atom.Entry{
			{
				Id:      "http://tideland.bi/pkg/net/atom/entry-1",
				Title:   &atom.Text{"Entry 1", "", "text"},
				Updated: atom.ComposeTime(time.Now()),
			},
			{
				Id:      "http://tideland.bi/pkg/net/atom/entry-2",
				Title:   &atom.Text{"Entry 2", "", "text"},
				Updated: atom.ComposeTime(time.Now()),
			},
		},
	}
	b := &bytes.Buffer{}

	err := atom.Encode(b, a1)
	assert.Nil(err, "Encoding returns no error.")
	assert.Substring(b.String(), `<title type="text">Test Encode/Decode</title>`, "Title has been encoded correctly.")

	a2, err := atom.Decode(b)
	assert.Nil(err, "Decoding returns no error.")
	assert.Equal(a2.Title.Text, "Test Encode/Decode", "Title has been decoded correctly.")
	assert.Length(a2.Entries, 2, "Decoded feed has the right number of items.")
}

// Test getting a feed.
func TestGet(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	u, _ := url.Parse("http://mue.tideland.biz/feeds/posts/default")
	f, err := atom.Get(u)
	assert.Nil(err, "Getting the Atom document returns no error.")
	err = f.Validate()
	assert.Nil(err, "Validating returns no error.")
	b := &bytes.Buffer{}
	err = atom.Encode(b, f)
	assert.Nil(err, "Encoding returns no error.")
	applog.Infof("--- Atom ---\n%s", b)
}

// EOF
