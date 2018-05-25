// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/lexLibrary/web"
	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
	"github.com/timshannon/sequence"
)

const (
	defaultTimeout = 5 * time.Second
)

var driver selenium.WebDriver
var llURL *url.URL

func newSequence() *sequence.Sequence {
	if os.Getenv("LLDEBUGONERR") == "true" {
		return sequence.Start(driver).OnError(func(err sequence.Error, s *sequence.Sequence) {
			s.Debug().
				Screenshot(fmt.Sprintf("SequenceError-[%s].png", err.Stage))
		})
	}
	return sequence.Start(driver)
}

func TestMain(m *testing.M) {
	err := data.TestingSetup()
	if err != nil {
		log.Fatal(err)
	}

	err = reset()
	if err != nil {
		log.Fatalf("Error resetting database: %s", err)
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
	llURL = &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", hostname, port),
	}

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

	err = firstRun()
	if err != nil {
		log.Fatalf("First run failed: %s", err)
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

func reset() error {
	_, err := data.NewQuery("delete from settings").Exec()
	if err != nil {
		return errors.Wrap(err, "Error emptying settings table before running tests")
	}

	_, err = data.NewQuery("delete from sessions").Exec()
	if err != nil {
		return errors.Wrap(err, "Error emptying sessions table before running tests")
	}

	_, err = data.NewQuery("delete from users").Exec()
	if err != nil {
		return errors.Wrap(err, "Error emptying users table before running tests")
	}
	return nil
}

func createUserAndLogin(username, password string, isAdmin bool) error {
	reset()
	adminUsername := "admin"
	adminPassword := "adminP@ssw0rd"
	if isAdmin {
		adminUsername = username
		adminPassword = password
	}

	user, err := app.FirstRunSetup(adminUsername, adminPassword)
	if err != nil {
		return errors.Wrap(err, "Error setting up admin user")
	}
	admin, err := user.Admin()
	if err != nil {
		return err
	}

	err = admin.SetSetting("AllowPublicSignups", true)
	if err != nil {
		return errors.Wrap(err, "Error allowing public signups for testing")
	}

	err = admin.SetSetting("URL", llURL.String())
	if err != nil {
		return errors.Wrap(err, "Error setting URL for testing")
	}

	err = driver.DeleteAllCookies()
	if err != nil {
		return errors.Wrap(err, "Error clearing all cookies for testing")
	}

	if isAdmin {
		uri := *llURL
		uri.Path = "login"
		err = newSequence().
			Get(uri.String()).
			Find("#inputUsername").SendKeys(username).
			Find("#inputPassword").SendKeys(password).
			Find(".btn.btn-primary.btn-block").Click().
			And().URL().Path("/").Eventually().
			End()
		if err != nil {
			return errors.Wrap(err, "Error signing up user")
		}
		return nil
	}

	uri := *llURL
	uri.Path = "signup"

	return newSequence().
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find("#inputPassword2").SendKeys(password).
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Count(0).
		And().URL().Path("/").Eventually().
		End()
}
