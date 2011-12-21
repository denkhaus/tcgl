// Tideland Common Go Library - Web
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

// The web package provides a framework for a component based web development.
//
// It is intended as a convenience to build web applications and servers following 
// the principles of REST. Internally it uses the standard http, template, json and xml
// packages. The business logic has to be implemented in components that fullfill the
// individual handler interfaces. They work on a context with some helpers but also
// have got access to the original Request and ResponseWriter arguments.
package web

// EOF
