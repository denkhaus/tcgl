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
	"code.google.com/p/tcgl/applog"
	"code.google.com/p/tcgl/asserts"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

//--------------------
// HELPER FUNCTIONS
//--------------------

// Start the internal test server.
func startTestServer() *httptest.Server {
	lazyCreateServer()
	prepareServer("", "")
	return httptest.NewServer(http.HandlerFunc(handleFunc))
}

// Type for the header.
type Hdr map[string]string

// Perform any local request.
func localDo(method string, ts *httptest.Server, path string, hdr Hdr, body []byte) ([]byte, error) {
	// First prepare it.
	tr := &http.Transport{}
	c := &http.Client{Transport: tr}
	url := ts.URL + path
	var bodyReader io.Reader
	if body != nil {
		bodyReader = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	// Now do it.
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return respBody, err
}

//--------------------
// TEST HANDLER
//--------------------

const TEST_TMPL_XML = `
<?xml version="1.0" encoding="UTF-8"?>
<test>
<domain>{{.domain}}</domain>
<resource>{{.resource}}</resource>
<resourceId>{{.resourceId}}</resourceId>
</test>
`

const TEST_TMPL_HTML = `
<?DOCTYPE html?>
<html>
<head><title>Test</title></head>
<body>
<dl>
<dt>Domain</dt><dd>{{.domain}}</dd>
<dt>Resource</dt><dd>{{.resource}}</dd>
<dt>Resource Id</dt><dd>{{.resourceId}}</dd>
</dl>
</body>
</html>
`

type TestData struct {
	Id    string
	Count int64
}

type TestHandler struct{}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (th *TestHandler) Init(domain, resource string) {
	ParseTemplate("test:context:xml", TEST_TMPL_XML, "application/xml")
	ParseTemplate("test:context:html", TEST_TMPL_HTML, "text/html")
}

func (th *TestHandler) Get(ctx *Context) bool {
	data := map[string]string{
		"domain":     ctx.Domain,
		"resource":   ctx.Resource,
		"resourceId": ctx.ResourceId,
	}
	switch {
	case ctx.AcceptsXML():
		applog.Debugf("get XML")
		ctx.RenderTemplate("test:context:xml", data)
	case ctx.AcceptsJSON():
		applog.Debugf("get JSON")
		ctx.MarshalJSON(data, true)
	default:
		applog.Debugf("get HTML")
		ctx.RenderTemplate("test:context:html", data)
	}
	return true
}

func (th *TestHandler) Put(ctx *Context) bool {
	data, err := ctx.GenericUnmarshalJSON()
	if err != nil {
		ctx.MarshalJSON(err, true)
	} else {
		ctx.MarshalJSON(data, true)
	}
	return true
}

func (th *TestHandler) Post(ctx *Context) bool {
	var data TestData
	err := ctx.UnmarshalGob(&data)
	if err != nil {
		ctx.MarshalGob(err)
	} else {
		ctx.MarshalGob(data)
	}
	return true
}

func (th *TestHandler) Delete(ctx *Context) bool {
	return false
}

//--------------------
// TESTS
//--------------------

// Test the GET command with an XML result.
func TestGetXML(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("test", "getxml", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/getxml/4711", Hdr{"Accept": "application/xml"}, nil)
	assert.Nil(err, "Local XML GET.")
	assert.Containment(string(body), "<resourceId>4711</resourceId>", "XML result.")
}

// Test the GET command with a JSON result.
func TestGetJSON(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("test", "getjson", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/getjson/4711", Hdr{"Accept": "application/json"}, nil)
	assert.Nil(err, "Local JSON GET.")
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	assert.Nil(err, "Unmarshal of the JSON data.")
	assert.Equal(data["resourceId"], "4711", "Unmarshaled JSON result.")
}

// Test the PUT command with a JSON payload and result.
func TestPutJSON(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("test", "putjson", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	inData := map[string]interface{}{"alpha": "foo", "beta": 4711.0, "gamma": true}
	b, _ := json.Marshal(inData)
	body, err := localDo("PUT", ts, "/test/putjson/4711", Hdr{"Content-Type": "application/json", "Accept": "application/json"}, b)
	assert.Nil(err, "Local JSON PUT.")
	outData := map[string]interface{}{}
	err = json.Unmarshal(body, &outData)
	assert.Nil(err, "Unmarshal of the JSON data.")
	for k, v := range inData {
		assert.Equal(outData[k], v, "JSON value")
	}
}

// Test the PUT command with a GOB payload and result.
func TestPutGob(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("test", "putgob", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	inData := TestData{"test", 4711}
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(inData)
	assert.Nil(err, "GOB encode.")
	body, err := localDo("POST", ts, "/test/putgob", Hdr{"Content-Type": "application/vnd.tideland.rwf"}, b.Bytes())
	assert.Nil(err, "Local GOB POST.")
	var outData TestData
	err = gob.NewDecoder(bytes.NewBuffer(body)).Decode(&outData)
	assert.Nil(err, "GOB decode.")
	assert.Equal(outData.Id, "test", "GOB decoded 'id'.")
	assert.Equal(outData.Count, int64(4711), "GOB decoded 'count'.")
}

// Test the redirection to default.
func TestRedirectDefault(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("default", "default", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/x/y", Hdr{}, nil)
	assert.Nil(err, "Local unknown GET for redirect.")
	assert.Containment(string(body), "<dd>default</dd>", "XML result.")
}

// Test the wrapper handler.
func TestWrapperHandler(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Prepare the server.
	AddResourceHandler("test", "wrapper", NewWrapperHandler(http.NotFound))
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/wrapper", Hdr{}, nil)
	assert.Nil(err, "Local wrapper GET.")
	assert.Equal(string(body), "404 page not found\n", "Wrapper result.")
}

// EOF