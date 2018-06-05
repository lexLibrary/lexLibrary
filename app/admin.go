package app

import (
	"fmt"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Admin is a wrapper around User that only provides access to admin level functions
type Admin struct {
	user *User
}

// ErrNotAdmin is returned when an admin activity is attempted by a non-admin user
var ErrNotAdmin = Unauthorized("This functionality is reserved for administrators only")

// The methods defined here should simply be wrappers around the actual code doing the inserts / updates / deletes
// The idea is you can come to the admin source file to see all of the funcitonality an admin can perform, but to see
// the detail of what is actually being performed you should goto the proper source file for the functionality: settings
// users, etc

// Setting will look for a setting that has the passed in id
func (a *Admin) Setting(id string) (Setting, error) {
	return settingGet(id)
}

// User returns the underlying user for the admin
func (a *Admin) User() *User {
	return a.user
}

// SetSetting updates a settings value
func (a *Admin) SetSetting(id string, value interface{}) error {
	return settingSet(nil, id, value)
}

// SetMultipleSettings sets multiple settings in the same transaction
func (a *Admin) SetMultipleSettings(settings map[string]interface{}) error {
	return settingSetMultiple(settings)
}

// SetUserActive sets the active status of the given user
func (a *Admin) SetUserActive(u *User, active bool, version int) error {
	return u.setActive(active, version)
}

// SetUserAdmin sets if a user is an Administrator or not
func (a *Admin) SetUserAdmin(u *User, admin bool, version int) error {
	return u.setAdmin(admin, version)
}

// Overview is a collection of statistics and information about the LL instance
type Overview struct {
	Instance struct {
		Users     int
		Documents int
		Sessions  int
		Size      struct {
			data.SizeStats
		}
		Uptime           time.Duration
		FirstLaunch      time.Time
		Version          string
		BuildDate        time.Time
		ErrorsTotal      int
		ErrorsSinceStart int
	}
	System  sysInfo
	Runtime RuntimeInfo

	data.Config
}

var (
	sqlInstanceStats = data.NewQuery(`
		select users.num, sessions.num, documents.num, errorsTotal.num, errorsSinceStart.num
		from
		(select count(*) as num from users where active = {{TRUE}}) as users,
		(select count(*) as num from sessions where expires > {{NOW}} and valid = {{TRUE}}) as sessions,
		(select 0 as num) as documents, 
		(select count(*) num from logs) as errorsTotal,
		(select count(*) num 
			from logs where occurred >= {{arg "start"}}
		) as errorsSinceStart
	`)
	sqlInstanceInit = data.NewQuery(`
		select occurred from schema_versions where version = 0
	`)

	sqlUsersAll      = data.NewQuery(fmt.Sprintf(`select %s from users`, userPublicColumns))
	sqlUsersActive   = data.NewQuery(fmt.Sprintf(`select %s from users where active = {{TRUE}}`, userPublicColumns))
	sqlUsersLoggedIn = data.NewQuery(fmt.Sprintf(`
		select 	%s 
		from 	users u,
			sessions s
		where 	u.id = s.user_id
		and 	s.expires > {{NOW}} 
		and 	s.valid = {{TRUE}}
	`, userPublicColumns))
)

// Overview returns statistics on the current instance
func (a *Admin) Overview() (*Overview, error) {
	o := &Overview{
		Config: data.CurrentCFG(),
	}
	o.Config.DatabaseURL = "" // hide to prevent showing db password

	// Instance Stats
	err := sqlInstanceStats.QueryRow(data.Arg("start", initTime)).Scan(
		&o.Instance.Users,
		&o.Instance.Sessions,
		&o.Instance.Documents,
		&o.Instance.ErrorsTotal,
		&o.Instance.ErrorsSinceStart,
	)
	if err != nil {
		return nil, err
	}

	o.Instance.Uptime = time.Since(initTime)

	var firstLaunch time.Time
	err = sqlInstanceInit.QueryRow().Scan(&firstLaunch)
	if err != nil {
		return nil, err
	}
	o.Instance.FirstLaunch = firstLaunch

	o.Instance.Version = Version()
	o.Instance.BuildDate = BuildDate()

	// Size Stats
	size, err := data.Size()
	if err != nil {
		return nil, err
	}

	o.Instance.Size.SizeStats = size

	o.Runtime = runtimeInfo
	o.System = systemInfo()

	return o, nil
}

// UsersAll returns a list of all of the current users in Lex Library
func (a *Admin) UsersAll() ([]*PublicProfile, error) {
	var users []*PublicProfile
	rows, err := sqlUsersAll.Query()
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

// UsersActive returns only the currently Active users in Lex Library
func (a *Admin) UsersActive() ([]*PublicProfile, error) {
	var users []*PublicProfile
	rows, err := sqlUsersActive.Query()
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

// UsersLoggedIn returns a list of all of the currently loggedin users
func (a *Admin) UsersLoggedIn() ([]*PublicProfile, error) {
	var users []*PublicProfile
	rows, err := sqlUsersLoggedIn.Query()
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
