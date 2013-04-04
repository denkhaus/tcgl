// Tideland Common Go Library - Redis
//
// Copyright (C) 2009-2013 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// A simple but powerful client for accessing the Redis database.
//
// After establishing a connection using NewRedisDatabase() commands can be
// executed with Command(). So every command of Redis is possible. The method
// returns a ResultSet with different methods for success testing and access
// to the retrieved values. The method MultiCommand() can be used for
// transactions. The passed function gets a MultiCommand instance as
// argument for calling the inner Command() methods.
package redis

// EOF
