// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gitlab.com/lexLibrary/lexLibrary/data"
	"gitlab.com/lexLibrary/lexLibrary/web"
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
			fmt.Println("No config file found, using default values")
			cfg.Web = web.DefaultConfig()
			cfg.Data = data.DefaultConfig()
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Found and loaded config file %s", viper.ConfigFileUsed())
	}

	viper.Unmarshal(&cfg)

	err = data.Init(cfg.Data)
	if err != nil {
		log.Fatalf("Error initializing data layer: %s", err)
	}

	err = web.Init(cfg.Web)
	if err != nil {
		log.Fatalf("Error initializing web server: %s", err)
	}

}
