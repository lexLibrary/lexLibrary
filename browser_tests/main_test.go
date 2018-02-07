package browser_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/sclevine/agouti"
)

var page *agouti.Page
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

	webCFG := web.DefaultConfig()
	webCFG.Port = 8090
	uri = fmt.Sprintf("http://localhost:%d", webCFG.Port)

	go func() {
		err = web.StartServer(webCFG, false)
		if err != nil {
			log.Fatalf("Error initializing web server: %s", err)
		}
	}()

	driver, err := startWebDriver()
	if err != nil {
		log.Fatalf("Error starting web driver: %s", err)
	}

	result := m.Run()

	err = driver.Stop()
	if err != nil {
		log.Fatalf("Failed to close pages and stop WebDriver: %s", err)
	}

	err = data.Teardown()
	if err != nil {
		log.Fatalf("Error tearing down data connections: %s", err)
	}
	os.Exit(result)
}

func startWebDriver() (*agouti.WebDriver, error) {
	browser := os.Getenv("BROWSER")

	// check environment variables to determine the driver
	var driver *agouti.WebDriver
	switch strings.ToLower(browser) {
	case "firefox":
		driver = agouti.NewWebDriver("http://localhost:4444",
			[]string{"geckodriver", "-host", "4444"},
			agouti.Browser("firefox"))
	case "chrome":
	default:
		//TODO: local headless browser?
		return nil, errors.New("Invalid BROWSER value")
	}

	err := driver.Start()
	if err != nil {
		return nil, err
	}

	page, err = driver.NewPage()
	if err != nil {
		return nil, err
	}

	return driver, nil
}
