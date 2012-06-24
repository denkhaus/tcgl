// Tideland Common Go Library - Networking / RSS
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package rss

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
	Version   = "2.0"
	rssDate   = "Mon, 02 Jan 2006 15:04 MST"
	rssDateV1 = "02 Jan 2006 15:04 MST"
	rssDateV2 = "Mon, 02 Jan 2006 15:04 -0700"
	rssDateV3 = "02 Jan 2006 15:04 -0700"
)

//--------------------
// MODEL
//--------------------

// RSS is the root element of the document.
type RSS struct {
	XMLName string  `xml:"rss"`
	Version string  `xml:"version,attr"`
	Channel Channel `xml:"channel"`
}

// Channel is the one channel element of the RSS document.
type Channel struct {
	Description    string     `xml:"description"`
	Link           string     `xml:"link"`
	Title          string     `xml:"title"`
	Categories     []Category `xml:"category,omitempty"`
	Cloud          Cloud      `xml:"cloud,omitempty"`
	Copyright      string     `xml:"copyright,omitempty"`
	Docs           string     `xml:"docs,omitempty"`
	Generator      string     `xml:"generator,omitempty"`
	Image          Image      `xml:"image,omitempty"`
	Language       string     `xml:"language,omitempty"`
	LastBuildDate  string     `xml:"lastBuildDate,omitempty"`
	ManagingEditor string     `xml:"managingEditor,omitempty"`
	PubDate        string     `xml:"pubDate,omitempty"`
	Rating         string     `xml:"rating,omitempty"`
	SkipDays       SkipDays   `xml:"skipDays,omitempty"`
	SkipHours      SkipHours  `xml:"skipHours,omitempty"`
	TextInput      string     `xml:"textInput,omitempty"`
	TTL            int        `xml:"ttl,omitempty"`
	WebMaster      string     `xml:"webMaster,omitempty"`
	Items          []Item     `xml:"item,omitempty"`
}

// Category identifies a category or tag to which the feed belongs.
type Category struct {
	Category string `xml:",chardata"`
	Domain   string `xml:"domain,attr,omitempty"`
}

// Cloud indicates that updates to the feed can be monitored using a web service 
// that implements the RssCloud application programming interface.
type Cloud struct {
	Domain            string `xml:"domain,attr"`
	Port              int    `xml:"port,attr,omitempty"`
	Path              string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol          string `xml:"protocol,attr"`
}

// Image supplies a graphical logo for the feed .
type Image struct {
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	URL         string `xml:"url"`
	Description string `xml:"description,omitempty"`
	Height      int    `xml:"height,omitempty"`
	Width       int    `xml:"width,omitempty"`
}

// SkipDays identifies days of the week during which the feed is not updated.
type SkipDays struct {
	Days []string `xml:"day"`
}

// SkipHours identifies the hours of the day during which the feed is not updated.
type SkipHours struct {
	Hours []int `xml:"hour"`
}

// TextInput defines a form to submit a text query to the feed's publisher over 
// the Common Gateway Interface (CGI).
type TextInput struct {
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Name        string `xml:"name"`
	Title       string `xml:"title"`
}

// Item represents distinct content published in the feed such as a news article, 
// weblog entry or some other form of discrete update. It must contain either a
// title or description.
type Item struct {
	Title       string     `xml:"title,omitempty"`
	Description string     `xml:"description,omitempty"`
	Author      string     `xml:"author,omitempty"`
	Categories  []Category `xml:"category,omitempty"`
	Comments    string     `xml:"comments,omitempty"`
	Enclosure   Enclosure  `xml:"enclosure,omitempty"`
	GUID        GUID       `xml:"guid,omitempty"`
	Link        string     `xml:"link,omitempty"`
	PubDate     string     `xml:"pubDate,omitempty"`
	Source      Source     `xml:"source,omitempty"`
}

// Enclosure associates a media object such as an audio or video file with the item.
type Enclosure struct {
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
}

// GUID provides a string that uniquely identifies the item.
type GUID struct {
	GUID        string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr,omitempty"`
}

// Source indicates the fact that the item has been republished from another RSS feed.
type Source struct {
	Source string `xml:",chardata"`
	URL    string `xml:"url,attr"`
}

//--------------------
// FUNCTIONS
//--------------------

// ParseTime analyzes the RSS date/time string and returns it as Go time.
func ParseTime(s string) (t time.Time, err error) {
	formats := []string{rssDate, rssDateV1, rssDateV2, rssDateV3, time.RFC822, time.RFC822Z}
	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			return
		}
	}
	return
}

// ComposeTime takes a Go time and converts it into a valid RSS time string.
func ComposeTime(t time.Time) string {
	return t.Format(rssDate)
}

// Encode writes the RSS document to the writer.
func Encode(w io.Writer, rss *RSS) error {
	enc := xml.NewEncoder(w)
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(rss)
}

// Decode reads the RSS document from the reader.
func Decode(r io.Reader) (*RSS, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = net.CharsetReader
	rss := &RSS{}
	if err := dec.Decode(rss); err != nil {
		return nil, err
	}
	return rss, nil
}

// Get retrieves an RSS document from the given URL.
func Get(u *url.URL) (*RSS, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Decode(resp.Body)
}

// EOF
