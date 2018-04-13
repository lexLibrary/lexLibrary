// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// RegistrationToken is a temporary token that can be used to register new logins for Lex Library
type RegistrationToken struct {
	Token   string        `json:"token"`
	Limit   int           `json:"limit"`   // number of times this token can be used
	Expires data.NullTime `json:"expires"` // when this token expires and is no longer valid
	Groups  []data.ID     `json:"groups"`  // users registered by this token will be members of these groups

	Valid   bool      `json:"valid"`
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`

	creator data.ID
}

var (
	sqlRegistrationTokenInsert = data.NewQuery(`
		insert into registration_tokens (
			token,
			{{limit}},
			expires,
			valid,
			updated,
			created,
			creator
		) values (
			{{arg "token"}},
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
	sqlRegistrationTokenDecrementLimit = data.NewQuery(`
		update 	registration_tokens
		set 	{{limit}} = {{limit}} - 1
		where 	token = {{arg "token"}}
		and 	{{limit}} > 0
	`)
)

var errRegistrationTokenInvalid = NewFailure("This registration URL has expired or is no longer valid.  Please contact your adminstrator for a new URL.")

// NewRegistrationToken generates a new token that can be used to register new users on private instances of Lex Library
// if limit == 0 there is no limit on the number of times the token can be used
// if expires.IsZero() then the token doesn't expire
// the user is automatically made a member of any groups specified in []groups
func (a *Admin) NewRegistrationToken(limit uint, expires time.Time, groups []data.ID) (*RegistrationToken, error) {
	setLimit := -1

	if limit != 0 {
		setLimit = int(limit)
	}
	t := &RegistrationToken{
		Token: Random(128),
		Limit: setLimit,
		Expires: data.NullTime{
			Valid: !expires.IsZero(),
			Time:  expires,
		},
		Groups:  groups,
		Valid:   true,
		Updated: time.Now(),
		Created: time.Now(),
		creator: a.User.ID,
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

func (t *RegistrationToken) validate() error {
	if !t.Expires.Time.IsZero() && t.Expires.Time.Before(time.Now()) {
		return NewFailure("Expires must be a date in the future")
	}
	if len(t.Groups) != 0 {
		query, args := sqlGroupsFromIDs(t.Groups, true)
		groupCount := 0
		err := query.QueryRow(args...).Scan(&groupCount)
		if err == sql.ErrNoRows {
			return NewFailure("One or more of the groups are invalid")
		}
		if err != nil {
			return err
		}

		if groupCount != len(t.Groups) {
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
		sql.Named("token", t.Token),
		sql.Named("limit", t.Limit),
		sql.Named("expires", t.Expires),
		sql.Named("valid", t.Valid),
		sql.Named("updated", t.Updated),
		sql.Named("created", t.Created),
		sql.Named("creator", t.creator),
	)
	if err != nil {
		return err
	}

	for i := range t.Groups {
		_, err = sqlRegistrationTokenGroupInsert.Tx(tx).Exec(
			sql.Named("token", t.Token),
			sql.Named("group_id", t.Groups[i]),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *RegistrationToken) Creator() (*PublicProfile, error) {
	return publicProfileGet(t.creator)
}

// RegisterUserFromToken creates a new user if the passed in token is valid
func RegisterUserFromToken(username, password, token string) (*User, error) {
	t, err := registrationTokenGet(token)
	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, errRegistrationTokenInvalid
	}

	if t.Expires.Time.Before(time.Now()) && !t.Expires.Time.IsZero() {
		return nil, errRegistrationTokenInvalid
	}

	if t.Limit == 0 {
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

		for i := range t.Groups {
			result, err := sqlGroupInsertMember.Tx(tx).Exec(
				sql.Named("group_id", t.Groups[i]),
				sql.Named("user_id", u.ID),
				sql.Named("admin", false),
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
		return nil
	})

	return u, err
}

func registrationTokenGet(token string) (*RegistrationToken, error) {
	t := &RegistrationToken{}
	rows, err := sqlRegistrationTokenGet.Query(sql.Named("token", token))
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
			&t.Limit,
			&t.Expires,
			&t.Valid,
			&t.Updated,
			&t.Created,
			&t.creator,
			&groupID,
		)
		if err != nil {
			return nil, err
		}

		if groupID.Valid {
			t.Groups = append(t.Groups, groupID)
		}
	}

	if !found {
		return nil, errRegistrationTokenInvalid
	}

	return t, nil
}

// decrementLimit decrements the available registration limit
func (t *RegistrationToken) decrementLimit(tx *sql.Tx) error {
	result, err := sqlRegistrationTokenDecrementLimit.Tx(tx).Exec(sql.Named("token", t.Token))
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

//TODO: SetValid by admin
