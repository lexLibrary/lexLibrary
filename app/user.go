// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// User is a user login to Lex Library
type User struct {
	PublicProfile

	AuthType           string        `json:"authType,omitempty"`
	PasswordExpiration data.NullTime `json:"passwordExpiration"`
	Version            int           `json:"version"` // version of this record starting with 0
	Updated            time.Time     `json:"updated,omitempty"`
	Created            time.Time     `json:"created,omitempty"`

	password          []byte
	passwordVersion   int
	profileImage      data.ID
	profileImageDraft data.ID
}

// PublicProfile is the publically viewable user information copied from a private user record
type PublicProfile struct {
	ID       data.ID `json:"id"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Admin    bool    `json:"admin"`
	Active   bool    `json:"active"` // whether or not the user is active and can log in
}

// AuthType determines the authentication method for a given user
const (
	AuthTypePassword = "password"
	//AuthType...
)

// user constants
const (
	UserMaxNameLength = 64 // This is pretty arbitrary but there should probably be some limit
	usernameMinLength = 3

	// images
	userImageWidth  = 300
	userImageHeight = 300
	userIconWidth   = 64
	userIconHeight  = 64
)

// ErrUserNotFound is when a user could not be found
var ErrUserNotFound = NotFound("User Not found")

// ErrUserConflict is when a user is updating an older version of a user record
var ErrUserConflict = Conflict("You are not editing the most current version of this user. Please refresh and try again")

var (
	sqlUserInsert = data.NewQuery(`
		insert into users (
			id,
			username, 
			name, 
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
			{{arg "name"}}, 
			{{arg "auth_type"}},
			{{arg "password"}},
			{{arg "password_version"}},
			{{arg "password_expiration"}},
			{{arg "active"}},
			{{arg "version"}},
			{{arg "updated"}}, 
			{{arg "created"}},
			{{arg "admin"}}
		)
	`)

	userPublicColumns  = "id, username, name, active, admin"
	userPrivateColumns = `id, 
		username, 
		name, 
		auth_type, 
		password, 
		password_version, 
		password_expiration, 
		active, 
		version, 
		updated, 
		created, 
		admin, 
		profile_image_id,
		profile_image_draft_id`

	sqlUserFromUsername = data.NewQuery(
		fmt.Sprintf(`select %s from users where username = {{arg "username"}}`, userPrivateColumns))
	sqlUserFromID = data.NewQuery(
		fmt.Sprintf(`select %s from users where id = {{arg "id"}}`, userPrivateColumns))
	sqlUserPublicProfileFromUsername = data.NewQuery(
		fmt.Sprintf(`select %s from users where username = {{arg "username"}}`, userPublicColumns))
	sqlUserPublicProfileFromID = data.NewQuery(
		fmt.Sprintf(`select %s from users where id = {{arg "id"}}`, userPublicColumns))

	sqlUserUpdate = func(columns ...string) *data.Query {
		updates := ""
		for i := range columns {
			updates += fmt.Sprintf(`%s = {{arg "%s"}},`, columns[i], columns[i])
		}
		return data.NewQuery(fmt.Sprintf(`
		update users set %s
			updated = {{NOW}}, 
			version = version + 1 
		where id = {{arg "id"}} 
		and version = {{arg "version"}}`, updates))
	}

	sqlUserUpdateActive            = sqlUserUpdate("active")
	sqlUserUpdateAdmin             = sqlUserUpdate("admin")
	sqlUserUpdatePassword          = sqlUserUpdate("password", "password_version", "password_expiration")
	sqlUserUpdateName              = sqlUserUpdate("name")
	sqlUserUpdateProfileImage      = sqlUserUpdate("profile_image_id", "profile_image_draft_id")
	sqlUserUpdateProfileDraftImage = sqlUserUpdate("profile_image_draft_id")
	sqlUserUpdateUsername          = sqlUserUpdate("username")
)

// UserNew creates a new user, from the sign up page
// returns unauthorized if public signups are disabled
func UserNew(username, password string) (*User, error) {
	if !SettingMust("AllowPublicSignups").Bool() {
		return nil, Unauthorized("Public signups are currently disabled")
	}

	return userNew(nil, username, password)
}

func userNew(tx *sql.Tx, username, password string) (*User, error) {
	u := &User{
		PublicProfile: PublicProfile{
			ID:       data.NewID(),
			Username: strings.ToLower(username),
			Active:   true,
		},
		AuthType: AuthTypePassword,
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
	return publicProfileFromRow(sqlUserPublicProfileFromUsername.
		QueryRow(sql.Named("username", strings.ToLower(username))))
}

func publicProfileGet(id data.ID) (*PublicProfile, error) {
	return publicProfileFromRow(sqlUserPublicProfileFromID.
		QueryRow(sql.Named("id", id)))
}

func publicProfileFromRow(row *sql.Row) (*PublicProfile, error) {
	u := &PublicProfile{}

	err := row.
		Scan(
			&u.ID,
			&u.Username,
			&u.Name,
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
		&u.Name,
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
		&u.profileImageDraft,
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
		&u.Name,
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
		&u.profileImageDraft,
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
		sql.Named("name", u.Name),
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

	if len(u.Username) < usernameMinLength {
		return NewFailure("A username must be more than %d characters", usernameMinLength)
	}

	if !urlify(u.Username).is() {
		return NewFailure("A username can only contain letters, numbers and dashes")
	}

	if len(u.Name) > UserMaxNameLength {
		return NewFailure("Name must be less than %d characters", UserMaxNameLength)
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
func (u *User) SetName(name string, version int) error {
	return u.update(func() (sql.Result, error) {
		u.Name = name
		err := u.validate()
		if err != nil {
			return nil, err
		}

		return sqlUserUpdateName.Exec(
			sql.Named("name", u.Name),
			sql.Named("id", u.ID),
			sql.Named("version", version),
		)
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

// ProfileImageDraft returns the draft profile image
func (u *User) ProfileImageDraft() *Image {
	return imageGet(u.profileImageDraft)
}

// UploadProfileImageDraft sets the current user's profile image
func (u *User) UploadProfileImageDraft(upload Upload, version int) error {
	i, err := imageNew(upload)
	if err != nil {
		return err
	}

	i.thumbMinDimension = userIconWidth

	return data.BeginTx(func(tx *sql.Tx) error {
		err = i.insert(tx)
		if err != nil {
			return err
		}

		err = u.update(func() (sql.Result, error) {
			return sqlUserUpdateProfileDraftImage.Tx(tx).Exec(
				sql.Named("profile_image_draft_id", i.id),
				sql.Named("id", u.ID),
				sql.Named("version", version),
			)
		})

		if err != nil {
			return err
		}

		if u.profileImageDraft.Valid {
			// a previous draft user image exists, delete it
			err = imageDelete(tx, u.profileImageDraft)
			if err != nil {
				return err
			}
		}
		u.profileImageDraft = i.id
		return nil
	})
}

// SetProfileImageFromDraft crops the exising user's profile image
func (u *User) SetProfileImageFromDraft(x0, y0, x1, y1 float64) error {
	i, err := imageGet(u.profileImageDraft).raw()
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

	return data.BeginTx(func(tx *sql.Tx) error {
		err := i.update(tx, i.version)
		if err != nil {
			return err
		}

		if u.profileImage.Valid {
			// a previous user image exists, delete it
			err = imageDelete(tx, u.profileImage)
			if err != nil {
				return err
			}

		}

		return u.update(func() (sql.Result, error) {
			return sqlUserUpdateProfileImage.Tx(tx).Exec(
				sql.Named("profile_image_id", u.profileImageDraft),
				sql.Named("profile_image_draft_id", nil),
				sql.Named("id", u.ID),
				sql.Named("version", u.Version),
			)
		})
	})
}

// DisplayName is the name displayed.  If no name is set then the username is displayed
func (u *User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Username
}

// DisplayInitials is two characters that display if no profile image is set
func (u *User) DisplayInitials() string {
	initials := strings.Split(u.DisplayName(), " ")
	if len(initials) == 1 {
		if len(initials[0]) == 1 {
			return initials[0]
		}
		return strings.ToUpper(string([]rune(initials[0])[:2]))
	}
	return strings.ToUpper(string([]rune(initials[0])[0]) + string([]rune(initials[len(initials)-1])[0]))
}

// Latest gets the latest version of user
func (u *User) Latest() (*User, error) {
	return userFromID(nil, u.ID)
}

// SetUsername updates the user's current username
func (u *User) SetUsername(username string, version int) error {
	if username == u.Username {
		// no change
		return nil
	}
	u.Username = strings.ToLower(username)
	err := u.validate()
	if err != nil {
		return err
	}

	_, err = userFromUsername(nil, username)

	if err == nil {
		return NewFailure("A user with the username %s already exists", username)
	}

	if err != ErrUserNotFound {
		return err
	}

	return u.update(func() (sql.Result, error) {
		return sqlUserUpdateUsername.Exec(
			sql.Named("username", u.Username),
			sql.Named("id", u.ID),
			sql.Named("version", version),
		)
	})
}
