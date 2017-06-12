// Copyright (c) 2017 Townsourced Inc.

package data

import "fmt"

func ensureSchema(forceRollback bool) error {
	// look up current schema version
	// if current schemaVer > database schemaVer
	// run the next schema version script
	// if current schemaVer < database schemaVer
	// return error unless force rollback
	return fmt.Errorf("TODO")
}
