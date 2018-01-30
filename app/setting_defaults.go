// Copyright (c) 2017-2018 Townsourced Inc.

package app

// settingDefaults are the default settings that Lex Library starts with.  If a setting is not overridden in the database
// then the default values here are used
var settingDefaults = []Setting{
	Setting{
		ID:          "AllowPublicDocuments",
		Description: "Whether or not to allow documents to be published that are accessible without logging in to Lex Library.",
		Value:       true,
	},
	Setting{
		ID:          "AuthenticationType",
		Description: "How users log into Lex Library.", // TODO add description of auth options into setting description
		Value:       AuthTypePassword,
		Options: []interface{}{
			AuthTypePassword,
		},
	},
	Setting{
		ID:          "PasswordMinLength",
		Description: "Required minimum length for passwords.",
		Value:       10,
		validate: func(value interface{}) error {
			if value.(int) < passwordMinLength {
				return NewFailure("Minimum password length must be greater than %d", passwordMinLength)
			}
			return nil
		},
	},
	Setting{
		ID:          "BadPasswordCheck",
		Description: "Don't allow passwords that exist in the top 10,000 most common passwords list.",
		Value:       true,
	},
	Setting{
		ID:          "PasswordRequireSpecial",
		Description: "Require that all new passwords have at least one special character.",
		Value:       false,
	},
	Setting{
		ID:          "PasswordRequireNumber",
		Description: "Require that all new passwords have at least one number.",
		Value:       false,
	},
	Setting{
		ID:          "PasswordRequireMixedCase",
		Description: "Require that all new passwords have at least one upper and one lower case letter.",
		Value:       false,
	},
	Setting{
		ID:          "RateLimit",
		Description: "Number of requests per minute a unique user or ip address is allowed to make.",
		Value:       2000,
	},
	Setting{
		ID:          "RememberSessionDays",
		Description: "How many days a session is valid for if Remember Me is checked when logging in.",
		Value:       15,
		validate: func(value interface{}) error {
			if value.(int) > sessionMaxDaysRemembered {
				return NewFailure("The maximium number of days a session can be remembered for is %d",
					sessionMaxDaysRemembered)
			}
			return nil
		},
	},
	Setting{
		ID: "AllowRuntimeInfoInIssues",
		Description: "Whether or not to allow runtime information (OS, CPU, architecture, etc) to be included " +
			"when a user submits an issue via the 'about' page. If false, admins can still access this information.",
		Value: true,
	},
	Setting{
		ID: "AllowPublicSignups",
		Description: "Whether or not to allow anyone to create a login for Lex Library.  If false, they'll only " +
			"be able to sign up via a link generated by an Admin.",
		Value: false,
	},
}
