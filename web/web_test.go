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
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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
		ctx.Debugf("get XML")
		ctx.RenderTemplate("test:context:xml", data)
	case ctx.AcceptsJSON():
		ctx.Debugf("get JSON")
		ctx.MarshalJSON(data, true)
	default:
		ctx.Debugf("get HTML")
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
	// Prepare the server.
	AddResourceHandler("test", "getxml", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/getxml/4711", Hdr{"Accept": "application/xml"}, nil)
	if err != nil {
		t.Fatalf("XML get error: %v", err)
	}
	xml := string(body)
	if !strings.Contains(xml, "<resourceId>4711</resourceId>") {
		t.Fatalf("XML contains error: %v", xml)
	}
}

// Test the GET command with a JSON result.
func TestGetJSON(t *testing.T) {
	// Prepare the server.
	AddResourceHandler("test", "getjson", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/getjson/4711", Hdr{"Accept": "application/json"}, nil)
	if err != nil {
		t.Fatalf("JSON get error: %v", err)
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if data["resourceId"] != "4711" {
		t.Fatalf("JSON contains error: %s / %v", body, data)
	}
}

// Test the PUT command with a JSON payload and result.
func TestPutJSON(t *testing.T) {
	// Prepare the server.
	AddResourceHandler("test", "putjson", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	inData := map[string]interface{}{"alpha": "foo", "beta": 4711, "gamma": true}
	b, _ := json.Marshal(inData)
	body, err := localDo("PUT", ts, "/test/putjson/4711", Hdr{"Content-Type": "application/json", "Accept": "application/json"}, b)
	if err != nil {
		t.Fatalf("JSON local put error: %v", err)
	}
	outData := map[string]interface{}{}
	err = json.Unmarshal(body, &outData)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v (%s)", err, body)
	}
	for k, v := range inData {
		ov := outData[k]
		switch tov := ov.(type) {
		case string:
			if tov != v {
				t.Fatalf("JSON compare string error: %s = %v (%v)", k, v, tov)
			}
		case float64:
			if tov != float64(v.(int)) {
				t.Fatalf("JSON compare number error: %s = %v (%v)", k, v, tov)
			}
		case bool:
			if tov != v {
				t.Fatalf("JSON compare bool error: %s = %v (%v)", k, v, tov)
			}
		default:
			t.Fatalf("JSON invalid type error: %s is %T, has to be %T", k, v, tov)
		}
	}
}

// Test the PUT command with a GOB payload and result.
func TestPutGob(t *testing.T) {
	// Prepare the server.
	AddResourceHandler("test", "putgob", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	inData := TestData{"test", 4711}
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(inData)
	if err != nil {
		t.Fatalf("GOB encode error: %v", err)
	}
	body, err := localDo("POST", ts, "/test/putgob", Hdr{"Content-Type": "application/vnd.tideland.rwf"}, b.Bytes())
	if err != nil {
		t.Fatalf("GOB local post error: %v", err)
	}
	var outData TestData
	err = gob.NewDecoder(bytes.NewBuffer(body)).Decode(&outData)
	if err != nil {
		t.Fatalf("GOB decode error: %v", err)
	}
	if outData.Id != "test" || outData.Count != 4711 {
		t.Fatalf("GOB contains error: %s", outData)
	}
}

// Test the redirection to default.
func TestRedirectDefault(t *testing.T) {
	// Prepare the server.
	AddResourceHandler("default", "default", NewTestHandler())
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/x/y", Hdr{}, nil)
	if err != nil {
		t.Fatalf("Redirect get error: %v", err)
	}
	xml := string(body)
	if !strings.Contains(xml, "<dd>default</dd>") {
		t.Fatalf("Redirect contains error: %v", xml)
	}
}

// Test the wrapper handler.
func TestWrapperHandler(t *testing.T) {
	// Prepare the server.
	AddResourceHandler("test", "wrapper", NewWrapperHandler(http.NotFound))
	ts := startTestServer()
	// Now the request.
	body, err := localDo("GET", ts, "/test/wrapper", Hdr{}, nil)
	if err != nil {
		t.Fatalf("Wrapper get error: %v", err)
	}
	if string(body) != "404 page not found\n" {
		t.Fatalf("Wrapper content error: '%s'", body)
	}
}

// EOF