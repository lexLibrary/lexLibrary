// Copyright (c) 2017 Townsourced Inc.
package data_test

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

func TestDataTypes(t *testing.T) {
	createTable := func() {
		t.Helper()
		_, err := data.NewQuery(`
			create table data_types (
				bytes_type {{bytes}},
				datetime_type {{datetime}},
				text_type {{text}},
				varchar_type {{varchar 30}},
				int_type {{int}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error creating data_types table: %s", err)
		}
	}
	dropTable := func() {
		t.Helper()
		_, err := data.NewQuery("drop table data_types").Exec()
		if err != nil {
			t.Fatalf("Error resetting data_types table: %s", err)
		}
	}
	reset := func() {
		t.Helper()
		dropTable()
		createTable()
	}
	createTable()

	t.Run("bytes", func(t *testing.T) {
		reset()
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
			Exec(sql.Named("bytes_type", buf.Bytes()))
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
		reset()

		expected := time.Now().Round(time.Second)

		_, err := data.NewQuery(`insert into data_types (datetime_type) values ({{arg "datetime_type"}})`).
			Exec(sql.Named("datetime_type", expected))
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
		reset()
		col := columnType + "_type"
		_, err := data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(sql.Named(col, expected))

		if err != nil {
			t.Fatalf("Error inserting case sensitive text: %s", err)
		}

		_, err = data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(sql.Named(col, strings.ToLower(expected)))
		if err != nil {
			t.Fatalf("Error inserting lowered string: %s", err)
		}

		_, err = data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(sql.Named(col, strings.ToUpper(expected)))
		if err != nil {
			t.Fatalf("Error inserting uppered string: %s", err)
		}

		got := ""
		err = data.NewQuery(`select ` + col + ` from data_types where ` + col + ` = {{arg "value"}}`).QueryRow(
			sql.Named("value", expected)).Scan(&got)

		if err != nil {
			t.Fatalf("Error retrieving case sensitive text: %s", err)
		}
		if expected != got {
			t.Fatalf("Could not retrieve equal case sensitive values. Expected %s got %s", expected, got)
		}

		count := 0

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` = {{arg "` + col + `"}}`).
			QueryRow(sql.Named(col, expected)).Scan(&count)
		if err != nil {
			t.Fatalf("Error testing sql equality for case: %s", err)
		}

		if count != 1 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 1, count)
		}

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` <> {{arg "` + col + `"}}`).
			QueryRow(sql.Named(col, expected)).Scan(&count)

		if err != nil {
			t.Fatalf("Error testing sql inequality for case: %s", err)
		}

		if count != 2 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 0, count)
		}
	}
	utf8Test := func(t *testing.T, columnType string, expected string) {
		t.Helper()
		reset()
		col := columnType + "_type"
		_, err := data.NewQuery(`insert into data_types (` + col + `) values ({{arg "` + col + `"}})`).
			Exec(sql.Named(col, expected))

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
			QueryRow(sql.Named(col, expected)).Scan(&count)
		if err != nil {
			t.Fatalf("Error testing sql equality for unicode: %s", err)
		}

		if count != 1 {
			t.Fatalf("Case is not propery equal in the database. Expected %d, got %d", 1, count)
		}

		err = data.NewQuery(`select count(*) from data_types where ` + col + ` <> {{arg "` + col + `"}}`).
			QueryRow(sql.Named(col, expected)).Scan(&count)

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

	dropTable()
}
