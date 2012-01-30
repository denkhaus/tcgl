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
	"net/http"
	"path/filepath"
	"strings"
)

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