// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var flagConfigFile string

const defaultConfigFile = "./config.yaml"

// TestingSetup setups the database for use with running on the go test suite
func TestingSetup(m *testing.M) error {
	if m == nil {
		return fmt.Errorf("TestingSetup must be run by the testing framework")
	}
	flag.StringVar(&flagConfigFile, "config", "./config.yaml", "Sets the path to the configuration file. Either a .YAML, .JSON, or .TOML file")

	flag.Parse()
	cfg := struct {
		Web  map[string]interface{}
		Data Config
	}{
		Data: Config{},
	}

	viper.SetConfigFile(flagConfigFile)

	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) && flagConfigFile == defaultConfigFile {
			log.Printf("No config file found, using default values: \n %+v\n", cfg)
			// This is necessary, because we can't use file::memory with a single connection
			// and transactions, otherwise we deadlock
			var tempDir string
			if runtime.GOOS == "linux" {
				tempDir = "/dev/shm/"
			} else {
				tempDir = os.TempDir()
			}
			tempDBFile := filepath.Join(tempDir, "lexLibray.db")
			if _, err := os.Stat(tempDBFile); err == nil {
				//file already exists, delete file
				err = os.Remove(tempDBFile)
				if err != nil {
					return errors.Wrap(err, "deleting old temp database file")
				}
			}
			cfg.Data = Config{
				DatabaseType: "sqlite",
				// DatabaseURL:        "file::memory:?mode=memory&cache=shared",
				DatabaseFile: tempDBFile,
				// MaxIdleConnections: 1,
				// MaxOpenConnections: 1,
			}
		} else {
			return err
		}
	} else {
		// Quick env check to prevent tests from being accidentally run against real data,
		//	as tables get truncated before
		// tests run, doesn't get checked if default in-memory sqlite db is used
		if os.Getenv("LLTEST") != "true" {
			return errors.New("LLTEST environment variable is not set to 'true'.  " +
				"Make sure you are not running the tests in a real environment")
		}
		err = viper.Unmarshal(&cfg)
		if err != nil {
			return err
		}
	}

	// All tests assume the database is empty
	return Init(cfg.Data)
}

func ResetDB(t *testing.T) {
	t.Helper()
	TruncateTable(t, "logs")
	TruncateTable(t, "sessions")
	TruncateTable(t, "group_users")
	TruncateTable(t, "registration_token_groups")
	TruncateTable(t, "registration_token_users")
	TruncateTable(t, "document_groups")
	TruncateTable(t, "document_tags")
	TruncateTable(t, "document_draft_tags")
	TruncateTable(t, "document_history")
	TruncateTable(t, "document_drafts")
	TruncateTable(t, "documents")
	TruncateTable(t, "groups")
	TruncateTable(t, "settings")
	TruncateTable(t, "registration_tokens")
	TruncateTable(t, "users")
	TruncateTable(t, "images")
}

func TruncateTable(t *testing.T, table string) {
	t.Helper()
	_, err := NewQuery(fmt.Sprintf("delete from %s", table)).Exec()
	if err != nil {
		t.Fatalf("Error emptying %s table before running tests: %s", table, err)
	}
}
