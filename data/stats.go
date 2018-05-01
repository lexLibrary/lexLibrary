package data

// SizeStats are statistics about the size of the data in LL
// all sizes are in bytes
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
	postgresSize = NewQuery(`
		select pg_database_size({{arg "db"}})
		union all
		select pg_total_relation_size({{arg "table"}})
	`)
	mysqlSize = NewQuery(`
		SELECT SUM(data_length + index_length) 
		FROM information_schema.tables 
		union all
		SELECT SUM(data_length + index_length) 
		FROM information_schema.tables 
		where table_name = {{arg "table"}}
	`)
	sqlserverSize = NewQuery(`
		SELECT size * 8 * 1000 
		FROM sys.database_files
		union all
		SELECT
		SUM(a.total_pages) * 8 * 1000
		FROM sys.tables t
		INNER JOIN sys.indexes i ON t.OBJECT_ID = i.object_id
		INNER JOIN sys.partitions p ON i.object_id = p.OBJECT_ID AND i.index_id = p.index_id
		INNER JOIN sys.allocation_units a ON p.partition_id = a.container_id
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		where t.name = {{arg "table"}}
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
	case postgres:
		var dbSize, tableSize uint64
		rows, err := postgresSize.Query(Arg("table", "images"), Arg("db", databaseName))
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
	case mysql:
		var dbSize, tableSize uint64
		rows, err := mysqlSize.Query(Arg("table", "images"))
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
	case sqlserver:
		var dbSize, tableSize uint64
		rows, err := sqlserverSize.Query(Arg("table", "images"))
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
	case cockroachdb:
	default:
		panic("Unsupported database type")
	}

	stats.Search = 0 // TODO:
	stats.Total = stats.Data + stats.Image + stats.Search
	return stats, nil
}
