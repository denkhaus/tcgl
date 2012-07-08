// Tideland Common Go Library - Web
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package web

//--------------------
// IMPORTS
//--------------------

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"text/template"
)

//--------------------
// CONST
//--------------------

const (
	CT_PLAIN = "text/plain"
	CT_HTML  = "text/html"
	CT_XML   = "application/xml"
	CT_JSON  = "application/json"
	CT_GOB   = "application/vnd.tideland.gob"
)

//--------------------
// KEY VALUE
//--------------------

// KeyValue assigns a value to a key.
type KeyValue struct {
	Key   string
	Value interface{}
}

// String prints the encoded form key=value for URLs.
func (kv KeyValue) String() string {
	return fmt.Sprintf("%v=%v", url.QueryEscape(kv.Key), url.QueryEscape(fmt.Sprintf("%v", kv.Value)))
}

//--------------------
// HTML HELPER
//--------------------

// HtmlInternalReference builds an internal reference out of the passed parts.
func HtmlInternalReference(domain, resource, resourceId string, query ...KeyValue) string {
	ref := BasePath() + domain + "/" + resource
	if resourceId != "" {
		ref = ref + "/" + resourceId
	}
	queryParts := len(query)
	lastIdx := queryParts - 1
	if queryParts > 0 {
		ref = ref + "?"
		for idx, part := range query {
			ref = ref + part.String()
			if idx < lastIdx {
				ref = ref + "&"
			}
		}
	}
	return ref
}

//--------------------
// GOB MARSHALING HELPER
//--------------------

// MarshalGob marshals data to a GOB encoded byte slice. It also
// returns the content-type.
func StreamAsGob(w io.Writer, h http.Header, data interface{}) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(data)
	h.Set("Content-Type", CT_GOB)
	return err
}

//--------------------
// TEMPLATE CACHE
//--------------------

// templateCacheEntry stores the parsed template and the
// content type.
type templateCacheEntry struct {
	parsedTemplate *template.Template
	contentType    string
}

// templateCache stores preparsed templates.
type templateCache struct {
	cache map[string]*templateCacheEntry
	mutex sync.RWMutex
}

// newTemplateCache creates a new cache.
func newTemplateCache() *templateCache {
	return &templateCache{
		cache: make(map[string]*templateCacheEntry),
	}
}

// parse parses a template an stores it.
func (tc *templateCache) parse(id, t, ct string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tmpl, err := template.New(id).Parse(t)
	if err != nil {
		panic(err)
	}
	tc.cache[id] = &templateCacheEntry{tmpl, ct}
}

// loadAndParse loads a template out of the filesystem, parses and stores it.
func (tc *templateCache) loadAndParse(id, fn, ct string) {
	t, _ := ioutil.ReadFile(fn)
	tc.parse(id, string(t), ct)
}

// render executes the pre-parsed template with the data. It also sets
// the content type header.
func (tc *templateCache) render(rw http.ResponseWriter, id string, data interface{}) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	rw.Header().Set("Content-Type", tc.cache[id].contentType)
	err := tc.cache[id].parsedTemplate.Execute(rw, data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// EOF
