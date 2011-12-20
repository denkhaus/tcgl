// Tideland Common Go Library - Web
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package web

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"gob"
	"http"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"template"
	"url"
)

//--------------------
// CONST
//--------------------

const (
	CT_PLAIN = "text/plain"
	CT_HTML  = "text/html"
	CT_XML   = "application/xml"
	CT_JSON  = "application/json"
	CT_GOB   = "application/vnd.tideland.rwf"
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
// LOGGER
//--------------------

// Logger is the interface for different logger implementations.
type Logger interface {
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warningf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Criticalf(format string, args ...interface{})
}

// StandardLogger is a logger implementation using the log package.
type StandardLogger struct {
	logger *log.Logger
}

// NewStandardLogger creates a logger using the log package.
func NewStandardLogger(out io.Writer, prefix string, flag int) *StandardLogger {
	return &StandardLogger{
		logger: log.New(out, prefix, flag),
	}
}

// Debugf logs a message at debug level.
func (sl *StandardLogger) Debugf(format string, args ...interface{}) {
	sl.logger.Printf("[debug] " + format, args...)
}

// Infof logs a message at info level.
func (sl *StandardLogger) Infof(format string, args ...interface{}) {
	sl.logger.Printf("[info] " + format, args...)
}

// Warningf logs a message at warning level.
func (sl *StandardLogger) Warningf(format string, args ...interface{}) {
	sl.logger.Printf("[warning] " + format, args...)
}

// Errorf logs a message at error level.
func (sl *StandardLogger) Errorf(format string, args ...interface{}) {
	sl.logger.Printf("[error] " + format, args...)
}

// Criticalf logs a message at critical level.
func (sl *StandardLogger) Criticalf(format string, args ...interface{}) {
	sl.logger.Printf("[critical] " + format, args...)
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
func StreamAsGob(w io.Writer, h http.Header, data interface{}) os.Error {
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
                http.Error(rw, err.String(), http.StatusInternalServerError)
        }
}

// EOF
