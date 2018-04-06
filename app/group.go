// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
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

// GroupAdmin is a user who can administer a group
type GroupAdmin struct {
	user  *User
	group *Group
}

const (
	groupMaxNameLength = 64
)

var (
	sqlGroupInsert = data.NewQuery(`
		insert into groups (
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
		)
	`)
	sqlUserToGroupInsert = data.NewQuery(`
		insert into user_to_groups (
			user_id,
			group_id,
			admin
		) values (
			{{arg "user_id"}}, 
			{{arg "group_id"}}, 
			{{arg "admin"}}
		)
	`)
	sqlGroupFromName = data.NewQuery(`
		select id, name, version, updated, created 
		from groups 
		where name = {{arg "name"}}
	`)
	sqlGroupsFromIDs = func(ids []data.ID) (*data.Query, []sql.NamedArg) {
		in := ""
		args := make([]sql.NamedArg, len(ids))
		for i := range ids {
			if i != 0 {
				in += ", "
			}
			in += `{{arg "id"}}`
			args[i].Name = "id"
			args[i].Value = ids[i]
		}
		return data.NewQuery(fmt.Sprintf(`
			select id, name, version, updated, created 
			from groups 
			where id in (%s)
		`, in)), args
	}
	sqlGroupGetMember = data.NewQuery(`
		select admin 
		from user_to_groups 
		where user_id = {{arg "user_id"}} 
		and group_id = {{arg "group_id"}}
	`)
	sqlGroupUpdate = data.NewQuery(`
		update groups set name = {{arg "name"}},
			updated = {{NOW}}, 
			version = version + 1 
		where id = {{arg "id"}} 
		and version = {{arg "version"}}
	`)

	sqlGroupInsertMember = data.NewQuery(`
		insert into user_to_groups (
			user_id,
			group_id,
			admin
		) select id,
			{{arg "group_id"}},
			{{arg "admin"}}
		from users
		where id = {{arg "user_id"}}
		and active = {{TRUE}}
	`)

	sqlGroupUpdateMember = data.NewQuery(`
		update user_to_groups
		set admin = {{arg "admin"}}
		where user_id = {{arg "user_id"}}
		and group_id = {{arg "group_id"}}
	`)
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
		return err
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

// Admin returns a group Admin for administering this specific group
func (g *Group) Admin(who *User) (*GroupAdmin, error) {
	if who == nil {
		return nil, Unauthorized("You must be logged in to administer this group")
	}

	// site admins have built in group admin permissions
	if who.Admin {
		return &GroupAdmin{
			user:  who,
			group: g,
		}, nil
	}
	admin := false
	err := sqlGroupGetMember.QueryRow(sql.Named("user_id", who.ID), sql.Named("group_id", g.ID)).Scan(
		&admin,
	)
	if err == sql.ErrNoRows {
		return nil, Unauthorized("You must be an admin to edit this group")
	}

	if err != nil {
		return nil, err
	}

	if !admin {
		return nil, Unauthorized("You must be an admin to edit this group")
	}

	return &GroupAdmin{
		user:  who,
		group: g,
	}, nil
}

// SetName sets the name of a group
func (ga *GroupAdmin) SetName(name string, version int) error {
	ga.group.Name = name
	err := ga.group.validate()
	if err != nil {
		return err
	}
	return ga.group.update(func() (sql.Result, error) {
		return sqlGroupUpdate.Exec(
			sql.Named("name", name),
			sql.Named("id", ga.group.ID),
			sql.Named("version", version),
		)
	})
}

// SetMember adds a new member to a group or updates an existing member
func (ga *GroupAdmin) SetMember(userID data.ID, admin bool) error {
	if !userID.Valid {
		return NewFailure("Cannot add an invalid user to a group")
	}

	currentAdmin := false
	err := sqlGroupGetMember.QueryRow(sql.Named("group_id", ga.group.ID), sql.Named("user_id", userID)).
		Scan(&currentAdmin)
	if err == nil {
		// user is already a member
		if admin == currentAdmin {
			// nothing to update
			return nil
		}
		_, err = sqlGroupUpdateMember.Exec(
			sql.Named("group_id", ga.group.ID),
			sql.Named("user_id", userID),
			sql.Named("admin", admin),
		)
		if err != nil {
			return err
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	// not currently a member
	result, err := sqlGroupInsertMember.Exec(
		sql.Named("group_id", ga.group.ID),
		sql.Named("user_id", userID),
		sql.Named("admin", admin),
	)

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return NewFailure("Cannot add an invalid user to a group")
	}

	return err
}
