// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Group is a collection of users that grants access to a set of documents
type Group struct {
	ID      data.ID   `json:"id"`
	Name    string    `json:"name"`
	Version int       `json:"version"`
	Updated time.Time `json:"updated,omitempty"`
	Created time.Time `json:"created,omitempty"`
}

const (
	groupMaxNameLength = 64
)

var (
	sqlGroupInsert = data.NewQuery(`insert into groups (
		id,
		name, 
		version,
		updated, 
		created
	) values (
		{{arg "id"}}, 
		{{arg "name"}}, 
		{{arg "version"}}, 
		{{arg "updated"}}, 
		{{arg "created"}}
	)`)
	sqlUserToGroupInsert = data.NewQuery(`insert into users_to_groups (
		user_id,
		group_id,
		admin
	) values (
		{{arg "user_id"}}, 
		{{arg "group_id"}}, 
		{{arg "admin"}}
	)`)

	sqlGroupFromName = data.NewQuery(`select id, name, version, updated, created from groups where name = {{arg "name"}}`)
)

var (
	// ErrGroupNotFound is returned when a group couldn't be found
	ErrGroupNotFound = NotFound("Group not found")
	// ErrGroupConflict occurs when someone updates an older version of a group
	ErrGroupConflict = Conflict("You are not editing the most current version of this group. Please refresh and try again.")
)

// NewGroup creates a new group and sets the creator as the groups Admin
func (u *User) NewGroup(name string) (*Group, error) {
	g := &Group{
		ID:      data.NewID(),
		Name:    name,
		Version: 0,
		Updated: time.Now(),
		Created: time.Now(),
	}

	err := g.validate()
	if err != nil {
		return nil, err
	}

	_, err = GroupFromName(name)
	if err == nil {
		return nil, NewFailure("A group already exists with the name %s. Please choose another.", name)
	}

	if err != ErrGroupNotFound {
		return nil, err
	}

	err = data.BeginTx(func(tx *sql.Tx) error {
		err = g.insert(tx)
		if err != nil {
			return err
		}

		// add current user as member and admin
		_, err = sqlUserToGroupInsert.Tx(tx).Exec(
			sql.Named("user_id", u.ID),
			sql.Named("group_id", g.ID),
			sql.Named("admin", true),
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

// GroupFromName returns a group based on the passed in name
func GroupFromName(name string) (*Group, error) {
	g := &Group{}

	err := sqlGroupFromName.QueryRow(sql.Named("name", name)).Scan(
		&g.ID,
		&g.Name,
		&g.Version,
		&g.Updated,
		&g.Created,
	)
	if err == sql.ErrNoRows {
		return nil, ErrGroupNotFound
	}
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Group) validate() error {
	if g.Name == "" {
		return NewFailure("A group name is required")
	}
	if len(g.Name) > groupMaxNameLength {
		return NewFailure("A group name must be less than %d characters", groupMaxNameLength)
	}
	return nil
}

func (g *Group) insert(tx *sql.Tx) error {
	_, err := sqlGroupInsert.Tx(tx).Exec(
		sql.Named("id", g.ID),
		sql.Named("name", g.Name),
		sql.Named("version", g.Version),
		sql.Named("updated", g.Updated),
		sql.Named("created", g.Created),
	)

	return err
}

func (g *Group) update(update func() (sql.Result, error)) error {
	r, err := update()

	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupConflict
	}
	g.Version++
	return nil
}
