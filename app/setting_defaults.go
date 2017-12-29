// Copyright (c) 2017 Townsourced Inc.

package app

// settingDefaults are the default settings that Lex Library starts with.  If a setting is not overridden in the database
// then the default values here are used
var settingDefaults = []Setting{
	Setting{
		ID:          "AllowPublic",
		Category:    "Documents",
		Description: "Whether or not to allow documents to be published that are accessible without logging in to Lex Library",
		Value:       true,
	},
	Setting{
		ID:          "AuthenticationType",
		Category:    "Users",
		Description: "How users log into Lex Library.", // TODO add description of auth options into setting description
		Value:       AuthTypePassword,
		Options: []interface{}{
			AuthTypePassword,
		},
	},
	Setting{
		ID:          "PasswordMinimumLength",
		Category:    "Users",
		Description: "Minimum required lenth for passwords",
		Value:       10,
	},
}
