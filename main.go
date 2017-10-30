// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/spf13/viper"
)

const defaultConfigFile = "./config.yaml"

var flagConfigFile string

func init() {
	flag.StringVar(&flagConfigFile, "config", defaultConfigFile, "Sets the path to the configuration file. Either a .YAML, .JSON, or .TOML file")
}

func main() {
	flag.Parse()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LEX")

	log.Println("Lex Library is starting up")

	log.Printf("Loading configuration from %s\n", flagConfigFile)
	if flagConfigFile == defaultConfigFile {
		log.Println("You can set the location of the config file with the -config flag")
	}
	viper.SetConfigFile(flagConfigFile)

	cfg := struct {
		Web  web.Config
		Data data.Config
	}{
		Web:  web.Config{},
		Data: data.Config{},
	}

	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) && flagConfigFile == defaultConfigFile {
			cfg.Web = web.DefaultConfig()
			cfg.Data = data.DefaultConfig()
			log.Printf("No config file found, using default values: \n %+v\n", cfg)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("Found and loaded config file %s\n", viper.ConfigFileUsed())
		viper.Unmarshal(&cfg)
	}

	err = data.Init(cfg.Data)
	if err != nil {
		log.Fatalf("Error initializing data layer: %s", err)
	}

	log.Println("Data layer initialized")

	err = web.StartServer(cfg.Web)
	if err != nil {
		log.Fatalf("Error initializing web server: %s", err)
	}

}
