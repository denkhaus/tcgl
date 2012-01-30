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
	"code.google.com/p/tcgl/identifier"
	"code.google.com/p/tcgl/monitoring"
	"code.google.com/p/tcgl/util"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library -  Web - Release 2012-01-29"

//--------------------
// RESOURCE HANDLER
//--------------------

// ResourceHandler is the interface for all resource
// handlers understanding the REST verbs. Beside initialization
// a handler has to understand the verb GET.
type ResourceHandler interface {
	Init(domain, resource string)
	Get(ctx *Context) bool
}

// PutResourceHandler is the additional interface for
// handlers understanding the verb PUT.
type PutResourceHandler interface {
	Put(ctx *Context) bool
}

// PostResourceHandler is the additional interface for
// handlers understanding the verb POST.
type PostResourceHandler interface {
	Post(ctx *Context) bool
}

// DeleteResourceHandler is the additional interface for
// handlers understanding the verb DELETE.
type DeleteResourceHandler interface {
	Delete(ctx *Context) bool
}

//--------------------
// CONFIGURATION
//--------------------

// handlerSlice is a list of resource handler.
type handlerSlice []ResourceHandler

// resourceMapping maps a resource id to a slice of resource handler.
type resourceMapping map[string]handlerSlice

// domainMapping maps a domain id to a resource mapping.
type domainMapping map[string]resourceMapping

//--------------------
// SERVER
//--------------------

// Server is the backend of the RWF.
type server struct {
	address         string
	basePath        string
	defaultDomain   string
	defaultResource string
	domains         domainMapping
	templateCache   *templateCache
	logger		util.Logger
}

// The central server.
var srv *server

// lazyCreateServer creates the central server instance
// configuration and work if this isn't yet done.
func lazyCreateServer() {
	if srv == nil {
		// Create the server.
		srv = &server{
			basePath:        "/",
			defaultDomain:   "default",
			defaultResource: "default",
			domains:         make(domainMapping),
			templateCache:   newTemplateCache(),
			logger:		 util.NewStandardLogger(os.Stdout, "[rwf] ", log.Ldate|log.Ltime),
		}
	}
}

// prepareServer prepares the server based on the passed
// configuration information.
func prepareServer(address, basePath string) {
	srv.address = address
	srv.basePath = basePath
	// Check passed parameters.
	if srv.address == "" {
		srv.address = ":8080"
	}
	if !strings.HasSuffix(srv.basePath, "/") {
		srv.basePath += "/"
	}
}

// handleFunc is the main function of the RWF server dispatching the
// requests to registered resource handler.
func handleFunc(rw http.ResponseWriter, r *http.Request) {
	ctx := newContext(rw, r)
	resources := srv.domains[ctx.Domain]
	if resources != nil {
		handlers := resources[ctx.Resource]
		if handlers != nil {
			m := monitoring.BeginMeasuring(identifier.Identifier("rwf", ctx.Domain, ctx.Resource, ctx.Request.Method))
			for _, h := range handlers {
				if !dispatch(ctx, h) {
					break
				}
			}
			m.EndMeasuring()
			return
		}
	}
	// No valid configuration, redirect to default (if not already).
	if ctx.Domain == srv.defaultDomain && ctx.Resource == srv.defaultResource {
		// No default handler registered.
		msg := fmt.Sprintf("domain '%v' and resource '%v' not found!", ctx.Domain, ctx.Resource)
		srv.logger.Errorf(msg)
		http.Error(ctx.ResponseWriter, msg, http.StatusNotFound)
	} else {
		// Redirect to default handler.
		srv.logger.Infof("domain '%v' and resource '%v' not found, redirecting to default", ctx.Domain, ctx.Resource)
		ctx.Redirect(srv.defaultDomain, srv.defaultResource, "")
	}
}

// Dispatch the encapsulated request to the according handler methods
// depending on the HTTP method.
func dispatch(ctx *Context, h ResourceHandler) bool {
	defer func() {
		if err := recover(); err != nil {
			// Shit happens! TODO: Better error handling.
			msg := fmt.Sprintf("internal server error: '%v' in context: '%v'", err, ctx)
			srv.logger.Criticalf(msg)
			http.Error(ctx.ResponseWriter, msg, http.StatusInternalServerError)
		}
	}()

	srv.logger.Infof("dispatching %s", ctx)
	switch ctx.Request.Method {
	case "GET":
		return h.Get(ctx)
	case "PUT":
		if ph, ok := h.(PutResourceHandler); ok {
			return ph.Put(ctx)
		}
	case "POST":
		if ph, ok := h.(PostResourceHandler); ok {
			return ph.Post(ctx)
		}
	case "DELETE":
		if dh, ok := h.(DeleteResourceHandler); ok {
			return dh.Delete(ctx)
		}
	}
	srv.logger.Errorf("method not allowed: %s", ctx)
	http.Error(ctx.ResponseWriter, "405 method not allowed", http.StatusMethodNotAllowed)
	return false
}

// StartServer has to be called with address and base path for the
// server. The resource handlers should be registered before but can also
// be added dynamically.
func StartServer(address, basePath string) {
	lazyCreateServer()
	prepareServer(address, basePath)
	http.HandleFunc(srv.basePath, handleFunc)
	http.ListenAndServe(srv.address, nil)
}

// SetDefault configures own default domain and resource ids.
func SetDefault(domain, resource string) {
	lazyCreateServer()
	srv.defaultDomain = domain
	srv.defaultResource = resource
}

// AttachToAppEngine initializes as attaches the RWF to the
// Google App Engine.
func AttachToAppEngine(basePath string) {
	lazyCreateServer()
	prepareServer("", basePath)
	http.HandleFunc(srv.basePath, handleFunc)
}

// AddResourceHandler assigns a resource handler to a domain and
// resource id. An existing one would be overwritten.
func AddResourceHandler(domain, resource string, handler ResourceHandler) ResourceHandler {
	lazyCreateServer()
	// Map domain to resources.
	resources := srv.domains[domain]
	if resources == nil {
		resources = make(resourceMapping)
		srv.domains[domain] = resources
	}
	// Map resource to handlers.
	handlers := resources[resource]
	if handlers == nil {
		handlers := make(handlerSlice, 1)
		resources[resource] = handlers
	}
	// Add and init handler.
	resources[resource] = append(handlers, handler)
	handler.Init(domain, resource)
	return handler
}

// ParseTemplate parses a template and stores it together with the 
// content type in the cache.
func ParseTemplate(templateId, template, contentType string) {
	srv.templateCache.parse(templateId, template, contentType)
}

// LoadAndParseTemplate loads a file, parses a template and stores it 
// together with the content type in the cache.
func LoadAndParseTemplate(templateId, filename, contentType string) {
	lazyCreateServer()
	srv.templateCache.loadAndParse(templateId, filename, contentType)
}

// BasePath returns the configured base path of the server. It's
// returned via a function so that it can't be changed after the 
// start of the server.
func BasePath() string {
	lazyCreateServer()
	return srv.basePath
}

// SetLogger sets a new logger.
func SetLogger(l util.Logger) {
	srv.logger = l
}

// EOF