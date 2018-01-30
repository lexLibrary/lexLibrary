// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"

	"github.com/lexLibrary/lexLibrary/data"
)

var sqlFirstRunCheck = data.NewQuery(`select count(*) from users`)

var firstRunTrigger func()

func firstRunCheck() error {
	count := 0
	err := sqlFirstRunCheck.QueryRow().Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		firstRunTrigger()
	}

	return nil
}

// FirstRunTrigger gets triggered if this is the first time Lex Library has been run
func FirstRunTrigger(fn func()) {
	firstRunTrigger = fn
}

// FirstRunSetup creates the first admin and sets Lex Library's settings
func FirstRunSetup(username, password string, settings map[string]interface{}) (*User, error) {
	var user *User

	count := 0
	err := sqlFirstRunCheck.QueryRow().Scan(&count)

	if err != nil {
		return nil, err
	}

	if count != 0 {
		return nil, NewFailure("The First Run setup has already been run, and cannot be run again")
	}

	err = data.BeginTx(func(tx *sql.Tx) error {
		u, err := userNew(tx, username, password)
		if err != nil {
			return err
		}

		for k, v := range settings {
			err = SettingSet(k, v)
			if err != nil {
				return err
			}
		}

		user = u
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
