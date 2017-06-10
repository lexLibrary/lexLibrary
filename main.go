// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/timshannon/lexLibrary/data"
	"github.com/timshannon/lexLibrary/web"
)

const appName = "lexLibrary"

func main() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LEX")
	viper.SetConfigName("config")

	fmt.Println("Lex Library is starting up")
	fmt.Println("Looking for config.yaml, config.toml, or config.json in the following locations: ")
	for _, location := range configLocations("lexLibrary") {
		viper.AddConfigPath(location)
		fmt.Println("\t" + location)
	}

	cfg := struct {
		Web  web.Config
		Data data.Config
	}{
		Web:  web.Config{},
		Data: data.Config{},
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// create default file
			cfg.Web = web.DefaultConfig()
		} else {
			log.Fatal(err)
		}
	}

	viper.Unmarshal(&cfg)

	if cfg.Data.DatabaseFile == "" {
		cfg.Data.DatabaseFile = getDataFile("lexLibrary.db")
	}

	if cfg.Data.SearchFile == "" {
		cfg.Data.SearchFile = getDataFile("lexLibrary.search")
	}

	err = data.Init(cfg.Data)
	if err != nil {
		log.Fatalf("Error initializing data layer: %s", err)
	}

	err = web.Init(cfg.Web)
	if err != nil {
		log.Fatalf("Error initializing web server: %s", err)
	}

}

// getDataFile will return the first data file it finds, or it will return the default location
func getDataFile(defaultLocation string) string {
	//FIXME
	locations := dataLocations(appName)

	file := ""

	for _, location := range locations {
		file = filepath.Join(location, filename)

		_, err := os.Stat(file)
		if err == nil {
			return file
		}
	}

	return file
}
