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
		firstname := "firstname"
		lastname := "lastname"

		u, err := app.UserNew(username, firstname, lastname, "ODSjflaksjdfhiasfd323")
		if err != nil {
			t.Fatalf("Error adding new user: %s", err)
		}

		// sleep for one second because that's the minimum precision of some database's datetime fields
		time.Sleep(1 * time.Second)

		if u.FirstName != firstname || u.LastName != lastname || u.Username != username {
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

		if other.FirstName != firstname {
			t.Fatalf("First Name not set properly expected %s, got %s", firstname, other.FirstName)
		}
		if other.LastName != lastname {
			t.Fatalf("Last Name not set properly expected %s, got %s", lastname, other.LastName)
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

		_, err := app.UserNew("testusername", firstname, "", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding too long first name")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on too long first name is not a failure")
		}

		_, err = app.UserNew("testusername", "", lastname, "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding too long last name")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on too long last name is not a failure")
		}
	})

	t.Run("Invalid Username", func(t *testing.T) {
		reset()
		_, err := app.UserNew("", "", "", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding user without a username")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on empty username is not a failure")
		}
		_, err = app.UserNew("username with space", "", "", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding username with a space")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on username with a space is not a failure")
		}
		_, err = app.UserNew("username_with_underscores", "", "", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error adding username with underscores")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on username with underscores is not a failure")
		}

	})

	t.Run("Duplicate Username", func(t *testing.T) {
		reset()
		existing, err := app.UserNew("existing", "", "", "ODSjflaksjdfhiasfd323")
		if err != nil {
			t.Fatalf("Error adding existing user: %s", err)
		}

		_, err = app.UserNew(existing.Username, "", "", "ODSjflaksjdfhiasfd323")
		if err == nil {
			t.Fatalf("No error when adding a duplicate user")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on duplicate user is not a failure")
		}

		_, err = app.UserNew(strings.ToUpper(existing.Username), "", "", "ODSjflaksjdfhiasfd323")
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
		_, err = app.UserNew("testuser", "", "", "123456qwerty")
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

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpasswordwithoutaspecialchar")
		if err == nil {
			t.Fatalf("No error when using a password without a special character")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without a special character is not a failure")
		}

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpasswordwithaspecialchar#")
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

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a password without a number")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without a number is not a failure")
		}

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpasswordwithanumber3")
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

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a password without mixed case")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without mixed case is not a failure")
		}

		_, err = app.UserNew("testuser", "", "", "REALLYGOODLONGPASSWORD")
		if err == nil {
			t.Fatalf("No error when using a password without mixed case")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on password without mixed case is not a failure")
		}

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpasswordwithMixedCase")
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

		_, err = app.UserNew("testuser", "", "", "short")
		if err == nil {
			t.Fatalf("No error when using a short password")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on short password is not a failure")
		}

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}

		err = app.SettingSet("PasswordMinLength", 50)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		_, err = app.UserNew("testuser", "", "", "reallygoodlongpassword")
		if err == nil {
			t.Fatalf("No error when using a short password")
		}

		if !app.IsFail(err) {
			t.Fatalf("Error on short password is not a failure")
		}
	})
	t.Run("SetActive", func(t *testing.T) {
		//TODO:
	})

	reset()
}
