// Tideland Common Go Library - Networking / Atom
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package atom

//--------------------
// IMPORTS
//--------------------

import (
	"code.google.com/p/tcgl/net"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"
)

//--------------------
// CONST
//--------------------

const (
	Version = "1.0"
	XMLNS   = "http://www.w3.org/2005/Atom"
)

//--------------------
// MODEL
//--------------------

// Feed is the root element of the document.
type Feed struct {
	XMLName      string        `xml:"feed"`
	XMLNS        string        `xml:"xmlns,attr"`
	Title        Text          `xml:"title"`
	Subtitle     Text          `xml:"subtitle,omitempty"`
	Id           string        `xml:"id"`
	Updated      string        `xml:"updated"`
	Author       Author        `xml:"author,omitempty"`
	Link         Link          `xml:"link,omitempty"`
	Categories   []Category    `xml:"category,omitempty"`
	Contributors []Contributor `xml:"contributor,omitempty"`
	Generator    Generator     `xml:"generator,omitempty"`
	Icon         string        `xml:"icon,omitempty"`
	Logo         string        `xml:"logo,omitempty"`
	Rights       Text          `xml:"rights,omitempty"`
	Entries      []Entry       `xml:"entry"`
}

// Text contains human-readable text, usually in small quantities. The type 
// attribute determines how this information is encoded.
type Text struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr,omitempty"`
}

// Author names the author of the feed.
type Author struct {
	Name  string `xml:"name"`
	EMail string `xml:"email"`
	URI   string `xml:"uri"`
}

// Link identifies a related web page.
type Link struct {
	HRef     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	HRefLang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   int    `xml:"lenght,attr,omitempty"`
}

// Category specifies a category that the feed belongs to.
type Category struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

// Contributor names one contributor to the feed.
type Contributor struct {
	Names string `xml:"name"`
}

// Generator identifies the software used to generate the feed, 
// for debugging and other purposes.
type Generator struct {
	Generator string `xml:",chardata"`
	URI       string `xml:"uri,attr,omitempty"`
	Version   string `xml:"version,attr,omitempty"`
}

// Entry defines one feed entry.
type Entry struct {
	Id           string        `xml:"id"`
	Title        Text          `xml:"title"`
	Updated      string        `xml:"updated"`
	Author       Author        `xml:"author,omitempty"`
	Content      Text          `xml:"content,omitempty"`
	Link         Link          `xml:"link,omitempty"`
	Summary      Text          `xml:"subtitle,omitempty"`
	Categories   []Category    `xml:"category,omitempty"`
	Contributors []Contributor `xml:"contributor,omitempty"`
	Published    string        `xml:"published,omitempty"`
	Rights       Text          `xml:"rights,omitempty"`
}

// Source preserves the source feeds metadata if the entry is copied
// from one feed into another feed.
type Source struct {
	Title        Text          `xml:"title,omitempty"`
	Subtitle     Text          `xml:"subtitle,omitempty"`
	Id           string        `xml:"id",omitempty`
	Updated      string        `xml:"updated,omitempty"`
	Author       Author        `xml:"author,omitempty"`
	Categories   []Category    `xml:"category,omitempty"`
	Contributors []Contributor `xml:"contributor,omitempty"`
	Generator    Generator     `xml:"generator,omitempty"`
	Link         Link          `xml:"link,omitempty"`
	Icon         string        `xml:"icon,omitempty"`
	Logo         string        `xml:"logo,omitempty"`
	Rights       Text          `xml:"rights,omitempty"`
}

//--------------------
// FUNCTIONS
//--------------------

// ParseTime analyzes the Atom date/time string and returns it as Go time.
func ParseTime(s string) (t time.Time, err error) {
	formats := []string{time.RFC3339, time.RFC3339Nano}
	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			return
		}
	}
	return
}

// ComposeTime takes a Go time and converts it into a valid Atom time string.
func ComposeTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Encode writes the feed to the writer.
func Encode(w io.Writer, feed *Feed) error {
	enc := xml.NewEncoder(w)
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(feed)
}

// Decode reads the feed from the reader.
func Decode(r io.Reader) (*Feed, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = net.CharsetReader
	feed := &Feed{}
	if err := dec.Decode(feed); err != nil {
		return nil, err
	}
	return feed, nil
}

// Get retrieves a feed from the given URL.
func Get(u *url.URL) (*Feed, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Decode(resp.Body)
}

// EOF
