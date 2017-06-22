// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"gitlab.com/lexLibrary/lexLibrary/app"
)

func initApp() error {
	// open sqlite db in memory for testing
	// TODO: Allow for running tests against any of the database backends for use with containerized test envs
	cfg := app.Config{
		DatabaseType:       "sqlite",
		DatabaseURL:        "file::memory:?mode=memory&cache=shared",
		MaxIdleConnections: 1,
		MaxOpenConnections: 1,
	}

	return app.Init(cfg)
}
