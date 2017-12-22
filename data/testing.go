// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"flag"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var flagConfigFile string

const defaultConfigFile = "./config.yaml"

func TestingSetup() error {
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
			// open sqlite db in memory for testing
			cfg.Data = Config{
				DatabaseType:       "sqlite",
				DatabaseURL:        "file::memory:?mode=memory&cache=shared",
				MaxIdleConnections: 1,
				MaxOpenConnections: 1,
			}
		} else {
			return err
		}
	} else {
		// Quick env check to prevent tests from being accidentally run against real data, as tables get truncated before
		// tests run, doesn't get checked if default in-memory sqlite db is used
		if os.Getenv("LLTEST") != "true" {
			return errors.New("LLTEST environment variable is not set to 'true'.  Make sure you are not running the tests in a real environment")
		}
		viper.Unmarshal(&cfg)
	}

	// All tests assume the database is empty
	return Init(cfg.Data)
}
