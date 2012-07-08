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
	"cgl.tideland.biz/net"
	"encoding/xml"
	"fmt"
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

	TextType  = "text"
	HTMLType  = "html"
	XHTMLType = "xhtml"

	AlternateRel = "alternate"
	EnclosureRel = "enclosure"
	RelatedRel   = "related"
	SelfRel      = "self"
	ViaRel       = "via"
)

//--------------------
// MODEL
//--------------------

// Feed is the root element of the document.
type Feed struct {
	XMLName      string         `xml:"feed"`
	XMLNS        string         `xml:"xmlns,attr"`
	Id           string         `xml:"id"`
	Title        *Text          `xml:"title"`
	Updated      string         `xml:"updated"`
	Authors      []*Author      `xml:"author,omitempty"`
	Link         *Link          `xml:"link,omitempty"`
	Categories   []*Category    `xml:"category,omitempty"`
	Contributors []*Contributor `xml:"contributor,omitempty"`
	Generator    *Generator     `xml:"generator,omitempty"`
	Icon         string         `xml:"icon,omitempty"`
	Logo         string         `xml:"logo,omitempty"`
	Rights       *Text          `xml:"rights,omitempty"`
	Subtitle     *Text          `xml:"subtitle,omitempty"`
	Entries      []*Entry       `xml:"entry"`
}

// Validate checks if the feed is valid.
func (f *Feed) Validate() error {
	if f.XMLNS != XMLNS {
		return newInvalidAtomError("feed namespace %q has to be %q", f.XMLNS, XMLNS)
	}
	if _, err := url.Parse(f.Id); err != nil {
		return newInvalidAtomError("feed id is not parsable: %v", err)
	}
	if err := validateText("feed title", f.Title, true); err != nil {
		return err
	}
	if _, err := ParseTime(f.Updated); err != nil {
		return newInvalidAtomError("feed update is not parsable: %v", err)
	}
	for _, author := range f.Authors {
		if err := author.Validate(); err != nil {
			return err
		}
	}
	if f.Link != nil {
		if err := f.Link.Validate(); err != nil {
			return err
		}
	}
	for _, category := range f.Categories {
		if err := category.Validate(); err != nil {
			return err
		}
	}
	for _, contributor := range f.Contributors {
		if err := contributor.Validate(); err != nil {
			return err
		}
	}
	if f.Generator != nil {
		if err := f.Generator.Validate(); err != nil {
			return err
		}
	}
	if err := validateText("feed rights", f.Rights, false); err != nil {
		return err
	}
	if err := validateText("feed subtitle", f.Subtitle, false); err != nil {
		return err
	}
	allEntriesWithAuthor := true
	for _, entry := range f.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		allEntriesWithAuthor = allEntriesWithAuthor && len(entry.Authors) > 0
	}
	if !allEntriesWithAuthor && len(f.Authors) == 0 {
		return newInvalidAtomError("feed needs at least one author or one or more in each entry")
	}
	return nil
}

// Text contains human-readable text, usually in small quantities. The type 
// attribute determines how this information is encoded.
type Text struct {
	Text string `xml:",chardata"`
	Src  string `xml:"src,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// validateText ensures that a text is set if it's mandatory and that
// the type is correct.
func validateText(description string, t *Text, mandatory bool) error {
	if (t == nil || t.Text == "") && mandatory {
		return newInvalidAtomError("%s must not be missing or empty", description)
	}
	if t != nil {
		if t.Src != "" {
			if _, err := url.Parse(t.Src); err != nil {
				return newInvalidAtomError("%s src is not parsable: %v", description, err)
			}
		}
		switch t.Type {
		case "", TextType, HTMLType, XHTMLType:
			// OK.
		default:
			return newInvalidAtomError("%s has illegal type %q", description, t.Type)
		}
	}
	return nil
}

// Author names the author of the feed.
type Author struct {
	Name  string `xml:"name"`
	URI   string `xml:"uri,omitempty"`
	EMail string `xml:"email,omitempty"`
}

// Validate checks if a feed author is valid.
func (a *Author) Validate() error {
	if a.Name == "" {
		return newInvalidAtomError("feed author name must not be empty")
	}
	if a.URI != "" {
		if _, err := url.Parse(a.URI); err != nil {
			return newInvalidAtomError("feed author uri is not parsable:", err)
		}
	}
	return nil
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

// Validate checks if the feed link is valid.
func (l *Link) Validate() error {
	if _, err := url.Parse(l.HRef); err != nil {
		return newInvalidAtomError("feed link href is not parsable:", err)
	}
	switch l.Rel {
	case "", AlternateRel, EnclosureRel, RelatedRel, SelfRel, ViaRel:
		// OK.
	default:
		if _, err := url.Parse(l.Rel); err != nil {
			return newInvalidAtomError("feed link rel is neither predefined nor parsable: %v", err)
		}
	}
	return nil
}

// Category specifies a category that the feed belongs to.
type Category struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

// Validate checks if a feed category is valid.
func (c *Category) Validate() error {
	if c.Term == "" {
		return newInvalidAtomError("feed category term must not be empty")
	}
	if c.Scheme != "" {
		if _, err := url.Parse(c.Scheme); err != nil {
			return newInvalidAtomError("feed category scheme is not parsable: %v", err)
		}
	}
	return nil
}

// Contributor names one contributor of the feed.
type Contributor struct {
	Name string `xml:"name"`
}

// Validate checks if a feed contributor is valid.
func (c *Contributor) Validate() error {
	if c.Name == "" {
		return newInvalidAtomError("feed contributor name must not be empty")
	}
	return nil
}

// Generator identifies the software used to generate the feed, 
// for debugging and other purposes.
type Generator struct {
	Generator string `xml:",chardata"`
	URI       string `xml:"uri,attr,omitempty"`
	Version   string `xml:"version,attr,omitempty"`
}

// Validate checks if a feed generator is valid.
func (g *Generator) Validate() error {
	if g.Generator == "" {
		return newInvalidAtomError("feed generator must not be empty")
	}
	if g.URI != "" {
		if _, err := url.Parse(g.URI); err != nil {
			return newInvalidAtomError("feed generator URI is not parsable: %v", err)
		}
	}
	return nil
}

// Entry defines one feed entry.
type Entry struct {
	Id           string         `xml:"id"`
	Title        *Text          `xml:"title"`
	Updated      string         `xml:"updated"`
	Authors      []*Author      `xml:"author,omitempty"`
	Content      *Text          `xml:"content,omitempty"`
	Link         *Link          `xml:"link,omitempty"`
	Summary      *Text          `xml:"subtitle,omitempty"`
	Categories   []*Category    `xml:"category,omitempty"`
	Contributors []*Contributor `xml:"contributor,omitempty"`
	Published    string         `xml:"published,omitempty"`
	Source       *Source        `xml:"source,omitempty"`
	Rights       *Text          `xml:"rights,omitempty"`
}

// Validate checks if the feed entry is valid.
func (e *Entry) Validate() error {
	if _, err := url.Parse(e.Id); err != nil {
		return newInvalidAtomError("feed entry id is not parsable: %v", err)
	}
	if err := validateText("feed entry title", e.Title, true); err != nil {
		return err
	}
	if _, err := ParseTime(e.Updated); err != nil {
		return newInvalidAtomError("feed entry update is not parsable: %v", err)
	}
	for _, author := range e.Authors {
		if err := author.Validate(); err != nil {
			return err
		}
	}
	if err := validateText("feed entry content", e.Content, false); err != nil {
		return err
	}
	if e.Link != nil {
		if err := e.Link.Validate(); err != nil {
			return err
		}
	}
	if err := validateText("feed entry summary", e.Summary, false); err != nil {
		return err
	}
	for _, category := range e.Categories {
		if err := category.Validate(); err != nil {
			return err
		}
	}
	for _, contributor := range e.Contributors {
		if err := contributor.Validate(); err != nil {
			return err
		}
	}
	if _, err := ParseTime(e.Published); err != nil {
		return newInvalidAtomError("feed entry published is not parsable: %v", err)
	}
	if e.Source != nil {
		if err := e.Source.Validate(); err != nil {
			return err
		}
	}
	if err := validateText("feed entry rights", e.Rights, false); err != nil {
		return err
	}
	return nil
}

// Source preserves the source feeds metadata if the entry is copied
// from one feed into another feed.
type Source struct {
	Authors      []*Author      `xml:"author,omitempty"`
	Categories   []*Category    `xml:"category,omitempty"`
	Contributors []*Contributor `xml:"contributor,omitempty"`
	Generator    *Generator     `xml:"generator,omitempty"`
	Icon         string         `xml:"icon,omitempty"`
	Id           string         `xml:"id",omitempty`
	Link         *Link          `xml:"link,omitempty"`
	Logo         string         `xml:"logo,omitempty"`
	Rights       *Text          `xml:"rights,omitempty"`
	Subtitle     *Text          `xml:"subtitle,omitempty"`
	Title        *Text          `xml:"title,omitempty"`
	Updated      string         `xml:"updated,omitempty"`
}

// Validate checks if a feed entry source is valid.
func (s *Source) Validate() error {
	for _, author := range s.Authors {
		if err := author.Validate(); err != nil {
			return err
		}
	}
	for _, category := range s.Categories {
		if err := category.Validate(); err != nil {
			return err
		}
	}
	for _, contributor := range s.Contributors {
		if err := contributor.Validate(); err != nil {
			return err
		}
	}
	if s.Generator != nil {
		if err := s.Generator.Validate(); err != nil {
			return err
		}
	}
	if _, err := url.Parse(s.Id); err != nil {
		return newInvalidAtomError("feed entry source id is not parsable: %v", err)
	}
	if s.Link != nil {
		if err := s.Link.Validate(); err != nil {
			return err
		}
	}
	if err := validateText("feed entry source rights", s.Rights, false); err != nil {
		return err
	}
	if err := validateText("feed entry source subtitle", s.Subtitle, false); err != nil {
		return err
	}
	if err := validateText("feed entry source title", s.Title, false); err != nil {
		return err
	}
	if _, err := ParseTime(s.Updated); err != nil {
		return newInvalidAtomError("feed entry source update is not parsable: %v", err)
	}
	return nil
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

//--------------------
// ERRORS
//--------------------

// InvalidAtomError will be returned if a validation fails.
type InvalidAtomError struct {
	Err error
}

// newInvalidAtomError creates a new error for invalid atom feeds.
func newInvalidAtomError(format string, args ...interface{}) InvalidAtomError {
	return InvalidAtomError{fmt.Errorf(format, args...)}
}

// Error returns the error as string.
func (e InvalidAtomError) Error() string {
	return e.Err.Error()
}

// EOF
