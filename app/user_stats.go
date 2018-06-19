// Copyright (c) 2017-2018 Townsourced Inc.

package app

import "github.com/lexLibrary/lexLibrary/data"

// UserStats are statistics for a specific user
type UserStats struct {
	DocumentsRead    int
	DocumentsWritten int
	Comments         int
}

var (
	//TODO: Write some actual statistics
	sqlUserStats = data.NewQuery(`
		select 	0 as documentsRead,
			0 as documentsWritten,
			0 as comments
	`)
)

// Stats returns the user's statistics
func (u *PublicProfile) Stats() (UserStats, error) {
	stats := UserStats{}
	err := sqlUserStats.QueryRow().Scan(
		&stats.DocumentsRead,
		&stats.DocumentsWritten,
		&stats.Comments,
	)
	if err != nil {
		return UserStats{}, err
	}
	return stats, nil
}
