// Tideland Common Go Library - Identifier
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
	"bytes"
	"crypto/rand"
	"fmt"
	"encoding/hex"
	"io"
	"reflect"
	"strings"
	"unicode"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Identifier - Release 2012-02-16"

//--------------------
// UUID
//--------------------

// UUID represent a universal identifier with 16 bytes.
type UUID []byte

// NewUUID generates a new UUID based on version 4.
func NewUUID() UUID {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(err)
	}
	// Set version (4) and variant (2).
	var version byte = 4 << 4
	var variant byte = 2 << 4
	uuid[6] = version | (uuid[6] & 15)
	uuid[8] = variant | (uuid[8] & 15)
	return uuid
}

// Raw returns a copy of the UUID bytes.
func (uuid UUID) Raw() []byte {
	raw := make([]byte, 16)
	copy(raw, uuid[0:16])
	return raw
}

// String returns a hexadecimal string representation with
// standardized separators.
func (uuid UUID) String() string {
	base := hex.EncodeToString(uuid.Raw())
	return base[0:8] + "-" + base[8:12] + "-" + base[12:16] + "-" + base[16:20] + "-" + base[20:32]
}

//--------------------
// MORE ID FUNCTIONS
//--------------------

// LimitedSepIdentifier builds an identifier out of multiple parts, 
// all as lowercase strings and concatenated with the separator
// Non letters and digits are exchanged with dashes and
// reduced to a maximum of one each. If limit is true only
// 'a' to 'z' and '0' to '9' are allowed.
func LimitedSepIdentifier(sep string, limit bool, parts ...interface{}) string {
	iparts := make([]string, 0)
	for _, p := range parts {
		tmp := strings.Map(func(r rune) rune {
			// Check letter and digit.
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				lcr := unicode.ToLower(r)
				if limit {
					// Only 'a' to 'z' and '0' to '9'.
					if lcr <= unicode.MaxASCII {
						return lcr
					} else {
						return ' '
					}
				} else {
					// Every char is allowed.
					return lcr
				}
			}
			return ' '
		}, fmt.Sprintf("%v", p))
		// Only use non-empty identifier parts.
		if ipart := strings.Join(strings.Fields(tmp), "-"); len(ipart) > 0 {
			iparts = append(iparts, ipart)
		}
	}
	return strings.Join(iparts, sep)
}

// SepIdentifier builds an identifier out of multiple parts, all
// as lowercase strings and concatenated with the separator
// Non letters and digits are exchanged with dashes and
// reduced to a maximum of one each.
func SepIdentifier(sep string, parts ...interface{}) string {
	return LimitedSepIdentifier(sep, false, parts...)
}

// Identifier works like SepIdentifier but the seperator
// is set to be a colon.
func Identifier(parts ...interface{}) string {
	return SepIdentifier(":", parts...)
}

// TypeAsIdentifierPart transforms the name of the arguments type into 
// a part for identifiers. It's splitted at each uppercase char, 
// concatenated with dashes and transferred to lowercase.
func TypeAsIdentifierPart(i interface{}) string {
	var buf bytes.Buffer
	fullTypeName := reflect.TypeOf(i).String()
	lastDot := strings.LastIndex(fullTypeName, ".")
	typeName := fullTypeName[lastDot+1:]
	for i, r := range typeName {
		if unicode.IsUpper(r) {
			if i > 0 {
				buf.WriteRune('-')
			}
		}
		buf.WriteRune(r)
	}
	return strings.ToLower(buf.String())
}

// EOF
