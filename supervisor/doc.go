// Tideland Common Go Library - Supervisor
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

// A supervisor controls the propoer execution of goroutines.
//
// Depending on the configured strategy a terminated goroutine
// by error or panic will be restarted up to a configured 
// restart frequency. With the strategy OneForOne it will be
// only the one goroutine, in case of one for all all goroutines
// will be terminated by sending a signal to them and then 
// restarted. If the restart frequency is exceeded the whole
// supervisor panics.
package supervisor

// EOF
