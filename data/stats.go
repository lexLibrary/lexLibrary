package data

import "database/sql"

// SizeStats are statistics about the size of the data in LL
type SizeStats struct {
	Data   uint64
	Search uint64
	Image  uint64
	Total  uint64
}

var (
	sqlitePageSize = NewQuery(`
		pragma page_size;
	`)
	sqlitePages = NewQuery(`
		pragma page_count;
	`)
	postgresDBSize = NewQuery(`
		select pg_database_size({{arg "db"}})
	`)
	postgresTableSize = NewQuery(`
		select pg_total_relation_size({{arg "table"}})
	`)
	mysqlSize = NewQuery(`
		SELECT SUM(data_length + index_length) 
		FROM information_schema.tables 
		GROUP BY table_schema; 
		union all
		SELECT SUM(data_length + index_length) 
		FROM information_schema.tables 
		where table_name = {{arg "table"}}
	`)
)

// Size returns size stats on the data layer
func Size() (SizeStats, error) {
	stats := SizeStats{}
	switch dbType {
	case sqlite:
		var pages, pageSize uint64
		err := sqlitePageSize.QueryRow().Scan(&pageSize)
		if err != nil {
			return stats, err
		}
		err = sqlitePages.QueryRow().Scan(&pages)
		if err != nil {
			return stats, err
		}

		stats.Data = pages * pageSize
		stats.Image = 0
	case postgres:
		var dbSize, tableSize uint64
		err := postgresDBSize.QueryRow(sql.Named("db", databaseName)).Scan(&dbSize)
		if err != nil {
			return stats, err
		}
		err = postgresTableSize.QueryRow(sql.Named("table", "images")).Scan(&tableSize)
		if err != nil {
			return stats, err
		}
		stats.Data = dbSize - tableSize
		stats.Image = tableSize
	case mysql:
		var dbSize, tableSize uint64
		rows, err := mysqlSize.Query(sql.Named("table", "images"))
		if err != nil {
			return stats, err
		}
		defer rows.Close()
		rows.Next()
		err = rows.Scan(&dbSize)
		if err != nil {
			return stats, err
		}
		rows.Next()
		err = rows.Scan(&tableSize)
		if err != nil {
			return stats, err
		}

		stats.Data = dbSize - tableSize
		stats.Image = tableSize
	default:
	}

	stats.Search = 0 // TODO:
	stats.Total = stats.Data + stats.Image + stats.Search
	return stats, nil
}
