// Copyright (c) 2017-2018 Townsourced Inc.
package browser

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/tebeka/selenium"
)

const (
	defaultTimeout = 2 * time.Second
)

var driver selenium.WebDriver
var uri = "http://localhost:8080"

func TestMain(m *testing.M) {
	err := data.TestingSetup()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Init()
	if err != nil {
		log.Fatalf("Error initializing app layer: %s", err)
	}

	hostname := "localhost"
	port := 8080

	if os.Getenv("LLHOST") != "" {
		hostname = os.Getenv("LLHOST")
	}

	if os.Getenv("LLPORT") != "" {
		port, err = strconv.Atoi(os.Getenv("LLPORT"))
		if err != nil {
			port = 8080
		}
	}

	webCFG := web.DefaultConfig()
	webCFG.Port = port
	uri = fmt.Sprintf("http://%s:%d", hostname, port)

	go func() {
		err = web.StartServer(webCFG, false)
		if err != nil {
			log.Fatalf("Error initializing web server: %s", err)
		}
	}()

	driver, err = startWebDriver()
	if err != nil {
		log.Fatalf("Error starting web driver: %s", err)
	}

	result := m.Run()

	err = driver.Quit()
	if err != nil {
		log.Fatalf("Failed to close pages and stop WebDriver: %s", err)
	}

	err = data.Teardown()
	if err != nil {
		log.Fatalf("Error tearing down data connections: %s", err)
	}
	os.Exit(result)
}

func startWebDriver() (selenium.WebDriver, error) {
	browser := os.Getenv("LLBROWSER")
	webDriverURL := "http://localhost:4444/wd/hub"
	if os.Getenv("LLWEBDRIVERURL") != "" {
		webDriverURL = os.Getenv("LLWEBDRIVERURL")
	}
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": browser}
	wd, err := selenium.NewRemote(caps, webDriverURL)
	if err != nil {
		return nil, err
	}
	err = wd.SetAsyncScriptTimeout(defaultTimeout)
	if err != nil {
		return nil, err
	}

	err = wd.SetPageLoadTimeout(defaultTimeout)
	if err != nil {
		return nil, err
	}

	err = wd.SetImplicitWaitTimeout(defaultTimeout)
	if err != nil {
		return nil, err
	}

	return wd, nil
}
