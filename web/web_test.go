// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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
var browser string

func newSequence() *sequence.Sequence {
	errScreenPath := os.Getenv("LLERRSCREENPATH")
	if errScreenPath != "" {
		return sequence.Start(driver).OnError(func(err sequence.Error, s *sequence.Sequence) {
			s.Screenshot(filepath.Join(errScreenPath, fmt.Sprintf("SequenceError-[%s].png", err.Stage)))
		})
	}
	return sequence.Start(driver)
}

func TestMain(m *testing.M) {
	err := data.TestingSetup(m)
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
	browser = os.Getenv("LLBROWSER")
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

func setupUserAndLogin(t *testing.T, username, password string, isAdmin bool) {
	data.ResetDB(t)
	adminUsername := "admin"
	adminPassword := "adminP@ssw0rd"
	if isAdmin {
		adminUsername = username
		adminPassword = password
	}

	user, err := app.FirstRunSetup(adminUsername, adminPassword)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Error setting up admin user"))
	}
	admin, err := user.Admin()
	if err != nil {
		t.Fatal(err)
	}

	err = admin.SetSetting("AllowPublicSignups", true)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Error allowing public signups for testing"))
	}

	err = admin.SetSetting("URL", llURL.String())
	if err != nil {
		t.Fatal(errors.Wrap(err, "Error setting URL for testing"))
	}

	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatal(errors.Wrap(err, "Error clearing all cookies for testing"))
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
			t.Fatal(errors.Wrap(err, "Error signing up user"))
		}
		return
	}

	uri := *llURL
	uri.Path = "signup"

	err = newSequence().
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find("#inputPassword2").SendKeys(password).
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Count(0).
		And().URL().Path("/").Eventually().
		End()
	if err != nil {
		t.Fatal(err)
	}
}

func dateInput(date time.Time) string {
	if browser == "chrome" {
		return date.Format("01022006")
	}
	return date.Format("2006-01-02")
}

// assert fails the test if the condition is false.
// func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
// 	if !condition {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
// 		tb.FailNow()
// 	}
// }

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
// func equals(tb testing.TB, exp, act interface{}) {
// 	if !reflect.DeepEqual(exp, act) {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
// 		tb.FailNow()
// 	}
// }
