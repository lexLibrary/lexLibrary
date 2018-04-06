// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// RegistrationToken is a temporary token that can be used to register new logins for Lex Library
type RegistrationToken struct {
	Token   string
	Limit   int       // number of times this token can be used
	Expires time.Time // when this token expires and is no longer valid
	Groups  []data.ID // users registered by this token will be members of these groups

	Valid   bool
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`

	creator data.ID
}

var (
	sqlRegistrationTokenInsert = data.NewQuery(`
		insert into registration_tokens (
			token,
			"limit",
			expires,
			valid,
			updated,
			created
		) values (
			{{arg "token"}},
			{{arg "limit"}},
			{{arg "expires"}},
			{{arg "valid"}},
			{{arg "updated"}},
			{{arg "created"}}
		)
	`)
	sqlRegistrationTokenGroupInsert = data.NewQuery(`
		insert into registration_tokens (
			token,
			group_id
		) values (
			{{arg "token"}},
			{{arg "group_id"}}
		)
	`)
)

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
		Token:   Random(128),
		Limit:   setLimit,
		Expires: expires,
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

	err = t.insert(nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *RegistrationToken) validate() error {
	if len(t.Groups) != 0 {
		query, args := sqlGroupsFromIDs(t.Groups)
		result, err := query.Exec(args...)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if int(rows) != len(t.Groups) {
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

// func (t *RegistrationToken) consume() error {
// }
