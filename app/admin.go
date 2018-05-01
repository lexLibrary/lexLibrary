package app

import (
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Admin is a wrapper around User that only provides access to admin level functions
type Admin struct {
	User *User
}

// ErrNotAdmin is returned when an admin activity is attempted by a non-admin user
var ErrNotAdmin = Unauthorized("This functionality is reserved for administrators only")

// The methods defined here should simply be wrappers around the actual code doing the inserts / updates / deletes
// The idea is you can come to the admin source file to see all of the funcitonality an admin can perform, but to see
// the detail of what is actually being performed you should goto the proper source file for the functionality: settings
// users, etc

func (a *Admin) isAdmin() bool {
	return a.User != nil && a.User.Admin
}

// Setting will look for a setting that has the passed in id
func (a *Admin) Setting(id string) (Setting, error) {
	if !a.isAdmin() {
		return Setting{}, ErrNotAdmin
	}
	return settingGet(id)
}

// SetSetting updates a settings value
func (a *Admin) SetSetting(id string, value interface{}) error {
	if !a.isAdmin() {
		return ErrNotAdmin
	}
	return settingSet(nil, id, value)
}

// SetMultipleSettings sets multiple settings in the same transaction
func (a *Admin) SetMultipleSettings(settings map[string]interface{}) error {
	if !a.isAdmin() {
		return ErrNotAdmin
	}
	return settingSetMultiple(settings)
}

// SetUserActive sets the active status of the given user
func (a *Admin) SetUserActive(u *User, active bool, version int) error {
	if !a.isAdmin() {
		return ErrNotAdmin
	}
	return u.setActive(active, version)
}

// SetUserAdmin sets if a user is an Administrator or not
func (a *Admin) SetUserAdmin(u *User, admin bool, version int) error {
	if !a.isAdmin() {
		return ErrNotAdmin
	}
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
		(select count(*) as num from users) as users,
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
)

// Overview returns statistics on the current instance
func (a *Admin) Overview() (*Overview, error) {
	if !a.isAdmin() {
		return nil, ErrNotAdmin
	}

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
