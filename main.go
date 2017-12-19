// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/spf13/viper"
)

const defaultConfigFile = "./config.yaml"

var flagConfigFile string
var flagDevMode bool

func init() {
	flag.StringVar(&flagConfigFile, "config", defaultConfigFile,
		"Sets the path to the configuration file. Either a .YAML, .JSON, or .TOML file")
	flag.BoolVar(&flagDevMode, "dev", false,
		"Runs Lex Library in Development mode where templates are reloaded and static files are not cached.")

	go func() {
		//Capture program shutdown, to make sure everything shuts down nicely
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			if sig == os.Interrupt {
				log.Print("Lex Library is shutting down")
				err := web.Teardown()
				if err != nil {
					log.Fatalf("Error Tearing down the Web layer: %s", err)
				}
				err = data.Teardown()
				if err != nil {
					log.Fatalf("Error Tearing down the Data layer: %s", err)
				}
				os.Exit(0)
			}
		}
	}()
}

func main() {
	flag.Parse()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LL")

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

	if flagDevMode {
		log.Println("Starting in Development mode")
	}

	err = data.Init(cfg.Data)
	if err != nil {
		log.Fatalf("Error initializing data layer: %s", err)
	}
	defer data.Teardown()

	log.Println("Data layer initialized")

	err = web.StartServer(cfg.Web, flagDevMode)
	if err != nil {
		log.Fatalf("Error initializing web server: %s", err)
	}
}
