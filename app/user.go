// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"io"
	"math"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// User is a user login to Lex Library
type User struct {
	ID                 data.ID       `json:"id"`
	Username           string        `json:"username"`
	FirstName          string        `json:"firstName"`
	LastName           string        `json:"lastName"`
	AuthType           string        `json:"authType,omitempty"`
	PasswordExpiration data.NullTime `json:"passwordExpiration"`
	Active             bool          `json:"active"`            // whether or not the user is active and can log in
	Version            int           `json:"version,omitempty"` // version of this record starting with 0
	Updated            time.Time     `json:"updated,omitempty"`
	Created            time.Time     `json:"created,omitempty"`
	Admin              bool          `json:"admin"`

	password        []byte
	passwordVersion int
	profileImage    data.ID
}

// PublicProfile is the publically viewable user information copied from a private user record
type PublicProfile struct {
	ID        data.ID `json:"id"`
	Username  string  `json:"username"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Admin     bool    `json:"admin"`
	Active    bool    `json:"active"` // whether or not the user is active and can log in
}

// AuthType determines the authentication method for a given user
const (
	AuthTypePassword = "password"
	//AuthType...
)

// user constants
const (
	UserMaxNameLength = 64 // This is pretty arbitrary but there should probably be some limit

	// images
	userImageWidth  = 300
	userImageHeight = 300
	userIconWidth   = 32
	userIconHeight  = 32
)

// ErrUserNotFound is when a user could not be found
var ErrUserNotFound = NotFound("User Not found")

// ErrUserConflict is when a user is updating an older version of a user record
var ErrUserConflict = Conflict("You are not editing the most current version of this user. Please refresh and try again")

var (
	sqlUserInsert = data.NewQuery(`insert into users (
		id,
		username, 
		first_name, 
		last_name, 
		auth_type,
		password,
		password_version,
		password_expiration,
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
		{{arg "password_expiration"}},
		{{arg "active"}},
		{{arg "version"}},
		{{arg "updated"}}, 
		{{arg "created"}},
		{{arg "admin"}}
	)`)
	sqlUserFromUsername = data.NewQuery(`
		select 	id, 
			username, 
			first_name, 
			last_name, 
			auth_type, 
			password, 
			password_version, 
			password_expiration, 
			active, 
			version, 
			updated, 
			created, 
			admin, 
			profile_image 
		from users where username = {{arg "username"}}
	`)
	sqlUserFromID = data.NewQuery(`
		select 	id, 
			username, 
			first_name, 
			last_name, 
			auth_type, 
			password, 
			password_version, 
			password_expiration, 
			active, 
			version, 
			updated, 
			created, 
			admin, 
			profile_image 
		from users where id = {{arg "id"}}
	`)
	sqlUserPublicProfile = data.NewQuery(`
		select id, username, first_name, last_name, active, admin 
		from users where username = {{arg "username"}}`)

	sqlUserUpdateActive = data.NewQuery(`update users set active = {{arg "active"}}, updated = {{now}}, version = version + 1 
		where id = {{arg "id"}} and version = {{arg "version"}}`)
	sqlUserUpdateAdmin = data.NewQuery(`update users set admin = {{arg "admin"}}, updated = {{now}}, version = version + 1
		where id = {{arg "id"}} and version = {{arg "version"}}`)
	sqlUserUpdatePassword = data.NewQuery(`update users
		set 	password = {{arg "password"}},
			password_version = {{arg "password_version"}},
			password_expiration = {{arg "password_expiration"}},
			updated = {{now}},
			version = version + 1
		where id = {{arg "id"}}
		and version = {{arg "version"}}`)
	sqlUserUpdateName = data.NewQuery(`update users set first_name = {{arg "first_name"}}, 
		last_name = {{arg "last_name"}}, updated = {{now}}, version = version + 1 where id = {{arg "id"}} 
		and version = {{arg "version"}}`)
	sqlUserUpdateProfileImage = data.NewQuery(`update users set profile_image = {{arg "profile_image"}}, updated = {{now}},
		version = version + 1
		where id = {{arg "id"}} and version = {{arg "version"}}`)
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
		ID:       data.NewID(),
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

	u.passwordVersion = passVer
	u.password = hash

	_, err = userFromUsername(tx, u.Username)

	if err == nil {
		return nil, NewFailure("A user with the username %s already exists", u.Username)
	}

	if err != ErrUserNotFound {
		return nil, err
	}

	if SettingMust("PasswordExpirationDays").Int() >= 0 {
		u.PasswordExpiration.Time = time.Now().AddDate(0, 0, SettingMust("PasswordExpirationDays").Int())
	}

	err = u.insert(tx)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// UserGet retrieves the publically viewable user profile information based from the passed in username
// app internal code should use un-exported funcs that contain the full User record
func UserGet(username string) (*PublicProfile, error) {
	u := &PublicProfile{}

	err := sqlUserPublicProfile.QueryRow(sql.Named("username", strings.ToLower(username))).
		Scan(
			&u.ID,
			&u.Username,
			&u.FirstName,
			&u.LastName,
			&u.Active,
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

func userFromUsername(tx *sql.Tx, username string) (*User, error) {
	u := &User{}

	err := sqlUserFromUsername.Tx(tx).QueryRow(sql.Named("username", strings.ToLower(username))).Scan(
		&u.ID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.AuthType,
		&u.password,
		&u.passwordVersion,
		&u.PasswordExpiration,
		&u.Active,
		&u.Version,
		&u.Updated,
		&u.Created,
		&u.Admin,
		&u.profileImage,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return u, nil
}

func userFromID(tx *sql.Tx, id data.ID) (*User, error) {
	u := &User{}
	err := sqlUserFromID.Tx(tx).QueryRow(sql.Named("id", id)).Scan(
		&u.ID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.AuthType,
		&u.password,
		&u.passwordVersion,
		&u.PasswordExpiration,
		&u.Active,
		&u.Version,
		&u.Updated,
		&u.Created,
		&u.Admin,
		&u.profileImage,
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
		sql.Named("password", u.password),
		sql.Named("password_version", u.passwordVersion),
		sql.Named("password_expiration", u.PasswordExpiration),
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

func (u *User) update(update func() (sql.Result, error)) error {
	r, err := update()

	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserConflict
	}
	u.Version++
	return nil
}

// setActive sets the active status of the given user
func (u *User) setActive(active bool, version int) error {
	err := u.update(func() (sql.Result, error) {
		return sqlUserUpdateActive.Exec(sql.Named("active", active), sql.Named("id", u.ID),
			sql.Named("version", version))
	})
	if err != nil {
		return err
	}

	u.Active = active
	return nil
}

// SetName sets the user's name
func (u *User) SetName(firstName, lastName string, version int) error {
	return u.update(func() (sql.Result, error) {
		u.FirstName = firstName
		u.LastName = lastName
		err := u.validate()
		if err != nil {
			return nil, err
		}

		return sqlUserUpdateName.Exec(sql.Named("first_name", u.FirstName), sql.Named("last_name", u.LastName),
			sql.Named("id", u.ID), sql.Named("version", version))
	})
}

// setAdmin sets if a user is an Administrator or not
func (u *User) setAdmin(admin bool, version int) error {
	err := u.update(func() (sql.Result, error) {
		return sqlUserUpdateAdmin.Exec(
			sql.Named("admin", admin),
			sql.Named("id", u.ID),
			sql.Named("version", version),
		)
	})
	if err != nil {
		return err
	}
	u.Admin = admin
	return nil
}

// UserSetExpiredPassword sets a user's password when it has expired and they can't login to change their password
func UserSetExpiredPassword(username, oldPassword, newPassword string) (*User, error) {
	u, err := userFromUsername(nil, username)
	if err != nil {
		return nil, err
	}

	err = u.SetPassword(oldPassword, newPassword, u.Version)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// SetPassword updates a users password, and invalidates any existing sessions opened with the old password
func (u *User) SetPassword(oldPassword, newPassword string, version int) error {
	err := passwordVersions[u.passwordVersion].compare(oldPassword, u.password)
	if err != nil {
		if err == ErrPasswordMismatch {
			return NewFailureFromErr(err)
		}
		return err
	}
	if oldPassword == newPassword {
		return NewFailure("The new password cannot match the previous password")
	}
	// hash new password
	err = ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	passVer := len(passwordVersions) - 1

	hash, err := passwordVersions[passVer].hash(newPassword)
	if err != nil {
		return err
	}

	return data.BeginTx(func(tx *sql.Tx) error {
		var expires data.NullTime
		if SettingMust("PasswordExpirationDays").Int() >= 0 {
			expires.Time = time.Now().AddDate(0, 0, SettingMust("PasswordExpirationDays").Int())
		}

		// update password, version and expiration
		err = u.update(func() (sql.Result, error) {
			return sqlUserUpdatePassword.Exec(
				sql.Named("password", hash),
				sql.Named("password_version", passVer),
				sql.Named("password_expiration", expires),
				sql.Named("id", u.ID),
				sql.Named("version", version),
			)
		})
		if err != nil {
			return err
		}
		// invalidate all sessions for user
		_, err := sqlSessionInvalidateAll.Exec(
			sql.Named("user_id", u.ID),
			sql.Named("now", time.Now()),
		)
		if err != nil {
			return err
		}
		u.passwordVersion = passVer
		u.password = hash
		u.PasswordExpiration = expires
		return nil
	})
}

// AsAdmin returns the Admin context for this user
func (u *User) AsAdmin() *Admin {
	return &Admin{u}
}

// ProfileImage returns the user's profile image
func (u *User) ProfileImage() *Image {
	return imageGet(u.profileImage)
}

// SetProfileImage sets the current user's profile image
func (u *User) SetProfileImage(rc io.ReadCloser, name, contentType string, version int) error {
	i, err := imageNew(name, contentType, rc)
	if err != nil {
		return err
	}

	i.thumbMinDimension = userIconWidth

	return data.BeginTx(func(tx *sql.Tx) error {
		if u.profileImage.Valid {
			// a previous user image exists, delete it
			err = imageDelete(tx, u.profileImage)
			if err != nil {
				return err
			}

		}

		err = i.insert(tx)
		if err != nil {
			return err
		}

		err = u.update(func() (sql.Result, error) {
			return sqlUserUpdateProfileImage.Tx(tx).Exec(
				sql.Named("profile_image", i.id),
				sql.Named("id", u.ID),
				sql.Named("version", version),
			)
		})
		if err != nil {
			return err
		}

		u.profileImage = i.id
		return nil
	})
}

// CropProfileImage crops the exising user's profile image
func (u *User) CropProfileImage(x0, y0, x1, y1 float64) error {
	i, err := imageGet(u.profileImage).raw()
	if err != nil {
		return err
	}

	// validate inputs, instead of failing set them to predefined values
	if x0 < 0 || y0 < 0 || x1 < 0 || y1 < 0 {
		x0 = 0
		x1 = float64(i.decoded.Bounds().Dx())
		y0 = 0
		y1 = float64(i.decoded.Bounds().Dy())
	}

	if (x1 - x0) < userIconWidth {
		x0 = 0
		x1 = float64(i.decoded.Bounds().Dx())
	}

	if (y1 - y0) < userIconHeight {
		y0 = 0
		y1 = float64(i.decoded.Bounds().Dy())
	}

	ratio := (x1 - x0) / (y1 - y0)
	//give a little wiggle room for rounding
	if ratio < 0.95 || ratio > 1.05 {
		min := math.Min(float64(i.decoded.Bounds().Dx()), float64(i.decoded.Bounds().Dy()))

		err = i.cropCenter(int(min), int(min))
		if err != nil {
			return err
		}
	} else {
		err = i.crop(int(math.Round(x0)), int(math.Round(y0)), int(math.Round(x1)), int(math.Round(y1)))
		if err != nil {
			return err
		}
	}

	err = i.resize(userImageWidth, userImageHeight)
	if err != nil {
		return err
	}

	i.thumbMinDimension = userIconWidth

	return i.update(nil, i.version)
}
