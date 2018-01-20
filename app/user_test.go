// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestUser(t *testing.T) {
	reset := func() {
		t.Helper()
		_, err := data.NewQuery("delete from users").Exec()
		if err != nil {
			t.Fatalf("Error emptying users table before running tests: %s", err)
		}
		_, err = data.NewQuery("delete from settings").Exec()
		if err != nil {
			t.Fatalf("Error emptying settings table before running tests: %s", err)
		}

	}

	t.Run("New", func(t *testing.T) {
		reset()
		username := "newusęr"

		u, err := app.UserNew(username, "ODSjflaksjdfhiasfd323")
		if err != nil {
			t.Fatalf("Error adding new user: %s", err)
		}

		// sleep for one second because that's the minimum precision of some database's datetime fields
		time.Sleep(1 * time.Second)

		if u.Username != username {
			t.Fatalf("Returned user doesn't match passed in values")
		}

		other := &app.User{}

		err = data.NewQuery(`
			select 	id, 
					username, 
					first_name, 
					last_name, 
					password, 
					password_version,
					auth_type,
					active,
					version,
					updated,
					created
			from users
			where id = {{arg "id"}}`).QueryRow(sql.Named("id", u.ID)).Scan(
			&other.ID,
			&other.Username,
			&other.FirstName,
			&other.LastName,
			&other.Password,
			&other.PasswordVersion,
			&other.AuthType,
			&other.Active,
			&other.Version,
			&other.Updated,
			&other.Created,
		)
		if err != nil {
			t.Fatalf("Error retrieving inserted user: %s", err)
		}

		if len(other.ID) != 12 {
			t.Fatalf("User ID incorrect length. Expected %d got %d", 12, len(other.ID))
		}

		if other.Username != username {
			t.Fatalf("Username not set properly expected %s, got %s", username, other.Username)
		}
		if other.Password == nil {
			t.Fatalf("Password not set properly")
		}

		if other.PasswordVersion < 0 {
			t.Fatalf("Invalid password version")
		}

		if other.AuthType != app.AuthTypePassword {
			t.Fatalf("Invalid Auth Type.  Expected %s, got %s", app.AuthTypePassword, other.AuthType)
		}

		if !other.Active {
			t.Fatalf("Newly created user was not marked as active")
		}

		if other.Version != 0 {
			t.Fatalf("Incorrect new user version. Expected %d, got %d", 0, other.Version)
		}

		if !other.Updated.Before(time.Now()) {
			t.Fatalf("Incorrect Updated date: %v", other.Updated)
		}
		if !other.Created.Before(time.Now()) {
			t.Fatalf("Incorrect Created date: %v", other.Created)
		}
		if other.Created.After(other.Updated) {
			t.Fatalf("User created data was after user updated date. Created %v Updated %v", other.Created,
				other.Updated)
		}
	})

	t.Run("Invalid Name", func(t *testing.T) {
		reset()
		firstname := fmt.Sprintf("%70s", "firstname")
		lastname := fmt.Sprintf("%70s", "firstname")

		u, err := app.UserNew("testusername", "ODSjflaksjdfhiasfd323")

		err = u.SetName(firstname, "", u)
		if err == nil {
			t.Fatalf("No error adding too long first name")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on too long first name is not a failure")
		}

		err = u.SetName("", lastname, u)
		if err == nil {
			t.Fatalf("No error adding too long last name")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on too long last name is not a failure")
		}
	})

	t.Run("Invalid Username", func(t *testing.T) {
		reset()
		_, err := app.UserNew("", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding user without a username")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on empty username is not a failure")
		}
		_, err = app.UserNew("username with space", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding username with a space")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on username with a space is not a failure")
		}
		_, err = app.UserNew("username_with_underscores", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding username with underscores")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on username with underscores is not a failure")
		}

	})

	t.Run("Duplicate Username", func(t *testing.T) {
		reset()
		existing, err := app.UserNew("existing", "ODSjflaksjdfhiasfd323")
		if err != nil {
			t.Fatalf("Error adding existing user: %s", err)
		}

		_, err = app.UserNew(existing.Username, "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error when adding a duplicate user")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on duplicate user is not a failure")
		}

		_, err = app.UserNew(strings.ToUpper(existing.Username), "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error when adding a duplicate user with different case")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on duplicate user with different case is not a failure")
		}
	})

	t.Run("Common Password", func(t *testing.T) {
		reset()
		err := app.SettingSet("BadPasswordCheck", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}
		_, err = app.UserNew("testuser", "123456qwerty")
		if err == nil {
			t.Fatalf("No error when using a common password")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on common password is not a failure")
		}
	})
	t.Run("Password Special", func(t *testing.T) {
		reset()
		err := app.SettingSet("PasswordRequireSpecial", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = app.SettingSet("BadPasswordCheck", false)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpasswordwithoutaspecialchar")
		if err == nil {
			t.Fatalf("No error when using a password without a special character")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without a special character is not a failure")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpasswordwithaspecialchar#")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}
	})

	t.Run("Password Number", func(t *testing.T) {
		reset()
		err := app.SettingSet("PasswordRequireNumber", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = app.SettingSet("BadPasswordCheck", false)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a password without a number")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without a number is not a failure")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpasswordwithanumber3")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}
	})
	t.Run("Password Mixed Case", func(t *testing.T) {
		reset()
		err := app.SettingSet("PasswordRequireMixedCase", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = app.SettingSet("BadPasswordCheck", false)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a password without mixed case")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without mixed case is not a failure")
		}

		_, err = app.UserNew("testuser", "REALLYGOODLONGPASSWORD")
		if err == nil {
			t.Fatalf("No error when using a password without mixed case")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without mixed case is not a failure")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpasswordwithMixedCase")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}
	})
	t.Run("Password Length", func(t *testing.T) {
		reset()
		err := app.SettingSet("PasswordMinLength", 8)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = app.SettingSet("BadPasswordCheck", false)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "short")
		if err == nil {
			t.Fatalf("No error when using a short password")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on short password is not a failure")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}

		err = app.SettingSet("PasswordMinLength", 50)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a short password")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on short password is not a failure")
		}
	})
	t.Run("SetActive", func(t *testing.T) {
		reset()
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user for SetActive testing")
		}

		other, err := app.UserNew("othertestuser", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding other user for SetActive testing")
		}

		err = u.SetActive(false, other)
		if err == nil {
			t.Fatalf("Setting active from other user did not fail")
		}

		err = u.SetActive(false, u)
		if err != nil {
			t.Fatalf("Error setting active to false: %s", err)
		}

		if u.Active {
			t.Fatalf("User was not inactive")
		}

		_, err = app.Login(username, password)
		if err != app.ErrLogonFailure {
			t.Fatalf("No logon failure error when logging in with an inactive user")
		}

	})
	t.Run("SetName", func(t *testing.T) {
		reset()
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user for SetName testing")
		}

		other, err := app.UserNew("othertestuser", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding other user for SetName testing")
		}

		fName := "firstname"
		lName := "lastname"

		err = u.SetName(fName, lName, other)
		if err == nil {
			t.Fatalf("Setting active from other user did not fail")
		}

		err = u.SetName(fName, lName, u)
		if err != nil {
			t.Fatalf("Error setting name: %s", err)
		}

		if u.FirstName != fName || u.LastName != lName {
			t.Fatalf("User name was not updated")
		}
	})
	t.Run("UserGet", func(t *testing.T) {
		reset()
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user")
		}

		other, err := app.UserNew("othertestuser", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding other user")
		}

		got, err := app.UserGet(u.Username, other)
		if err != nil {
			t.Fatalf("Error getting user: %s", err)
		}

		if got.Password != nil || got.PasswordVersion != 0 || got.Version != 0 || got.AuthType != "" {
			t.Fatalf("Getting user from other user returned private data")
		}

		got, err = app.UserGet(u.Username, u)
		if err != nil {
			t.Fatalf("Error getting user: %s", err)
		}

		if got.Password != nil || got.PasswordVersion != 0 {
			t.Fatalf("Exported UserGet call returned password data")
		}

		if u.FirstName != got.FirstName || u.LastName != got.LastName || u.ID != got.ID ||
			u.Username != got.Username {
			t.Fatalf("Retrieved user does not match.  Wanted %v, got %v", u, got)
		}

	})

	reset()
}
