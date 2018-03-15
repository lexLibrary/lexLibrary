// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/rs/xid"
)

func TestUser(t *testing.T) {
	var admin *app.User
	reset := func(t *testing.T) {
		t.Helper()

		_, err := data.NewQuery("delete from sessions").Exec()
		if err != nil {
			t.Fatalf("Error emptying sessions table before running tests: %s", err)
		}

		_, err = data.NewQuery("delete from users").Exec()
		if err != nil {
			t.Fatalf("Error emptying users table before running tests: %s", err)
		}
		_, err = data.NewQuery("delete from settings").Exec()
		if err != nil {
			t.Fatalf("Error emptying settings table before running tests: %s", err)
		}
		admin, err = app.FirstRunSetup("admin", "adminpassword")
		if err != nil {
			t.Fatalf("Error setting up admin user: %s", err)
		}

		err = admin.AsAdmin().SetSetting("AllowPublicSignups", true)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
	}

	t.Run("New", func(t *testing.T) {
		reset(t)
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

		var oID xid.ID
		var oUsername string
		var oPassword []byte
		var oPasswordVersion int
		var oAuthType string
		var oActive bool
		var oVersion int
		var oUpdated time.Time
		var oCreated time.Time

		err = data.NewQuery(`
			select 	id, 
					username, 
					password, 
					password_version,
					auth_type,
					active,
					version,
					updated,
					created
			from users
			where id = {{arg "id"}}`).QueryRow(sql.Named("id", u.ID)).Scan(
			&oID,
			&oUsername,
			&oPassword,
			&oPasswordVersion,
			&oAuthType,
			&oActive,
			&oVersion,
			&oUpdated,
			&oCreated,
		)
		if err != nil {
			t.Fatalf("Error retrieving inserted user: %s", err)
		}

		if len(oID) != 12 {
			t.Fatalf("User ID incorrect length. Expected %d got %d", 12, len(oID))
		}

		if oUsername != username {
			t.Fatalf("Username not set properly expected %s, got %s", username, oUsername)
		}
		if oPassword == nil {
			t.Fatalf("Password not set properly")
		}

		if oPasswordVersion < 0 {
			t.Fatalf("Invalid password version")
		}

		if oAuthType != app.AuthTypePassword {
			t.Fatalf("Invalid Auth Type.  Expected %s, got %s", app.AuthTypePassword, oAuthType)
		}

		if !oActive {
			t.Fatalf("Newly created user was not marked as active")
		}

		if oVersion != 0 {
			t.Fatalf("Incorrect new user version. Expected %d, got %d", 0, oVersion)
		}

		if !oUpdated.Before(time.Now()) {
			t.Fatalf("Incorrect Updated date: %v", oUpdated)
		}
		if !oCreated.Before(time.Now()) {
			t.Fatalf("Incorrect Created date: %v", oCreated)
		}
		if oCreated.After(oUpdated) {
			t.Fatalf("User created data was after user updated date. Created %v Updated %v", oCreated,
				oUpdated)
		}
	})

	t.Run("Invalid Name", func(t *testing.T) {
		reset(t)
		fullName := fmt.Sprintf("%70s", "full name")

		u, err := app.UserNew("testusername", "ODSjflaksjdfhiasfd323")
		if err != nil {
			t.Fatalf("Error adding user")
		}

		err = u.SetFullName(fullName, u.Version)
		if err == nil {
			t.Fatalf("No error adding too long full name")
		}
		if !app.IsFail(err) {
			t.Fatalf("Error on too long full name is not a failure")
		}
	})

	t.Run("Invalid Username", func(t *testing.T) {
		reset(t)
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
		reset(t)
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
		reset(t)
		err := admin.AsAdmin().SetSetting("BadPasswordCheck", true)
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
		reset(t)
		err := admin.AsAdmin().SetSetting("PasswordRequireSpecial", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = admin.AsAdmin().SetSetting("BadPasswordCheck", false)
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
		reset(t)
		err := admin.AsAdmin().SetSetting("PasswordRequireNumber", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = admin.AsAdmin().SetSetting("BadPasswordCheck", false)
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
		reset(t)
		err := admin.AsAdmin().SetSetting("PasswordRequireMixedCase", true)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = admin.AsAdmin().SetSetting("BadPasswordCheck", false)
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
		reset(t)
		err := admin.AsAdmin().SetSetting("PasswordMinLength", 8)
		if err != nil {
			t.Fatalf("Error updating setting")
		}

		err = admin.AsAdmin().SetSetting("BadPasswordCheck", false)
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

		err = admin.AsAdmin().SetSetting("PasswordMinLength", 50)
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
		reset(t)
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

		err = other.AsAdmin().SetUserActive(u, false, u.Version)
		if err == nil {
			t.Fatalf("Setting active from other user did not fail")
		}

		err = admin.AsAdmin().SetUserActive(u, false, u.Version)
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
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user for SetName testing")
		}

		fName := "firstname lastname"

		err = u.SetFullName(fName, u.Version)
		if err != nil {
			t.Fatalf("Error setting name: %s", err)
		}

		if u.FullName != fName {
			t.Fatalf("User name was not updated")
		}
	})
	t.Run("UserGet", func(t *testing.T) {
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user")
		}

		got, err := app.UserGet(u.Username)
		if err != nil {
			t.Fatalf("Error getting user: %s", err)
		}

		if u.FullName != got.FullName || u.ID != got.ID ||
			u.Username != got.Username {
			t.Fatalf("Retrieved user does not match.  Wanted %v, got %v", u, got)
		}

	})
	t.Run("SetAdmin", func(t *testing.T) {
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user for SetAdmin testing")
		}

		other, err := app.UserNew("othertestuser", "reallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error adding other user for SetAdmin testing")
		}

		err = other.AsAdmin().SetUserAdmin(u, true, u.Version)
		if err == nil {
			t.Fatalf("Setting admin from other user did not fail")
		}

		err = u.AsAdmin().SetUserAdmin(u, true, u.Version)
		if err == nil {
			t.Fatalf("Setting admin from non-admin self did not fail")
		}

		other.Admin = true // this seems like cheating

		err = other.AsAdmin().SetUserAdmin(u, true, u.Version)
		if err != nil {
			t.Fatalf("Error setting admin by another admin: %s", err)
		}
	})

	t.Run("Public Signups Disabled", func(t *testing.T) {
		reset(t)
		err := admin.AsAdmin().SetSetting("AllowPublicSignups", false)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
		username := "testuser"
		password := "reallygoodlongpassword"

		_, err = app.UserNew(username, password)
		if err == nil {
			t.Fatalf("No error was returned when AllowPublicSignups is false")
		}
	})

	t.Run("Versions", func(t *testing.T) {
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user for SetName testing")
		}

		if u.Version != 0 {
			t.Fatalf("Incorrect first version of the user record. Got %d, wanted %d", u.Version, 0)
		}

		// get copy of current user version
		old := *u

		err = u.SetFullName("version one", u.Version)
		if err != nil {
			t.Fatalf("Error setting name: %s", err)
		}

		if u.Version != 1 {
			t.Fatalf("Incorrect first version of the user record. Got %d, wanted %d", u.Version, 1)
		}

		err = old.SetFullName("version old", old.Version)
		if err != app.ErrUserConflict {
			t.Fatalf("Updating an older version of a user did not return a Conflict")
		}

	})

	t.Run("SetPassword", func(t *testing.T) {
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user")
		}

		oldSession, err := app.SessionNew(u, time.Now().AddDate(0, 0, 10), "", "")
		if err != nil {
			t.Fatalf("Error creating new session: %s", err)
		}

		err = u.SetPassword(password, password, u.Version)
		if err == nil {
			t.Fatalf("Setting password to the same password did not return an error")
		}

		err = u.SetPassword(password, "newreallygoodlongpassword", u.Version)
		if err != nil {
			t.Fatalf("Error setting password: %s", err)
		}

		_, err = app.SessionGet(oldSession.UserID, oldSession.ID)
		if err != app.ErrSessionInvalid {
			t.Fatalf("Old session was not exired when changing passwords")
		}

	})
	t.Run("UserSetExpiredPassword", func(t *testing.T) {
		reset(t)
		username := "testuser"
		password := "reallygoodlongpassword"

		u, err := app.UserNew(username, password)
		if err != nil {
			t.Fatalf("Error adding user")
		}

		oldSession, err := app.SessionNew(u, time.Now().AddDate(0, 0, 10), "", "")
		if err != nil {
			t.Fatalf("Error creating new session: %s", err)
		}

		_, err = app.UserSetExpiredPassword(u.Username, password, password)
		if err == nil {
			t.Fatalf("Setting password to the same password did not return an error")
		}

		newu, err := app.UserSetExpiredPassword(u.Username, password, "newreallygoodlongpassword")
		if err != nil {
			t.Fatalf("Error setting password: %s", err)
		}

		if newu.ID != u.ID {
			t.Fatalf("Invalid user returned from SetExpiredPassword. Wanted %s, got %s", u.ID, newu.ID)
		}

		_, err = app.SessionGet(oldSession.UserID, oldSession.ID)
		if err != app.ErrSessionInvalid {
			t.Fatalf("Old session was not exired when changing passwords")
		}
	})

	//TODO: Profile Image tests, use testdata dir

	reset(t)
}
