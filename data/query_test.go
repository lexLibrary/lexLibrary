// Copyright (c) 2017 Townsourced Inc.
package data_test

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"io"
	"testing"

	"github.com/lexLibrary/lexLibrary/data"
)

func TestDataTypes(t *testing.T) {
	createTable := func() {
		_, err := data.NewQuery(`
			create table data_types (
				bytes_type {{bytes}},
				datetime_type {{datetime}},
				text_type {{text}},
				citext_type {{citext}},
				varchar_type {{varchar 30}}
			)
		`).Exec()
		if err != nil {
			t.Fatalf("Error creating data_types table: %s", err)
		}
	}
	dropTable := func() {
		_, err := data.NewQuery("drop table data_types").Exec()
		if err != nil {
			t.Fatalf("Error resetting data_types table: %s", err)
		}
	}
	reset := func() {
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

	})
	t.Run("text", func(t *testing.T) {

	})
	t.Run("citext", func(t *testing.T) {

	})
	t.Run("varchar", func(t *testing.T) {

	})
	dropTable()
}
