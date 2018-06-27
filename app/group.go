// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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

var sqlGroup = struct {
	insert,
	userInsert,
	byName,
	search,
	member,
	update,
	updateMember,
	insertMember *data.Query
	byIDs func(ids []data.ID, count bool) (*data.Query, []data.Argument)
}{
	insert: data.NewQuery(`
		insert into groups (
			id,
			name, 
			name_search,
			version,
			updated, 
			created
		) values (
			{{arg "id"}}, 
			{{arg "name"}}, 
			{{arg "name_search"}}, 
			{{arg "version"}}, 
			{{arg "updated"}}, 
			{{arg "created"}}
		)
	`),
	userInsert: data.NewQuery(`
		insert into user_to_groups (
			user_id,
			group_id,
			admin
		) values (
			{{arg "user_id"}}, 
			{{arg "group_id"}}, 
			{{arg "admin"}}
		)
	`),
	byName: data.NewQuery(`
		select id, name, version, updated, created 
		from groups 
		where name_search = {{arg "name"}}
	`),
	search: data.NewQuery(`
		select id, name, version, updated, created 
		from groups 
		where name_search like {{arg "name"}}
		order by name_search
		{{if sqlserver}}
			OFFSET 0 ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
		{{else}}
			LIMIT {{arg "limit"}}
		{{end}}
	`),
	byIDs: func(ids []data.ID, count bool) (*data.Query, []data.Argument) {
		in := ""
		args := make([]data.Argument, len(ids))
		for i := range ids {
			if i != 0 {
				in += ", "
			}
			name := "id" + strconv.Itoa(i)
			in += fmt.Sprintf(`{{arg "%s"}}`, name)
			args[i].Name = name
			args[i].Value = ids[i]
		}
		sel := `select id, name, version, updated, created`
		if count {
			sel = `select count(id)`
		}
		return data.NewQuery(fmt.Sprintf(`
			%s 
			from groups 
			where id in (%s)
		`, sel, in)), args
	},
	member: data.NewQuery(`
		select admin 
		from user_to_groups 
		where user_id = {{arg "user_id"}} 
		and group_id = {{arg "group_id"}}
	`),
	update: data.NewQuery(`
		update groups set name = {{arg "name"}},
			name_search = {{arg "name_search"}},
			updated = {{NOW}}, 
			version = version + 1 
		where id = {{arg "id"}} 
		and version = {{arg "version"}}
	`),
	insertMember: data.NewQuery(`
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
	`),
	updateMember: data.NewQuery(`
		update user_to_groups
		set admin = {{arg "admin"}}
		where user_id = {{arg "user_id"}}
		and group_id = {{arg "group_id"}}
	`),
}

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

	_, err = u.GroupFromName(name)
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
		_, err = sqlGroup.userInsert.Tx(tx).Exec(
			data.Arg("user_id", u.ID),
			data.Arg("group_id", g.ID),
			data.Arg("admin", true),
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Group) scan(record Scanner) error {
	err := record.Scan(
		&g.ID,
		&g.Name,
		&g.Version,
		&g.Updated,
		&g.Created,
	)
	if err == sql.ErrNoRows {
		return ErrGroupNotFound
	}
	return err
}

// GroupSearch returns a list of groups that start with the name part
func (u *User) GroupSearch(namePart string) ([]*Group, error) {
	var groups []*Group

	search := "%" + strings.ToLower(namePart) + "%" // trailing matches only for now

	rows, err := sqlGroup.search.Query(data.Arg("name", search), data.Arg("limit", maxRows))
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

// GroupFromName returns a group based on the passed in name
func (u *User) GroupFromName(name string) (*Group, error) {
	g := &Group{}

	err := g.scan(sqlGroup.byName.QueryRow(data.Arg("name", strings.ToLower(name))))
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Group) validate() error {
	err := data.FieldValidate("group.name", g.Name)
	if err != nil {
		return NewFailureFromErr(err)
	}
	return nil
}

func (g *Group) insert(tx *sql.Tx) error {
	_, err := sqlGroup.insert.Tx(tx).Exec(
		data.Arg("id", g.ID),
		data.Arg("name", g.Name),
		data.Arg("name_search", strings.ToLower(g.Name)),
		data.Arg("version", g.Version),
		data.Arg("updated", g.Updated),
		data.Arg("created", g.Created),
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
	if who.IsAdmin() {
		return &GroupAdmin{
			user:  who,
			group: g,
		}, nil
	}
	admin := false
	err := sqlGroup.member.QueryRow(data.Arg("user_id", who.ID), data.Arg("group_id", g.ID)).Scan(
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
		return sqlGroup.update.Exec(
			data.Arg("name", name),
			data.Arg("name_search", strings.ToLower(name)),
			data.Arg("id", ga.group.ID),
			data.Arg("version", version),
		)
	})
}

// SetMember adds a new member to a group or updates an existing member
func (ga *GroupAdmin) SetMember(userID data.ID, admin bool) error {
	if userID.IsNil() {
		return NewFailure("Cannot add an invalid user to a group")
	}

	currentAdmin := false
	err := sqlGroup.member.QueryRow(data.Arg("group_id", ga.group.ID), data.Arg("user_id", userID)).
		Scan(&currentAdmin)
	if err == nil {
		// user is already a member
		if admin == currentAdmin {
			// nothing to update
			return nil
		}
		_, err = sqlGroup.updateMember.Exec(
			data.Arg("group_id", ga.group.ID),
			data.Arg("user_id", userID),
			data.Arg("admin", admin),
		)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}

	// not currently a member
	result, err := sqlGroup.insertMember.Exec(
		data.Arg("group_id", ga.group.ID),
		data.Arg("user_id", userID),
		data.Arg("admin", admin),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return NewFailure("Cannot add an invalid user to a group")
	}

	return err
}
