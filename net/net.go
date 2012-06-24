// Tideland Common Go Library - Networking
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package net

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

//--------------------
// CHARSET READER
//--------------------

// iso88591CharsetReader converts ISO-8859-1 into UTF-8.
type iso88591CharsetReader struct {
	reader io.ByteReader
	buffer *bytes.Buffer
}

// newISO88591CharsetReader creates a new charset reader.
func newISO88591CharsetReader(reader io.Reader) *iso88591CharsetReader {
	buffer := bytes.NewBuffer(make([]byte, 0, utf8.UTFMax))
	return &iso88591CharsetReader{reader.(io.ByteReader), buffer}
}

// ReadByte reads one byte from the reader.
func (cr *iso88591CharsetReader) ReadByte() (b byte, err error) {
	if cr.buffer.Len() <= 0 {
		r, err := cr.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		if r < utf8.RuneSelf {
			return r, nil
		}
		cr.buffer.WriteRune(rune(r))
	}
	return cr.buffer.ReadByte()
}

// Read reads a number of byte from the reader. It's invalid in
// this context.
func (cr *iso88591CharsetReader) Read(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

var mapping = map[string]string{
	"":                "utf-8",
	"utf-8":           "utf-8",
	"iso-8859-1":      "iso-8859-1",
	"iso_8859-1:1987": "iso-8859-1",
	"iso-ir-100":      "iso-8859-1",
	"iso_8859-1":      "iso-8859-1",
	"latin1":          "iso-8859-1",
	"l1":              "iso-8859-1",
	"ibm819":          "iso-8859-1",
	"cp819":           "iso-8859-1",
	"csisolatin1":     "iso-8859-1",
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch mapping[strings.ToLower(charset)] {
	case "utf-8":
		return input, nil
	case "iso-8859-1":
		return newISO88591CharsetReader(input), nil
	}
	return nil, fmt.Errorf("charset %q is not supported", charset)
}

// EOF
