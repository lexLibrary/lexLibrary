// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/rs/xid"
)

// User is a user login to Lex Library
type User struct {
	ID              xid.ID    `json:"id"`
	Username        string    `json:"username"`
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	AuthType        string    `json:"authType,omitempty"`
	Password        []byte    `json:"-"`
	PasswordVersion int       `json:"-"`
	Active          bool      `json:"active"`            // whether or not the user is active and can log in
	Version         int       `json:"version,omitempty"` // version of this record starting with 0
	Updated         time.Time `json:"updated,omitempty"`
	Created         time.Time `json:"created,omitempty"`
	Admin           bool      `json:"admin,omitempty"`
}

// AuthType determines the authentication method for a given user
const (
	AuthTypePassword = "password"
	//AuthType...
)

// user constants
const (
	UserMaxNameLength = 64 // This is pretty arbitrary but there should probably be some limit
)

var ErrUserNotFound = NotFound("User Not found")

var (
	sqlUserInsert = data.NewQuery(`insert into users (
		id,
		username, 
		first_name, 
		last_name, 
		auth_type,
		password,
		password_version,
		active,
		version,
		updated, 
		created,
		admin
	) values (
		{{arg "id"}}, 
		{{arg "username"}}, 
		{{arg "first_name"}}, 
		{{arg "last_name"}}, 
		{{arg "auth_type"}},
		{{arg "password"}},
		{{arg "password_version"}},
		{{arg "active"}},
		{{arg "version"}},
		{{arg "updated"}}, 
		{{arg "created"}},
		{{arg "admin"}}
	)`)
	sqlUserFromID = data.NewQuery(`
		select id, username, first_name, last_name, auth_type, active, version, updated, created, admin 
		from users where id = {{arg "id"}}`)
	sqlUserFromUsername = data.NewQuery(`
		select id, username, first_name, last_name, auth_type, password, password_version, active, version, updated, created, admin 
		from users where username = {{arg "username"}}`)
	sqlUserUpdateActive = data.NewQuery(`update users set active = {{arg "active"}} where id = {{arg "id"}}`)
	sqlUserUpdateAdmin  = data.NewQuery(`update users set admin = {{arg "admin"}} where id = {{arg "id"}}`)
	sqlUserUpdateName   = data.NewQuery(`update users set first_name = {{arg "first_name"}}, 
		last_name = {{arg "last_name"}} where id = {{arg "id"}}`)
)

// UserNew creates a new user, from the sign up page
// returns unauthorized if public signups are disabled
func UserNew(username, password string) (*User, error) {
	if !SettingMust("AllowPublicSignups").Bool() {
		return nil, Unauthorized("Public signups are currently disabled")
	}

	return userNew(nil, username, password)
}

// UserNewFromURL creates a new user if the passed in token is valid
// func UserNewFromURL(username, password, token string) (*User, error) {
// 	err := data.BeginTx(func(tx *sql.Tx) error {
// 		//TODO:
// 		return errors.New("TODO")
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return nil, nil
// }

func userNew(tx *sql.Tx, username, password string) (*User, error) {
	u := &User{
		ID:       xid.New(),
		Username: strings.ToLower(username),
		AuthType: AuthTypePassword,
		Active:   true,
		Version:  0,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	err := u.validate()
	if err != nil {
		return nil, err
	}

	err = ValidatePassword(password)
	if err != nil {
		return nil, err
	}

	passVer := len(passwordVersions) - 1

	hash, err := passwordVersions[passVer].hash(password)
	if err != nil {
		return nil, err
	}

	u.PasswordVersion = passVer
	u.Password = hash

	_, err = userGet(tx, u.Username)

	if err == nil {
		return nil, NewFailure("A user with the username %s already exists", u.Username)
	}

	if err != ErrUserNotFound {
		return nil, err
	}

	err = u.insert(tx)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// UserGet retrieves a user based on the passed in username
func UserGet(username string, who *User) (*User, error) {
	u, err := userGet(nil, username)
	if err != nil {
		return nil, err
	}

	if who == nil || who.ID != u.ID {
		u.clearPrivate()
	} else {
		u.clearPassword()
	}

	return u, nil
}

func userGet(tx *sql.Tx, username string) (*User, error) {
	u := &User{}

	err := sqlUserFromUsername.Tx(tx).QueryRow(sql.Named("username", username)).
		Scan(
			&u.ID,
			&u.Username,
			&u.FirstName,
			&u.LastName,
			&u.AuthType,
			&u.Password,
			&u.PasswordVersion,
			&u.Active,
			&u.Version,
			&u.Updated,
			&u.Created,
			&u.Admin,
		)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) insert(tx *sql.Tx) error {
	_, err := sqlUserInsert.Tx(tx).Exec(
		sql.Named("id", u.ID),
		sql.Named("username", u.Username),
		sql.Named("first_name", u.FirstName),
		sql.Named("last_name", u.LastName),
		sql.Named("auth_type", u.AuthType),
		sql.Named("password", u.Password),
		sql.Named("password_version", u.PasswordVersion),
		sql.Named("active", u.Active),
		sql.Named("version", u.Version),
		sql.Named("updated", u.Updated),
		sql.Named("created", u.Created),
		sql.Named("admin", u.Admin),
	)

	return err
}

func (u *User) validate() error {
	if u.Username == "" {
		return NewFailure("A username is required")
	}

	if !urlify(u.Username).is() {
		return NewFailure("A username can only contain letters, numbers and dashes")
	}

	if len(u.FirstName) > UserMaxNameLength {
		return NewFailure("First name must be less than %d characters", UserMaxNameLength)
	}
	if len(u.LastName) > UserMaxNameLength {
		return NewFailure("Last name must be less than %d characters", UserMaxNameLength)
	}

	if u.AuthType != AuthTypePassword {
		return NewFailure("Invalid user Authentication Type")
	}

	return nil
}

func (u *User) canUpdate(who *User) bool {
	return who == nil || (who.ID != u.ID && !who.Admin)
}

// SetActive sets the active status of the given user
func (u *User) SetActive(active bool, who *User) error {
	if u.canUpdate(who) {
		return Unauthorized("You do not have permission to update this user")
	}

	u.Active = active
	_, err := sqlUserUpdateActive.Exec(sql.Named("active", u.Active), sql.Named("id", u.ID))
	return err
}

// SetName sets the user's name
func (u *User) SetName(firstName, lastName string, who *User) error {
	if u.canUpdate(who) {
		return Unauthorized("You do not have permission to update this user")
	}

	u.FirstName = firstName
	u.LastName = lastName
	err := u.validate()
	if err != nil {
		return err
	}

	_, err = sqlUserUpdateName.Exec(sql.Named("first_name", u.FirstName), sql.Named("last_name", u.LastName),
		sql.Named("id", u.ID))
	return err
}

func (u *User) clearPrivate() {
	u.clearPassword()
	u.AuthType = ""
	u.Version = 0
	u.Updated = time.Time{}
	u.Created = time.Time{}
}

func (u *User) clearPassword() {
	u.Password = nil
	u.PasswordVersion = 0
}

// SetAdmin sets if a user is an Administrator or not
func (u *User) SetAdmin(admin bool, who *User) error {
	if who == nil || !who.Admin {
		return Unauthorized("You do not have permission to update this user")
	}

	u.Admin = admin
	_, err := sqlUserUpdateAdmin.Exec(sql.Named("admin", u.Admin), sql.Named("id", u.ID))
	return err
}
