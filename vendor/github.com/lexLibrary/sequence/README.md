# sequence
Sequence is a frontend for Webdriver testing.  It's main goal are to be easy to use and to make writing frontend tests 
easier in Go.

# Overview
A sequence is a chain of functions that test a web page. If any item in the sequence fails, it drops out early
and returns the error. Compare the test example from the [Go Selenium webdriver](https://raw.githubusercontent.com/tebeka/selenium/master/example_test.go) to Sequence:


```Go
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
    Find("#output").Text().Contains("Hello WebDriver").Eventually().
    End()
if err != nil {
    panic(err)
}
```


The underlying WebDriver library does all the work.  Sequence simply presents a easier API to work with.

See the [documentation](https://godoc.org/github.com/lexLibrary/sequence) for more details.