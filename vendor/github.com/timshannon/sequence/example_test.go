package sequence_test

import (
	"fmt"
	"os"

	"github.com/lexLibrary/sequence"
	"github.com/tebeka/selenium"
)

// This is a Sequence of the same example from https://github.com/tebeka/selenium/blob/master/example_test.go
// This example shows how to navigate to a http://play.golang.org page, input a
// short program, run it, and inspect its output.
func Example() {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const (
		// These paths will be different on your system.
		seleniumPath    = "vendor/selenium-server-standalone-3.4.jar"
		geckoDriverPath = "vendor/geckodriver-v0.18.0-linux64"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// once you have a webdriver initiated, then you simply pass it into a new sequence

	err = sequence.Start(wd).
		Get("http://play.golang.org/?simple=1").
		Find("#code").Clear().SendKeys(`
			package main
			import "fmt"

			func main() {
				fmt.Println("Hello WebDriver!\n")
			}
		`).
		Find("#run").Click().
		Find("#output").
		Text().Contains("Hello WebDriver").Eventually().
		End()
	if err != nil {
		panic(err)
	}
}
