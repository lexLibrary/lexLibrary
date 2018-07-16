// Copyright (c) 2017-2018 Townsourced Inc.
package data_test

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

func TestDataTypes(t *testing.T) {
	createTable := func(t *testing.T) {
		_, err := data.NewQuery(`
			create table data_types (
				bytes_type {{bytes}},
				datetime_type {{datetime}},
				text_type {{text}},
				varchar_type {{varchar 30}},
				int_type {{int}},
				bool_type {{bool}},
				id_type {{id}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error creating data_types table: %s", err)
		}
	}
	dropTable := func(t *testing.T) {
		_, err := data.NewQuery("drop table data_types").Exec()
		if err != nil {
			t.Fatalf("Error resetting data_types table: %s", err)
		}
	}
	reset := func(t *testing.T) {
		dropTable(t)
		createTable(t)
	}
	createTable(t)

	t.Run("bytes", func(t *testing.T) {
		reset(t)
		expected := []byte("test data string to be compressed and stored in a field in the database")
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)

		_, err := zw.Write(expected)
		if err != nil {
			t.Fatal(err)
		}

		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		_, err = data.NewQuery(`insert into data_types (bytes_type) values ({{arg "bytes_type"}})`).
			Exec(data.Arg("bytes_type", buf.Bytes()))
		if err != nil {
			t.Fatalf("Error inserting bytes data: %s", err)
		}

		var results []byte
		var got bytes.Buffer

		err = data.NewQuery(`select bytes_type from data_types`).QueryRow().Scan(&results)
		if err != nil {
			t.Fatalf("Error retrieving bytes data: %s", err)
		}

		zr, err := gzip.NewReader(bytes.NewBuffer(results))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := io.Copy(&got, zr); err != nil {
			t.Fatal(err)
		}

		if err := zr.Close(); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(expected, got.Bytes()) {
			t.Fatalf("Bytes results from table does not match.  Expected '%s', got '%s'", expected, got.Bytes())
		}

	})
	t.Run("datetime", func(t *testing.T) {
		reset(t)
		expected := time.Now()

		expectedRound := expected.Round(time.Millisecond)
		expectedTruncate := expected.Truncate(time.Millisecond)

		_, err := data.NewQuery(`insert into data_types (datetime_type) values ({{arg "datetime_type"}})`).
			Exec(data.Arg("datetime_type", expected))
		if err != nil {
			t.Fatalf("Error inserting datetime type: %s", err)
		}

		var got time.Time
		err = data.NewQuery("Select datetime_type from data_types").QueryRow().Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving datetime type: %s", err)
		}

		gotRound := got.Round(time.Millisecond)
		gotTruncate := got.Truncate(time.Millisecond)

		// Databases will either round or truncate, so depending on how close to the nearest millisecond it was
		// it may be up or down one
		if !expected.Equal(got) && !expectedTruncate.Equal(gotTruncate) && !expectedRound.Equal(gotRound) {
			t.Fatalf("datetime type does not match expected %v, got %v", expected, got)
		}
	})
	t.Run("datetime overflow", func(t *testing.T) {
		reset(t)

		expected, err := time.Parse("2006-01-02 15:04:05", "9999-12-31 23:59:59.9")
		expected.Round(time.Millisecond)
		if err != nil {
			t.Fatalf("Error parsing overflow date: %s", err)
		}

		_, err = data.NewQuery(`insert into data_types (datetime_type) values ({{arg "datetime_type"}})`).
			Exec(data.Arg("datetime_type", expected))
		if err != nil {
			t.Fatalf("Error inserting datetime type: %s", err)
		}

		var got time.Time
		err = data.NewQuery("Select datetime_type from data_types").QueryRow().Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving datetime type: %s", err)
		}

		if !expected.Equal(got) {
			t.Fatalf("datetime type does not match expected %v, got %v", expected, got)
		}

	})
	caseTest := func(t *testing.T, columnType string, expected string) {
		t.Helper()
		reset(t)
		col := columnType + "_type"
		_, err := data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(data.Arg(col, expected))

		if err != nil {
			t.Fatalf("Error inserting case sensitive text: %s", err)
		}

		_, err = data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(data.Arg(col, strings.ToLower(expected)))
		if err != nil {
			t.Fatalf("Error inserting lowered string: %s", err)
		}

		_, err = data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(data.Arg(col, strings.ToUpper(expected)))
		if err != nil {
			t.Fatalf("Error inserting uppered string: %s", err)
		}

		got := ""
		err = data.NewQuery(`select ` + col + ` from data_types where ` + col + ` = {{arg "value"}}`).QueryRow(
			data.Arg("value", expected)).Scan(&got)

		if err != nil {
			t.Fatalf("Error retrieving case sensitive text: %s", err)
		}
		if expected != got {
			t.Fatalf("Could not retrieve equal case sensitive values. Expected %s got %s", expected, got)
		}

		count := 0

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` = {{arg "` + col + `"}}`).
			QueryRow(data.Arg(col, expected)).Scan(&count)
		if err != nil {
			t.Fatalf("Error testing sql equality for case: %s", err)
		}

		if count != 1 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 1, count)
		}

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` <> {{arg "` + col + `"}}`).
			QueryRow(data.Arg(col, expected)).Scan(&count)

		if err != nil {
			t.Fatalf("Error testing sql inequality for case: %s", err)
		}

		if count != 2 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 0, count)
		}
	}
	utf8Test := func(t *testing.T, columnType string, expected string) {
		t.Helper()
		reset(t)
		col := columnType + "_type"
		_, err := data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(data.Arg(col, expected))

		if err != nil {
			t.Fatalf("Error inserting unicode text: %s", err)
		}

		got := ""
		err = data.NewQuery(`select ` + col + ` from data_types`).QueryRow().Scan(&got)

		if err != nil {
			t.Fatalf("Error retrieving unicode text: %s", err)
		}
		if expected != got {
			t.Fatalf("Could not retrieve equal unicode values. Expected %s got %s", expected, got)
		}

		count := 0

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` = {{arg "` + col + `"}}`).
			QueryRow(data.Arg(col, expected)).Scan(&count)
		if err != nil {
			t.Fatalf("Error testing sql equality for unicode: %s", err)
		}

		if count != 1 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 1, count)
		}

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` <> {{arg "` + col + `"}}`).
			QueryRow(data.Arg(col, expected)).Scan(&count)

		if err != nil {
			t.Fatalf("Error testing sql inequality for unicode: %s", err)
		}

		if count != 0 {
			t.Fatalf("Unicode is not propery equal in the database. Expected %d, got %d", 0, count)
		}
	}

	t.Run("text unicode", func(t *testing.T) {
		utf8Test(t, "text", "♻⛄♪")
	})
	t.Run("text case sensitivity", func(t *testing.T) {
		caseTest(t, "text", "CaseSEnsitiveStrIng")
	})
	t.Run("varchar unicode", func(t *testing.T) {
		utf8Test(t, "varchar", "♻⛄♪")
	})
	t.Run("varchar", func(t *testing.T) {
		caseTest(t, "varchar", "CaseSEnsitiveStrIng")
	})

	testInt := func(t *testing.T, expected int) {
		reset(t)
		t.Helper()

		_, err := data.NewQuery(`insert into data_types (int_type) values ({{arg "int_type"}})`).
			Exec(data.Arg("int_type", expected))
		if err != nil {
			t.Fatalf("Error inserting int type: %s", err)
		}

		var got int
		err = data.NewQuery("Select int_type from data_types").QueryRow().Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving int type: %s", err)
		}

		if expected != got {
			t.Fatalf("int type does not match expected %v, got %v", expected, got)
		}

	}

	t.Run("int", func(t *testing.T) {
		testInt(t, 32)
	})

	t.Run("int max", func(t *testing.T) {
		testInt(t, math.MaxInt64)
	})

	t.Run("int negative", func(t *testing.T) {
		testInt(t, -1*math.MaxInt64)
	})

	t.Run("bool", func(t *testing.T) {
		reset(t)
		expected := true

		_, err := data.NewQuery(`insert into data_types (bool_type) values ({{arg "bool_type"}})`).
			Exec(data.Arg("bool_type", expected))
		if err != nil {
			t.Fatalf("Error inserting bool type: %s", err)
		}

		var got bool
		err = data.NewQuery("Select bool_type from data_types").QueryRow().Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving bool type: %s", err)
		}

		if expected != got {
			t.Fatalf("bool type does not match expected %v, got %v", expected, got)
		}
	})

	t.Run("id", func(t *testing.T) {
		reset(t)
		expected := data.NewID()

		_, err := data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).
			Exec(data.Arg("id_type", expected))
		if err != nil {
			t.Fatalf("Error inserting id type: %s", err)
		}
		_, err = data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).
			Exec(data.Arg("id_type", data.NewID()))
		if err != nil {
			t.Fatalf("Error inserting id type: %s", err)
		}

		var got data.ID
		err = data.NewQuery(`Select id_type from data_types where id_type = {{arg "id_type"}}`).QueryRow(
			data.Arg("id_type", expected)).Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving id type: %s", err)
		}

		if expected != got {
			t.Fatalf("id type does not match expected %v, got %v", expected, got)
		}

		err = data.NewQuery(`Select id_type from data_types where id_type <> {{arg "id_type"}}`).QueryRow(
			data.Arg("id_type", expected)).Scan(&got)
		if err != nil {
			t.Fatalf("Error retrieving id type: %s", err)
		}

		if expected == got {
			t.Fatalf("id type matches wrong value")
		}
	})

	t.Run("Transaction", func(t *testing.T) {
		reset(t)
		id1 := data.NewID()
		id2 := data.NewID()
		var otherId data.ID
		var errTest = errors.New("Rollback transaction for test")
		err := data.BeginTx(func(tx *sql.Tx) error {
			_, err := data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).Tx(tx).
				Exec(data.Arg("id_type", id1))

			if err != nil {
				return err
			}
			err = data.NewQuery(`select id_type from data_types where id_type = {{arg "id_type"}}`).Tx(tx).
				QueryRow(data.Arg("id_type", id1)).Scan(&otherId)
			if err != nil {
				return err
			}

			_, err = data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).Tx(tx).
				Exec(data.Arg("id_type", id2))
			if err != nil {
				return err
			}

			return errTest
		})

		if err == nil {
			t.Fatalf("Transaction rollback did not return an error")
		}

		if err != errTest {
			t.Fatalf("An unexpected error occurred during transaction rollback: %s", err)
		}

		err = data.NewQuery(`select id_type from data_types where id_type = {{arg "id_type"}}`).
			QueryRow(data.Arg("id_type", id1)).Scan(&otherId)
		if err != sql.ErrNoRows {
			t.Fatalf("Rows were returned after transaction rollback.")
		}

	})

	testDefaultNull := func(t *testing.T) {
		id := data.NewID()
		var datetime_type data.NullTime
		var text_type sql.NullString
		var varchar_type sql.NullString
		var int_type sql.NullInt64
		var bool_type sql.NullBool
		var id_type data.ID
		_, err := data.NewQuery(`insert into data_types (id) values ({{arg "id"}})`).
			Exec(data.Arg("id", id))
		if err != nil {
			t.Fatalf("Error inserting data for null /default testing: %s", err)
		}
		sel := func(t *testing.T) {
			err = data.NewQuery(`select datetime_type, text_type, varchar_type, int_type, bool_type, id_type
				from data_types where id = {{arg "id"}}`).QueryRow(data.Arg("id", id)).Scan(
				&datetime_type,
				&text_type,
				&varchar_type,
				&int_type,
				&bool_type,
				&id_type,
			)
			if err != nil {
				t.Fatalf("Error selecting empty entries for default / null testing: %s", err)
			}

			if !datetime_type.Time.IsZero() {
				t.Fatalf("Empty time value was not Zero time: %s", datetime_type.Time)
			}

			if text_type.String != "" {
				t.Fatalf("Incorrect empty type for text. Expected '%s' got '%s'", "", text_type.String)
			}
			if varchar_type.String != "" {
				t.Fatalf("Incorrect empty type for varchar. Expected '%s' got '%s'", "", varchar_type.String)
			}
			if int_type.Int64 != 0 {
				t.Fatalf("Incorrect empty type for int. Expected %d got %d", 0, int_type.Int64)
			}
			if bool_type.Bool {
				t.Fatalf("Incorrect empty type for boolean. Expected %v got %v", false, bool_type.Bool)
			}

			if !id_type.IsNil() {
				t.Fatalf("Incorrect empty type for id. Expected %v got %v", data.ID{}, id_type)
			}

		}
		sel(t)
		dropTable(t)
		_, err = data.NewQuery(`
			create table data_types (
				id {{id}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error dropping data columns: %s", err)
		}
		_, err = data.NewQuery(`insert into data_types (id) values ({{arg "id"}})`).
			Exec(data.Arg("id", id))
		if err != nil {
			t.Fatalf("Error inserting data for null /default testing: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add datetime_type {{datetime}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add text_type {{text}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add varchar_type {{varchar 30}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add int_type {{int}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add bool_type {{bool}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		_, err = data.NewQuery(`alter table data_types add id_type {{id}}`).Exec()
		if err != nil {
			t.Fatalf("Error adding data_types columns: %s", err)
		}

		sel(t)

	}

	t.Run("Defaults", func(t *testing.T) {
		dropTable(t)
		// create table with defaults
		// note mysql doesn't support defaults on text
		_, err := data.NewQuery(`
			create table data_types (
				id {{id}},
				datetime_type {{datetime}} NOT NULL DEFAULT '{{defaultDateTime}}',
				text_type {{text}},
				varchar_type {{varchar 30}} NOT NULL DEFAULT '',
				int_type {{int}} NOT NULL DEFAULT 0,
				bool_type {{bool}} NOT NULL DEFAULT {{FALSE}},
				id_type {{id}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error creating data_types table: %s", err)
		}
		testDefaultNull(t)
	})

	t.Run("Nulls", func(t *testing.T) {
		dropTable(t)
		// create table with nulls
		_, err := data.NewQuery(`
			create table data_types (
				id {{id}},
				datetime_type {{datetime}},
				text_type {{text}},
				varchar_type {{varchar 30}},
				int_type {{int}},
				bool_type {{bool}},
				id_type {{id}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error creating data_types table: %s", err)
		}
		testDefaultNull(t)
	})

	t.Run("Now", func(t *testing.T) {
		reset(t)
		round := 2 * time.Second
		other := time.Now().Round(round)
		_, err := data.NewQuery(`insert into data_types (datetime_type) values ({{NOW}})`).Exec()
		if err != nil {
			t.Fatalf("Error inserting now record: %s", err)
		}
		var now time.Time
		err = data.NewQuery(`select datetime_type from data_types`).QueryRow().Scan(&now)
		if err != nil {
			t.Fatalf("Error retrieving now record: %s", err)
		}

		if !now.Round(round).Equal(other) {
			t.Fatalf("Now func did not return an accurate now value.  Expected %s, got %s", other, now)
		}
	})

	t.Run("In Queries", func(t *testing.T) {
		reset(t)

		ids := make([]data.ID, 10)

		for i := range ids {
			ids[i] = data.NewID()
			_, err := data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).
				Exec(data.Arg("id_type", ids[i]))
			if err != nil {
				t.Fatal(err)
			}
		}

		tests := []struct {
			name   string
			query  string
			args   []data.Argument
			result []data.ID
		}{
			{"simple in", `select id_type from data_types where id_type in ({{args "ids"}})`,
				data.Args("ids", ids[3:7]), ids[3:7]},
			{"mulitple args + trailing in", `
				select id_type from data_types 
				where id_type <> {{arg "not_id"}}
				and id_type in ({{args "ids"}})
			`,
				append(data.Args("ids", ids[3:7]), data.Arg("not_id", ids[4])),
				[]data.ID{ids[3], ids[5], ids[6]},
			},
			{"mulitple args + leading in out of order", `
				select id_type from data_types 
				where id_type in ({{args "ids"}})
				and id_type <> {{arg "not_id"}}
			`,
				append(data.Args("ids", ids[3:7]), data.Arg("not_id", ids[4])),
				[]data.ID{ids[3], ids[5], ids[6]},
			},
			{"single in", `select id_type from data_types where id_type in ({{args "ids"}})`,
				data.Args("ids", ids[3:4]), ids[3:4]},
			{"mulitple in args", `
				select id_type from data_types
				where id_type in ({{args "ids"}})
				and id_type not in ({{args "not_ids"}})
			`,
				append(data.Args("ids", ids[3:7]), data.Args("not_ids", ids[4:6])...),
				[]data.ID{ids[3], ids[6]},
			},
			{"mulitple args between two ins out of order", `
				select id_type from data_types
				where id_type in ({{args "ids"}})
				and id_type <> {{arg "single"}}
				and id_type not in ({{args "not_ids"}})
			`,
				append(append(data.Args("ids", ids[3:7]), data.Args("not_ids", ids[4:6])...),
					data.Arg("single", ids[6])),
				[]data.ID{ids[3]},
			},
			{"mulitple in args out of order", `
				select id_type from data_types
				where id_type not in ({{args "not_ids"}})
				and id_type in ({{args "ids"}})
			`,
				append(data.Args("ids", ids[3:7]), data.Args("not_ids", ids[4:6])...),
				[]data.ID{ids[3], ids[6]},
			},
		}

		for _, test := range tests {
			rows, err := data.NewQuery(test.query).
				Query(test.args...)

			if err != nil {
				t.Fatalf("Error executing test %s Query:%s: %s", test.name, test.query, err)
			}

			var result []data.ID
			for rows.Next() {
				var id data.ID

				err = rows.Scan(&id)
				if err != nil {
					t.Fatalf("Error executing test %s Query:%s: %s", test.name, test.query, err)
				}
				result = append(result, id)
			}

			if len(test.result) != len(result) {
				rows.Close()
				t.Fatalf("Test: %s Result length is incorrect. Expected %d, got %d\nQuery: %s",
					test.name, len(test.result), len(result), test.query)
			}

			for i := range test.result {
				found := false
				for j := range result {
					if result[j] == test.result[i] {
						found = true
						break
					}
				}
				if !found {
					rows.Close()
					t.Fatalf("Test: %s ID not found in result set: %s\nQuery: %s",
						test.name, test.result[i], test.query)
				}
			}
		}

	})
	t.Run("In Queries Side effects", func(t *testing.T) {
		reset(t)

		ids := make([]data.ID, 10)

		for i := range ids {
			ids[i] = data.NewID()
			_, err := data.NewQuery(`insert into data_types (id_type) values ({{arg "id_type"}})`).
				Exec(data.Arg("id_type", ids[i]))
			if err != nil {
				t.Fatal(err)
			}
		}

		qry := data.NewQuery(`select id_type from data_types where id_type in ({{args "ids"}})`)

		tests := []struct {
			args   []data.Argument
			result []data.ID
		}{
			{data.Args("ids", ids[:]), ids[:]},
			{data.Args("ids", ids[7:8]), ids[7:8]},
			{data.Args("ids", ids[3:7]), ids[3:7]},
			{data.Args("ids", ids[1:3]), ids[1:3]},
			{data.Args("ids", ids[1:2]), ids[1:2]},
		}

		for i, test := range tests {
			rows, err := qry.Query(test.args...)

			if err != nil {
				t.Fatal(err)
			}

			var result []data.ID
			for rows.Next() {
				var id data.ID

				err = rows.Scan(&id)
				if err != nil {
					t.Fatalf("Error executing test %d: %s", i, err)
				}
				result = append(result, id)
			}

			if len(test.result) != len(result) {
				rows.Close()
				t.Fatalf("Test: %d Result length is incorrect. Expected %d, got %d",
					i, len(test.result), len(result))
			}

			for i := range test.result {
				found := false
				for j := range result {
					if result[j] == test.result[i] {
						found = true
						break
					}
				}
				if !found {
					rows.Close()
					t.Fatalf("Test: %d ID not found in result set: %s\n",
						i, test.result[i])
				}
			}
		}

	})

	t.Run("In Queries with 0 args", func(t *testing.T) {
		reset(t)
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("Running an in query without arguments did not panic")
			}
		}()

		data.NewQuery(`select * from table where field in ({{args "args"}})`).
			QueryRow(data.Args("args", []int{})...)

		data.NewQuery(`select * from table where field in ({{args "args"}}) and id = {{arg "id"}}`).
			QueryRow(append(data.Args("args", []int{}), data.Arg("id", 1))...)
	})
	dropTable(t)
}

//TODO: Test Count and Page
