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
	"http"
	"log"
	"path/filepath"
	"strings"
)

//--------------------
// LOGGING HANDLER
//--------------------

// LoggingHandler logs the request using the passed function.
type LoggingHandler struct {
	prefix       string
	shortPattern string
	longPattern  string
}

// NewLoggingHandler creates a new logging handler.
func NewLoggingHandler(p string) *LoggingHandler {
	return &LoggingHandler{p, "", ""}
}

// Init initializes the pattern for future logging.
func (lh *LoggingHandler) Init(domain, resource string) {
	lh.shortPattern = "[" + lh.prefix + "] %v /" + domain + "/" + resource
	lh.longPattern = lh.shortPattern + "/%v"
}

// Get handles a GET request.
func (lh *LoggingHandler) Get(ctx *Context) bool {
	if ctx.ResourceId == "" {
		log.Printf(lh.shortPattern, ctx.Request.Method)
	} else {
		log.Printf(lh.longPattern, ctx.Request.Method, ctx.ResourceId)
	}

	return true
}

// Put handles a PUT request.
func (lh *LoggingHandler) Put(ctx *Context) bool {
	return lh.Get(ctx)
}

// Post handles a POST request.
func (lh *LoggingHandler) Post(ctx *Context) bool {
	return lh.Get(ctx)
}

// Delete handles a DELETE request.
func (lh *LoggingHandler) Delete(ctx *Context) bool {
	return lh.Get(ctx)
}

//--------------------
// WRAPPER HANDLER
//--------------------

// WrapperHandler wraps existing handler functions for a usage inside
// the framework.
type WrapperHandler struct {
	handle func(http.ResponseWriter, *http.Request)
}

// NewWrapperHandler creates a new wrapper around a handler function.
func NewWrapperHandler(h func(http.ResponseWriter, *http.Request)) *WrapperHandler {
	return &WrapperHandler{h}
}

// Init does nothing here.
func (wh *WrapperHandler) Init(domain, resource string) {
}

// Get handles a GET request.
func (wh *WrapperHandler) Get(ctx *Context) bool {
	wh.handle(ctx.ResponseWriter, ctx.Request)

	return true
}

// Put handles a PUT request.
func (wh *WrapperHandler) Put(ctx *Context) bool {
	wh.handle(ctx.ResponseWriter, ctx.Request)

	return true
}

// Post handles a POST request.
func (wh *WrapperHandler) Post(ctx *Context) bool {
	wh.handle(ctx.ResponseWriter, ctx.Request)

	return true
}

// Delete handles a DELETE request.
func (wh *WrapperHandler) Delete(ctx *Context) bool {
	wh.handle(ctx.ResponseWriter, ctx.Request)

	return true
}

//--------------------
// FILE SERVING HANDLER
//--------------------

// FileServingHandler serves the file out of one directory.
type FileServingHandler struct {
	dir string
}

// NewFileServingHandler creates a new handler with a directory.
func NewFileServingHandler(d string) *FileServingHandler {
	pd := filepath.FromSlash(d)

	if !strings.HasSuffix(pd, string(filepath.Separator)) {
		pd += string(filepath.Separator)
	}

	return &FileServingHandler{pd}
}

// Init does nothing here.
func (fsh *FileServingHandler) Init(domain, resource string) {
}

// Get handles a GET request.
func (fsh *FileServingHandler) Get(ctx *Context) bool {
	http.ServeFile(ctx.ResponseWriter, ctx.Request, fsh.dir+ctx.ResourceId)

	return true
}

// EOF
