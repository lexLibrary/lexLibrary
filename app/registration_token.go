// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const RegistrationTokenPath = "/registration"

// RegistrationToken is a temporary token that can be used to register new logins for Lex Library
type RegistrationToken struct {
	Token       string        `json:"token"`
	Description string        `json:"description"`
	Limit       int           `json:"limit"`   // number of times this token can be used
	Expires     data.NullTime `json:"expires"` // when this token expires and is no longer valid
	groups      []data.ID

	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`

	valid   bool
	creator data.ID

	creatorCache *PublicProfile
}

var (
	sqlRegistrationTokenInsert = data.NewQuery(`
		insert into registration_tokens (
			token,
			description,
			{{limit}},
			expires,
			valid,
			updated,
			created,
			creator
		) values (
			{{arg "token"}},
			{{arg "description"}},
			{{arg "limit"}},
			{{arg "expires"}},
			{{arg "valid"}},
			{{arg "updated"}},
			{{arg "created"}},
			{{arg "creator"}}
		)
	`)
	sqlRegistrationTokenGroupInsert = data.NewQuery(`
		insert into registration_token_groups (
			token,
			group_id
		) values (
			{{arg "token"}},
			{{arg "group_id"}}
		)
	`)
	sqlRegistrationTokenGet = data.NewQuery(`
		select	t.token,
			t.description,
			t.{{limit}},
			t.expires,
			t.valid,
			t.updated,
			t.created,
			t.creator,
			g.group_id
		from 	registration_tokens t
			left outer join registration_token_groups g 
				on t.token = g.token
		where 	t.token = {{arg "token"}}		
	`)

	sqlRegistrationTokenList = func(validOnly, total bool) *data.Query {
		qry := `
			select	token,
				description,
				{{limit}},
				expires,
				valid,
				updated,
				created,
				creator
			from 	registration_tokens
		`
		if total {
			qry = `select count(*) from registration_tokens`
		}
		if validOnly {
			qry += `
				where valid = {{TRUE}}
				and (expires > {{NOW}} or expires is null)
				and {{limit}} <> 0
			`
		}
		if !total {
			qry += `
				order by created desc
				{{if sqlserver}}
					OFFSET {{arg "offset"}} ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
				{{else}}
					LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
				{{end}}
			`
		}
		return data.NewQuery(qry)
	}
	sqlRegistrationTokenDecrementLimit = data.NewQuery(`
		update 	registration_tokens
		set 	{{limit}} = {{limit}} - 1
		where 	token = {{arg "token"}}
		and 	{{limit}} > 0
	`)

	sqlRegistrationTokenInsertUser = data.NewQuery(`
		insert into registration_token_users (
			token,
			user_id
		) values (
			{{arg "token"}},
			{{arg "user_id"}}
		)
	`)

	sqlRegistrationTokenGroups = data.NewQuery(`
		select	g.id,
			g.name, 
			g.version,
			g.updated, 
			g.created
		from 	groups g,
			registration_token_groups t
		where 	g.id = t.group_id
		and 	t.token = {{arg "token"}}
	`)

	sqlRegistrationTokenUsers = data.NewQuery(fmt.Sprintf(`
		select	%s
		from 	users u,
			registration_token_users t
		where 	u.id = t.user_id
		and 	t.token = {{arg "token"}}
	`, userPublicColumns))

	sqlRegistrationTokenValid = data.NewQuery(`
		update 	registration_tokens
		set 	valid = {{arg "valid"}}
		where 	token = {{arg "token"}}
	`)
)

var errRegistrationTokenInvalid = NewFailure("This registration URL has expired or is no longer valid. " +
	"Please contact your adminstrator for a new URL.")

// NewRegistrationToken generates a new token that can be used to register new users on private instances of Lex Library
// if limit == 0 there is no limit on the number of times the token can be used
// if expires.IsZero() then the token doesn't expire
// the user is automatically made a member of any groups specified in []groups
func (a *Admin) NewRegistrationToken(description string, limit uint, expires time.Time,
	groups []data.ID) (*RegistrationToken, error) {
	setLimit := -1

	if limit != 0 {
		setLimit = int(limit)
	}
	t := &RegistrationToken{
		Token:       Random(128),
		Description: description,
		Limit:       setLimit,
		groups:      groups,
		valid:       true,
		Updated:     time.Now(),
		Created:     time.Now(),
		creator:     a.User().ID,
	}

	if !expires.IsZero() {
		t.Expires = data.NullTime{
			Valid: true,
			Time:  expires,
		}
	}

	err := t.validate()
	if err != nil {
		return nil, err
	}

	err = data.BeginTx(func(tx *sql.Tx) error {
		return t.insert(tx)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// RegistrationTokenList returns a list of Registration Tokens
func (a *Admin) RegistrationTokenList(validOnly bool, offset, limit int) (tokens []*RegistrationToken, total int, err error) {
	if limit == 0 || limit > maxRows {
		limit = 10
	}

	var g errgroup.Group

	g.Go(func() error {
		rows, err := sqlRegistrationTokenList(validOnly, false).
			Query(data.Arg("offset", offset), data.Arg("limit", limit))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			t := &RegistrationToken{}
			err = rows.Scan(&t.Token,
				&t.Description,
				&t.Limit,
				&t.Expires,
				&t.valid,
				&t.Updated,
				&t.Created,
				&t.creator,
			)
			if err != nil {
				return err
			}
			tokens = append(tokens, t)
		}
		return nil
	})

	g.Go(func() error {
		return sqlRegistrationTokenList(validOnly, true).QueryRow().Scan(&total)
	})

	err = g.Wait()
	if err != nil {
		return nil, total, err
	}

	return tokens, total, nil
}

func (t *RegistrationToken) validate() error {
	if strings.TrimSpace(t.Description) == "" {
		return NewFailure("A description is required")
	}
	if t.Expires.Valid && t.Expires.Time.Before(time.Now()) {
		return NewFailure("Expires must be a date in the future")
	}
	if len(t.groups) != 0 {
		query, args := sqlGroupsFromIDs(t.groups, true)
		groupCount := 0
		err := query.QueryRow(args...).Scan(&groupCount)
		if err == sql.ErrNoRows {
			return NewFailure("One or more of the groups are invalid")
		}
		if err != nil {
			return err
		}

		if groupCount != len(t.groups) {
			// one or more groups were not found
			return NewFailure("One or more of the groups are invalid")
		}
	}
	return nil
}

func (t *RegistrationToken) insert(tx *sql.Tx) error {
	if tx == nil {
		panic("A transaction is required for adding registration tokens")
	}

	_, err := sqlRegistrationTokenInsert.Tx(tx).Exec(
		data.Arg("token", t.Token),
		data.Arg("description", t.Description),
		data.Arg("limit", t.Limit),
		data.Arg("expires", t.Expires),
		data.Arg("valid", t.valid),
		data.Arg("updated", t.Updated),
		data.Arg("created", t.Created),
		data.Arg("creator", t.creator),
	)
	if err != nil {
		return err
	}

	for i := range t.groups {
		_, err = sqlRegistrationTokenGroupInsert.Tx(tx).Exec(
			data.Arg("token", t.Token),
			data.Arg("group_id", t.groups[i]),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Creator returns the creator of the registration token
func (t *RegistrationToken) Creator() (*PublicProfile, error) {
	if t.creatorCache != nil {
		return t.creatorCache, nil
	}
	creator, err := publicProfileGet(t.creator)
	if err != nil {
		return nil, err
	}
	t.creatorCache = creator
	return creator, nil
}

// Valid returns whether or not the token is valid, hasn't expired and has any registrations left
func (t *RegistrationToken) Valid() bool {
	if !t.valid {
		return false
	}

	if t.Expires.Valid && t.Expires.Time.Before(time.Now()) && !t.Expires.Time.IsZero() {
		return false
	}

	if t.Limit == 0 {
		return false
	}
	return true
}

// RegisterUserFromToken creates a new user if the passed in token is valid
func RegisterUserFromToken(username, password, token string) (*User, error) {
	t, err := registrationTokenGet(token)
	if err != nil {
		return nil, err
	}

	if !t.Valid() {
		return nil, errRegistrationTokenInvalid
	}

	var u *User

	err = data.BeginTx(func(tx *sql.Tx) error {
		if t.Limit != -1 {
			err = t.decrementLimit(tx)
			if err != nil {
				return err
			}
		}
		u, err = userNew(tx, username, password)
		if err != nil {
			return err
		}

		for i := range t.groups {
			result, err := sqlGroupInsertMember.Tx(tx).Exec(
				data.Arg("group_id", t.groups[i]),
				data.Arg("user_id", u.ID),
				data.Arg("admin", false),
			)
			if err != nil {
				return err
			}

			rows, err := result.RowsAffected()
			if err != nil {
				return err
			}

			if rows == 0 {
				// shouldn't happen
				return NewFailure("Cannot add an invalid user to a group")
			}
		}
		_, err = sqlRegistrationTokenInsertUser.Tx(tx).Exec(
			data.Arg("token", t.Token),
			data.Arg("user_id", u.ID),
		)

		return err
	})

	return u, err
}

func (a *Admin) RegistrationToken(token string) (*RegistrationToken, error) {
	return registrationTokenGet(token)
}

func registrationTokenGet(token string) (*RegistrationToken, error) {
	t := &RegistrationToken{}
	rows, err := sqlRegistrationTokenGet.Query(data.Arg("token", token))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	found := false

	for rows.Next() {
		found = true
		var groupID data.ID

		err = rows.Scan(
			&t.Token,
			&t.Description,
			&t.Limit,
			&t.Expires,
			&t.valid,
			&t.Updated,
			&t.Created,
			&t.creator,
			&groupID,
		)
		if err != nil {
			return nil, err
		}

		if !groupID.IsNil() {
			t.groups = append(t.groups, groupID)
		}
	}

	if !found {
		return nil, errRegistrationTokenInvalid
	}

	return t, nil
}

// decrementLimit decrements the available registration limit
func (t *RegistrationToken) decrementLimit(tx *sql.Tx) error {
	result, err := sqlRegistrationTokenDecrementLimit.Tx(tx).Exec(data.Arg("token", t.Token))
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errRegistrationTokenInvalid
	}
	return nil
}

// Invalidate sets a registration token's valid status to false
func (t *RegistrationToken) Invalidate() error {
	result, err := sqlRegistrationTokenValid.Exec(data.Arg("token", t.Token), data.Arg("valid", false))
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errRegistrationTokenInvalid
	}

	return nil
}

// Groups returns the groups associated with this registration token, if any
func (t *RegistrationToken) Groups() ([]*Group, error) {
	var groups []*Group
	rows, err := sqlRegistrationTokenGroups.Query(data.Arg("token", t.Token))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		g := &Group{}
		err = g.scan(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Users returns the users registered with this given registration token
func (t *RegistrationToken) Users() ([]*PublicProfile, error) {
	var users []*PublicProfile
	rows, err := sqlRegistrationTokenUsers.Query(data.Arg("token", t.Token))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &PublicProfile{}
		err = u.scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (t *RegistrationToken) URL() (string, error) {
	u, err := url.Parse(SettingMust("URL").String())
	if err != nil {
		return "", errors.Wrap(err, "URL Setting is an invalid URL")
	}
	u.Path = path.Join(RegistrationTokenPath, t.Token)
	return u.String(), nil
}
