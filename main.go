// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
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

	err := viper.ReadInConfig()
	if err != nil && os.IsNotExist(err) {
		log.Fatal(err)
	}

}

// getDataFile will return the first data file it finds, or it will return the last of the available locations
func getDataFile(filename string) string {
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
