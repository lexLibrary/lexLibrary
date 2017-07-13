// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"log"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/spf13/viper"
)

const appName = "lexLibrary"

func main() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LEX")
	viper.SetConfigName("config")

	log.Println("Lex Library is starting up")
	msg := "Looking for config.yaml, config.toml, or config.json in the following locations:"
	for _, location := range configLocations("lexLibrary") {
		viper.AddConfigPath(location)
		msg += "\n\t" + location
	}

	log.Println(msg)

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
