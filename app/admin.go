package app

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
