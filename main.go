// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"flag"
	"log"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/spf13/viper"
)

const appName = "lexLibrary"
const defaultConfigFile = "./config.yaml"

var flagConfigFile string

func init() {
	flag.StringVar(&flagConfigFile, "config", defaultConfigFile, "Sets the path to the configuration file. Either a .YAML, .JSON, or .TOML file")
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LEX")
	viper.SetConfigName("config")

	log.Println("Lex Library is starting up")

	log.Printf("Loading configuration from %s\n", *flagConfigFile)
	if flagConfigFile == defaultConfigFile {
		log.Println("You can set the location of the config file with the -config flag")
	}

	cfg := struct {
		Web  web.Config
		Data app.Config
	}{
		Web:  web.Config{},
		Data: app.Config{},
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using default values")
			cfg.Web = web.DefaultConfig()
			cfg.Data = app.DefaultConfig()
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("Found and loaded config file %s\n", viper.ConfigFileUsed())
		viper.Unmarshal(&cfg)
	}

	err = app.Init(cfg.Data)
	if err != nil {
		log.Fatalf("Error initializing data layer: %s", err)
	}

	log.Println("Data layer initialized")

	err = web.StartServer(cfg.Web)
	if err != nil {
		log.Fatalf("Error initializing web server: %s", err)
	}

}
