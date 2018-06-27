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

	password          []byte
	passwordVersion   int
	profileImageDraft data.ID
}

// PublicProfile is the publically viewable user information copied from a private user record
type PublicProfile struct {
	ID           data.ID   `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Active       bool      `json:"active"` // whether or not the user is active and can log in
	Created      time.Time `json:"created,omitempty"`
	profileImage data.ID

	admin bool
}

// IsAdmin returns whether or not the user is an admin
func (p *PublicProfile) IsAdmin() bool {
	return p.admin
}

// AuthType determines the authentication method for a given user
const (
	AuthTypePassword = "password"
	//AuthType...
)

// user constants
const (
	userImageWidth  = 300
	userImageHeight = 300
	userIconWidth   = 64
	userIconHeight  = 64
)

// ErrUserNotFound is when a user could not be found
var ErrUserNotFound = NotFound("User Not found")

// ErrUserConflict is when a user is updating an older version of a user record
var ErrUserConflict = Conflict("You are not editing the most current version of this user. Please refresh and try again")

const (
	userPublicColumns  = "u.id, u.username, u.name, u.active, u.profile_image_id, u.admin, u.created"
	userPrivateColumns = `u.id, 
		u.username, 
		u.name, 
		u.auth_type, 
		u.password, 
		u.password_version, 
		u.password_expiration, 
		u.active, 
		u.version, 
		u.updated, 
		u.created, 
		u.admin, 
		u.profile_image_id,
		u.profile_image_draft_id`
)

var sqlUser = struct {
	insert,
	byUsername,
	byID,
	publicProfileByUsername,
	publicProfileByID,
	updateActive,
	updateAdmin,
	updatePassword,
	updateName,
	updateProfileImage,
	updateProfileDraftImage,
	updateUsername *data.Query
}{
	insert: data.NewQuery(`
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
	`),
	byUsername: data.NewQuery(
		fmt.Sprintf(`select %s from users u where username = {{arg "username"}}`, userPrivateColumns)),
	byID: data.NewQuery(
		fmt.Sprintf(`select %s from users u where id = {{arg "id"}}`, userPrivateColumns)),
	publicProfileByUsername: data.NewQuery(
		fmt.Sprintf(`select %s from users u where username = {{arg "username"}}`, userPublicColumns)),
	publicProfileByID: data.NewQuery(
		fmt.Sprintf(`select %s from users u where id = {{arg "id"}}`, userPublicColumns)),
	updateActive:            sqlUserUpdate("active"),
	updateAdmin:             sqlUserUpdate("admin"),
	updatePassword:          sqlUserUpdate("password", "password_version", "password_expiration"),
	updateName:              sqlUserUpdate("name"),
	updateProfileImage:      sqlUserUpdate("profile_image_id", "profile_image_draft_id"),
	updateProfileDraftImage: sqlUserUpdate("profile_image_draft_id"),
	updateUsername:          sqlUserUpdate("username"),
}

var sqlUserUpdate = func(columns ...string) *data.Query {
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
			Created:  time.Now(),
		},
		AuthType: AuthTypePassword,
		Version:  0,
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
		u.PasswordExpiration = data.NullTime{
			Valid: true,
			Time:  time.Now().AddDate(0, 0, SettingMust("PasswordExpirationDays").Int()),
		}
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
	err := u.scan(sqlUser.publicProfileByUsername.
		QueryRow(data.Arg("username", strings.ToLower(username))))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func publicProfileGet(id data.ID) (*PublicProfile, error) {
	u := &PublicProfile{}
	err := u.scan(sqlUser.publicProfileByID.QueryRow(data.Arg("id", id)))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *PublicProfile) scan(record Scanner) error {
	err := record.Scan(
		&p.ID,
		&p.Username,
		&p.Name,
		&p.Active,
		&p.profileImage,
		&p.admin,
		&p.Created,
	)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	return err
}

func (u *User) scan(record Scanner) error {
	err := record.Scan(
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
		&u.admin,
		&u.profileImage,
		&u.profileImageDraft,
	)

	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}

	return err
}

func userFromUsername(tx *sql.Tx, username string) (*User, error) {
	u := &User{}

	err := u.scan(sqlUser.byUsername.Tx(tx).QueryRow(data.Arg("username", strings.ToLower(username))))
	if err != nil {
		return nil, err
	}

	return u, nil
}

func userFromID(tx *sql.Tx, id data.ID) (*User, error) {
	u := &User{}
	err := u.scan(sqlUser.byID.Tx(tx).QueryRow(data.Arg("id", id)))
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) insert(tx *sql.Tx) error {
	_, err := sqlUser.insert.Tx(tx).Exec(
		data.Arg("id", u.ID),
		data.Arg("username", u.Username),
		data.Arg("name", u.Name),
		data.Arg("auth_type", u.AuthType),
		data.Arg("password", u.password),
		data.Arg("password_version", u.passwordVersion),
		data.Arg("password_expiration", u.PasswordExpiration),
		data.Arg("active", u.Active),
		data.Arg("version", u.Version),
		data.Arg("updated", u.Updated),
		data.Arg("created", u.Created),
		data.Arg("admin", u.admin),
	)

	return err
}

func (u *User) validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return NewFailure("A username is required")
	}
	err := data.FieldValidate("user.name", u.Name)
	if err != nil {
		return NewFailureFromErr(err)
	}

	err = data.FieldValidate("user.username", u.Username)
	if err != nil {
		return NewFailureFromErr(err)
	}

	if !urlify(u.Username).is() {
		return NewFailure("A username can only contain letters, numbers and dashes")
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

// setActive sets the active status of the given user, if user is inactivated, then any open sessions are invalidated
func (u *User) setActive(active bool, version int) error {
	return data.BeginTx(func(tx *sql.Tx) error {
		err := u.update(func() (sql.Result, error) {
			return sqlUser.updateActive.Exec(data.Arg("active", active), data.Arg("id", u.ID),
				data.Arg("version", version))
		})
		if err != nil {
			return err
		}

		if !active {
			_, err = sqlSession.invalidateAll.Exec(data.Arg("user_id", u.ID))
			if err != nil {
				return err
			}
		}

		u.Active = active
		return nil
	})
}

// SetName sets the user's name
func (u *User) SetName(name string, version int) error {
	return u.update(func() (sql.Result, error) {
		u.Name = name
		err := u.validate()
		if err != nil {
			return nil, err
		}

		return sqlUser.updateName.Exec(
			data.Arg("name", u.Name),
			data.Arg("id", u.ID),
			data.Arg("version", version),
		)
	})
}

// setAdmin sets if a user is an Administrator or not
func (u *User) setAdmin(admin bool, version int) error {
	err := u.update(func() (sql.Result, error) {
		return sqlUser.updateAdmin.Exec(
			data.Arg("admin", admin),
			data.Arg("id", u.ID),
			data.Arg("version", version),
		)
	})
	if err != nil {
		return err
	}
	u.admin = admin
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
			return sqlUser.updatePassword.Exec(
				data.Arg("password", hash),
				data.Arg("password_version", passVer),
				data.Arg("password_expiration", expires),
				data.Arg("id", u.ID),
				data.Arg("version", version),
			)
		})
		if err != nil {
			return err
		}
		// invalidate all sessions for user
		_, err := sqlSession.invalidateAll.Exec(
			data.Arg("user_id", u.ID),
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

// Admin returns the Admin context for this user, or an error if the user is not an admin
func (u *User) Admin() (*Admin, error) {
	if u == nil || !u.admin {
		return nil, ErrNotAdmin
	}
	return &Admin{u}, nil
}

// ProfileImage returns the user's profile image
func (p *PublicProfile) ProfileImage() *Image {
	return imageGet(p.profileImage)
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
			return sqlUser.updateProfileDraftImage.Tx(tx).Exec(
				data.Arg("profile_image_draft_id", i.id),
				data.Arg("id", u.ID),
				data.Arg("version", version),
			)
		})

		if err != nil {
			return err
		}

		if !u.profileImageDraft.IsNil() {
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

		if !u.profileImage.IsNil() {
			// a previous user image exists, delete it
			err = imageDelete(tx, u.profileImage)
			if err != nil {
				return err
			}

		}

		return u.update(func() (sql.Result, error) {
			return sqlUser.updateProfileImage.Tx(tx).Exec(
				data.Arg("profile_image_id", u.profileImageDraft),
				data.Arg("profile_image_draft_id", nil),
				data.Arg("id", u.ID),
				data.Arg("version", u.Version),
			)
		})
	})
}

// DisplayName is the name displayed.  If no name is set then the username is displayed
func (p *PublicProfile) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Username
}

// DisplayInitials is two characters that display if no profile image is set
func (p *PublicProfile) DisplayInitials() string {
	initials := strings.Split(p.DisplayName(), " ")
	if len(initials) == 1 {
		if len(initials[0]) == 1 {
			return initials[0]
		}
		return strings.ToUpper(string([]rune(initials[0])[:2]))
	}
	return strings.ToUpper(string([]rune(initials[0])[0]) + string([]rune(initials[len(initials)-1])[0]))
}

// Refresh gets the latest version of user
func (u *User) Refresh() error {
	return u.scan(sqlUser.byID.QueryRow(data.Arg("id", u.ID)))
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
		return sqlUser.updateUsername.Exec(
			data.Arg("username", u.Username),
			data.Arg("id", u.ID),
			data.Arg("version", version),
		)
	})
}

// RemoveProfileImage removes the user's current profile image
func (u *User) RemoveProfileImage() error {
	if u.profileImage.IsNil() {
		return nil
	}

	return data.BeginTx(func(tx *sql.Tx) error {
		err := u.update(func() (sql.Result, error) {
			return sqlUser.updateProfileImage.Tx(tx).Exec(
				data.Arg("profile_image_id", nil),
				data.Arg("profile_image_draft_id", nil),
				data.Arg("id", u.ID),
				data.Arg("version", u.Version),
			)
		})
		if err != nil {
			return err
		}

		return imageDelete(tx, u.profileImage)
	})
}

// func (p *publicProfile) Documents(who *app.User) {

// }
