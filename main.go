// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"

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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := struct {
		Web  web.Config
		Data data.Config
	}{
		Web:  web.DefaultConfig(),
		Data: data.DefaultConfig(),
	}

	viper.SetDefault("Web", cfg.Web)
	viper.SetDefault("Data", cfg.Data)
	setDefaultSubKeys("web.", cfg.Web)
	setDefaultSubKeys("data.", cfg.Data)

	log.Println("Lex Library is starting up")

	log.Printf("Loading configuration from %s\n", flagConfigFile)
	if flagConfigFile == defaultConfigFile {
		log.Println("You can set the location of the config file with the -config flag")
	}
	viper.SetConfigFile(flagConfigFile)

	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) && flagConfigFile == defaultConfigFile {
			log.Printf("No config file found, using default values: \n %+v\n", cfg)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("Found and loaded config file %s\n", viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
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

// necessary work around for the fact that viper doesn't seem to handle environment overides for nested
// keys when unmarshaled. TODO: update this with an issue link
func setDefaultSubKeys(prefix string, cfg interface{}) {
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)

	for i := 0; i < t.NumField(); i++ {
		viper.SetDefault(prefix+t.Field(i).Name, v.FieldByName(t.Field(i).Name).Interface())
	}
}
