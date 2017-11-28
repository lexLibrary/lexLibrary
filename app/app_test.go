// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/spf13/viper"
)

var flagConfigFile string

const defaultConfigFile = "./config.yaml"

func TestMain(m *testing.M) {

	// Quick env check to prevent tests from being accidentally run against real data, as tables get truncated before
	// tests run
	if os.Getenv("LLTEST") != "true" {
		log.Fatal("LLTEST environment variable is not set to 'true'.  Make sure you are not running the tests in a real environment")
	}
	flag.StringVar(&flagConfigFile, "config", "./config.yaml", "Sets the path to the configuration file. Either a .YAML, .JSON, or .TOML file")

	flag.Parse()
	cfg := struct {
		Web  web.Config
		Data data.Config
	}{
		Web:  web.Config{},
		Data: data.Config{},
	}

	viper.SetConfigFile(flagConfigFile)

	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) && flagConfigFile == defaultConfigFile {
			log.Printf("No config file found, using default values: \n %+v\n", cfg)
			// open sqlite db in memory for testing
			cfg.Data = data.Config{
				DatabaseType:       "sqlite",
				DatabaseURL:        "file::memory:?mode=memory&cache=shared",
				MaxIdleConnections: 1,
				MaxOpenConnections: 1,
			}
		} else {
			log.Fatal(err)
		}
	} else {
		viper.Unmarshal(&cfg)
	}

	// All tests assume the database is empty
	err = data.Init(cfg.Data)
	if err != nil {
		log.Fatal(err)
	}

	result := m.Run()
	err = data.Teardown()
	if err != nil {
		log.Fatalf("Error tearing down data connections: %s", err)
	}
	os.Exit(result)
}
