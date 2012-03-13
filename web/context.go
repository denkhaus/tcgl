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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//--------------------
// ENVELOPE
//--------------------

// Envelope is a helper to give a qualified feedback in RESTful requests.
// It contains wether the request has been successful, in case of an
// error an additional message and the payload.
type Envelope struct {
	Success bool
	Message string
	Payload interface{}
}

//--------------------
// CONTEXT
//--------------------

// Context encapsulates all needed data for handling a request.
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Domain         string
	Resource       string
	ResourceId     string
}

// Creates a new context.
func newContext(rw http.ResponseWriter, r *http.Request) *Context {
	// Init the context.
	ctx := &Context{
		ResponseWriter: rw,
		Request:        r,
	}
	// Split path for REST identifiers.
	parts := strings.Split(r.URL.Path[len(srv.basePath):], "/")
	switch len(parts) {
	case 3:
		ctx.ResourceId = parts[2]
		ctx.Resource = parts[1]
		ctx.Domain = parts[0]
	case 2:
		ctx.Resource = parts[1]
		ctx.Domain = parts[0]
	default:
		ctx.Resource = srv.defaultResource
		ctx.Domain = srv.defaultDomain
	}
	return ctx
}

// String returns domain, resource and resource id of the context.
func (ctx *Context) String() string {
	return fmt.Sprintf("%s /%s/%s/%s", ctx.Request.Method, ctx.Domain, ctx.Resource, ctx.ResourceId)
}

// Checks if the requestor accepts plain text as a content type.
func (ctx *Context) AcceptsPlain() bool {
	return strings.Contains(ctx.Request.Header.Get("Accept"), CT_PLAIN)
}

// Checks if the requestor accepts HTML as a content type.
func (ctx *Context) AcceptsHTML() bool {
	return strings.Contains(ctx.Request.Header.Get("Accept"), CT_HTML)
}

// Checks if the requestor accepts XML as a content type.
func (ctx *Context) AcceptsXML() bool {
	return strings.Contains(ctx.Request.Header.Get("Accept"), CT_XML)
}

// Checks if the requestor accepts JSON as a content type.
func (ctx *Context) AcceptsJSON() bool {
	return strings.Contains(ctx.Request.Header.Get("Accept"), CT_JSON)
}

// Redirect to a domain, resource and resource id (optional).
func (ctx *Context) Redirect(domain, resource, resourceId string) {
	url := srv.basePath + domain + "/" + resource
	if resourceId != "" {
		url = url + "/" + resourceId
	}
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(http.StatusMovedPermanently)
}

// RenderTemplate renders a template with the passed data to the response writer.
func (ctx *Context) RenderTemplate(templateId string, data interface{}) {
	srv.templateCache.render(ctx.ResponseWriter, templateId, data)
}

// MarshalJSON marshals the passed data to JSON and writes it to the response writer.
// The HTML flag controls the data encoding.
func (ctx *Context) MarshalJSON(data interface{}, html bool) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
	if html {
		var buf bytes.Buffer
		json.HTMLEscape(&buf, b)
		b = buf.Bytes()
	}
	ctx.ResponseWriter.Header().Set("Content-Type", CT_JSON)
	ctx.ResponseWriter.Write(b)
}

// PositiveJSONFeedback produces a positive feedback envelope
// encoded in JSON.
func (ctx *Context) PositiveJSONFeedback(m string, p interface{}, args ...interface{}) {
	rm := fmt.Sprintf(m, args...)
	ctx.MarshalJSON(&Envelope{true, rm, p}, true)
}

// NegativeJSONFeedback produces a negative feedback envelope
// encoded in JSON.
func (ctx *Context) NegativeJSONFeedback(m string, args ...interface{}) {
	rm := fmt.Sprintf(m, args...)
	ctx.MarshalJSON(&Envelope{false, rm, nil}, true)
}

// UnmarshalJSON checks if the request content type is JSON, reads its body
// and unmarshals it to the value pointed to by data.
func (ctx *Context) UnmarshalJSON(data interface{}) error {
	if ctx.Request.Header.Get("Content-Type") != CT_JSON {
		return errors.New("request content-type isn't application/json")
	}
	body, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Request.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &data)
}

// GenericUnmarshalJSON works like UnmarshalJSON but can be used if the transmitted
// type is unknown or has no Go representation. It will a mapping according to
// http://golang.org/pkg/json/#Unmarshal.
func (ctx *Context) GenericUnmarshalJSON() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	err := ctx.UnmarshalJSON(&data)
	return data, err
}

// MarshalGob marshals the passed data to GOB and writes it to the response writer.
func (ctx *Context) MarshalGob(data interface{}) {
	enc := gob.NewEncoder(ctx.ResponseWriter)
	ctx.ResponseWriter.Header().Set("Content-Type", CT_GOB)
	enc.Encode(data)
}

// UnmarshalGob checks if the request content type is GOB, reads its body
// and unmarshals it to the value pointed to by data.
func (ctx *Context) UnmarshalGob(data interface{}) error {
	if ctx.Request.Header.Get("Content-Type") != CT_GOB {
		return errors.New("request content-type isn't application/vnd.tideland.rwf")
	}
	dec := gob.NewDecoder(ctx.Request.Body)
	err := dec.Decode(data)
	ctx.Request.Body.Close()
	return err
}

// EOF
