// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
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
	byIDs,
	byIDsTotal,
	insertMember *data.Query
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
		insert into group_users (
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
	byIDsTotal: data.NewQuery(`
		select count(id)
		from groups 
		where id in ({{args "ids"}})
	`),
	byIDs: data.NewQuery(`
		select id, name, version, updated, created
		from groups 
		where id in ({{args "ids"}})
	`),
	member: data.NewQuery(`
		select admin 
		from group_users 
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
		insert into group_users (
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
		update group_users
		set admin = {{arg "admin"}}
		where user_id = {{arg "user_id"}}
		and group_id = {{arg "group_id"}}
	`),
}

var (
	// ErrGroupNotFound is returned when a group couldn't be found
	ErrGroupNotFound = NotFound("Group not found")
	// ErrGroupConflict occurs when someone updates an older version of a group
	ErrGroupConflict = Conflict("You are not editing the most current version of this group. Please refresh and " +
		"try again.")
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

// func groupsFromIDs(ids ...data.ID) ([]*Group, error) {
// 	groups := make([]*Group, 0, len(ids))

// 	if len(ids) == 0 {
// 		return groups, nil
// 	}

// 	rows, err := sqlGroup.byIDs.Query(data.Args("id", ids)...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		g := &Group{}
// 		err = g.scan(rows)
// 		if err != nil {
// 			return nil, err
// 		}
// 		groups = append(groups, g)
// 	}

// 	return groups, nil
// }

func validateGroups(groupIDs ...data.ID) error {
	if len(groupIDs) == 0 {
		return nil
	}
	groupCount := 0
	err := sqlGroup.byIDsTotal.QueryRow(data.Args("ids", groupIDs)...).Scan(&groupCount)
	if err != nil {
		return err
	}

	if groupCount != len(groupIDs) {
		// one or more groups were not found
		return NewFailure("One or more of the groups are invalid")
	}
	return nil
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

// AddMember adds a new member to a group, if user is already an member and an admin they are demoted
// to normal member
func (ga *GroupAdmin) AddMember(userID data.ID) error {
	return ga.setMember(userID, false)
}

// AddAddmin adds a new member to a group as an admin or updates an existing member to admin
func (ga *GroupAdmin) AddAdmin(userID data.ID) error {
	return ga.setMember(userID, true)
}

func (ga *GroupAdmin) setMember(userID data.ID, admin bool) error {
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
