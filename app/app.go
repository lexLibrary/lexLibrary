// Copyright (c) 2017-2018 Townsourced Inc.

// Package app handles the application logic for Lex Library
// All rules and logic that apply to application structures should happen in this library
// Transactions should all be self contained in this library, and not be initiated in the Web layer
// No web structures or packages (http, cookies, etc) should show up in this package
package app

const maxRows = 10000

// Init initializes the application layer
func Init() error {
	err := settingTriggerInit()
	if err != nil {
		return err
	}

	err = loadVersion()

	if err != nil {
		return err
	}

	err = firstRunCheck()
	return err
}
