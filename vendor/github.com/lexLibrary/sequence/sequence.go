// Copyright (c) 2017-2018 Townsourced Inc.

package sequence

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
)

// Sequence is a helper structs of chaining selecting elements and testing them
// if any part of the sequence fails the sequence ends and returns the error
// built to make writing tests easier
type Sequence struct {
	driver          selenium.WebDriver
	err             error
	EventualPoll    time.Duration
	EventualTimeout time.Duration
	last            func() *Sequence
}

// Error describes an error that occured during the sequence processing.
type Error struct {
	Stage   string
	Element selenium.WebElement
	Err     error
}

// Error fulfills the error interface
func (e *Error) Error() string {
	if e.Element != nil {
		return fmt.Sprintf("An error occurred during %s on element %s: %s", e.Stage, elementString(e.Element),
			e.Err)
	}
	return fmt.Sprintf("An error occurred during %s:  %s", e.Stage, e.Err)
}

// Errors is multiple sequence errors
type Errors []error

func (e Errors) Error() string {
	str := "Multiple errors occurred: \n"
	for i := range e {
		str += "\t" + e[i].Error() + "\n"
	}
	return str
}

func elementString(element selenium.WebElement) string {
	if element == nil {
		return ""
	}
	id, err := element.GetAttribute("id")
	if err == nil && id != "" {
		return fmt.Sprintf("#%s", id)
	}
	tag, err := element.TagName()
	if err != nil {
		return fmt.Sprintf("%v", element)
	}
	text, err := element.Text()
	if err != nil {
		return fmt.Sprintf("%v", element)
	}

	if len(text) > 25 {
		text = text[:25]
	}

	return fmt.Sprintf("<%s>%s</%s>", tag, text, tag)
}

// Elements is a collections of web elements
type Elements struct {
	s          *Sequence
	e          []selenium.WebElement
	selector   string
	selectFunc func(selector string) ([]selenium.WebElement, error)
	last       func() *Elements
	all        bool
	any        bool
}

// NewSequence starts a new sequence of tests
func Start(driver selenium.WebDriver) *Sequence {
	return &Sequence{
		driver:          driver,
		EventualPoll:    100 * time.Millisecond,
		EventualTimeout: 60 * time.Second,
	}
}

// End ends a sequence and returns any errors
func (s *Sequence) End() error {
	return s.err
}

// Driver returns the underlying WebDriver
func (s *Sequence) Driver() selenium.WebDriver {
	return s.driver
}

// Eventually will retry the previous test if it returns an error every EventuallyPoll duration until EventualTimeout
// is reached
func (s *Sequence) Eventually() *Sequence {
	if s.err == nil {
		return s
	}

	err := s.driver.WaitWithTimeoutAndInterval(func(d selenium.WebDriver) (bool, error) {
		s.err = nil
		s = s.last()
		if s.err != nil {
			return false, nil
		}
		return true, nil
	}, s.EventualTimeout, s.EventualPoll)
	if err != nil {
		s.err = errors.Wrap(s.err, "Eventually timed out")
	}
	return s
}

// Eventually will retry the previous test if it returns an error every EventuallyPoll duration until EventualTimeout
// is reached
func (e *Elements) Eventually() *Elements {
	if e.s.err == nil {
		return e
	}

	err := e.s.driver.WaitWithTimeoutAndInterval(func(d selenium.WebDriver) (bool, error) {
		e.s.err = nil
		e = e.last()
		if e.s.err != nil {
			return false, nil
		}
		return true, nil
	}, e.s.EventualTimeout, e.s.EventualPoll)
	if err != nil {
		e.s.err = errors.Wrap(e.s.err, "Eventually timed out")
	}
	return e
}

// Test runs an arbitrary test against the entire page
func (s *Sequence) Test(testName string, fn func(d selenium.WebDriver) error) *Sequence {
	s.last = func() *Sequence {
		if s.err != nil {
			return s
		}

		err := fn(s.driver)

		if err != nil {
			s.err = &Error{
				Stage: testName,
				Err:   err,
			}
		}
		return s
	}
	return s.last()
}

// TitleMatch is for testing the value of the title
type TitleMatch struct {
	title string
	s     *Sequence
}

func (t *TitleMatch) test(testName string, fn func() error) *Sequence {
	t.s.last = func() *Sequence {
		if t.s.err != nil {
			return t.s
		}
		title, err := t.s.driver.Title()
		if err != nil {
			t.s.err = &Error{
				Stage: "Title " + testName,
				Err:   err,
			}
			return t.s
		}
		t.title = title
		err = fn()
		if err != nil {
			t.s.err = &Error{
				Stage: "Title " + testName,
				Err:   err,
			}
		}
		return t.s
	}
	return t.s.last()
}

// Equals tests if the title matches the passed in value exactly
func (t *TitleMatch) Equals(match string) *Sequence {
	return t.test("Equals", func() error {
		if t.title != match {
			return errors.Errorf("The page's title does not equal '%s'. Got '%s'", match, t.title)
		}
		return nil
	})
}

// Contains tests if the title contains the passed in value
func (t *TitleMatch) Contains(match string) *Sequence {
	return t.test("Contains", func() error {
		if !strings.Contains(t.title, match) {
			return errors.Errorf("The pages's title does not contain '%s'. Got '%s'", match, t.title)
		}
		return nil
	})
}

// StartsWith tests if the title starts with the passed in value
func (t *TitleMatch) StartsWith(match string) *Sequence {
	return t.test("Starts With", func() error {
		if !strings.HasPrefix(t.title, match) {
			return errors.Errorf("The pages's title does not start with '%s'. Got '%s'", match, t.title)
		}
		return nil
	})
}

// EndsWith tests if the title ends with the passed in value
func (t *TitleMatch) EndsWith(match string) *Sequence {
	return t.test("Ends With", func() error {
		if !strings.HasSuffix(t.title, match) {
			return errors.Errorf("The pages's title does not end with '%s'. Got '%s'", match, t.title)
		}
		return nil
	})
}

// Regexp tests if the title matches the regular expression
func (t *TitleMatch) Regexp(exp *regexp.Regexp) *Sequence {
	return t.test("Matches RegExp", func() error {
		if !exp.MatchString(t.title) {
			return errors.Errorf("The pages's title does not match the regular expression '%s'. Title: '%s'",
				exp, t.title)
		}
		return nil
	})
}

// Title checks the match against the page's title
func (s *Sequence) Title() *TitleMatch {
	return &TitleMatch{
		s: s,
	}
}

// Get navigates to the passed in URI
func (s *Sequence) Get(uri string) *Sequence {
	s.last = func() *Sequence {
		if s.err != nil {
			return s
		}

		err := s.driver.Get(uri)
		if err != nil {
			s.err = &Error{
				Stage: "Get",
				Err:   err,
			}
		}
		return s
	}
	return s.last()
}

// URLMatch is for testing the value of the page's URL
type URLMatch struct {
	url *url.URL
	s   *Sequence
}

func (u *URLMatch) test(testName string, fn func() error) *Sequence {
	u.s.last = func() *Sequence {
		if u.s.err != nil {
			return u.s
		}
		uri, err := u.s.driver.CurrentURL()
		if err != nil {
			u.s.err = &Error{
				Stage: "URL " + testName,
				Err:   err,
			}
			return u.s
		}

		u.url, err = url.Parse(uri)
		if err != nil {
			u.s.err = &Error{
				Stage: "URL " + testName,
				Err:   err,
			}
			return u.s
		}
		err = fn()
		if err != nil {
			u.s.err = &Error{
				Stage: "URL " + testName,
				Err:   err,
			}
		}
		return u.s
	}
	return u.s.last()
}

// Path tests if the page's url path matches the passed in value
func (u *URLMatch) Path(match string) *Sequence {
	return u.test("Path Matches", func() error {
		if u.url.Path != match {
			return errors.Errorf("URL's path does not match %s, got %s", match, u.url.Path)
		}
		return nil
	})
}

// QueryValue tests if the page's url contains the url query matches the value
func (u *URLMatch) QueryValue(key, value string) *Sequence {
	return u.test("Query Value Matches", func() error {
		values := u.url.Query()
		if v, ok := values[key]; ok {
			found := false
			for i := range v {
				if v[i] == value {
					found = true
					break
				}

			}
			if !found {
				return errors.Errorf("URL does not contain the value '%s' for the key '%s'. Values: %s",
					value, key, v)
			}
			return nil
		}

		return errors.Errorf("URL does not contain the query key '%s'. URL: %s", key, u.url)
	})
}

// Fragment tests if the page's url fragment (#) matches the passed in value
func (u *URLMatch) Fragment(match string) *Sequence {
	return u.test("Fragment Matches", func() error {
		if u.url.Fragment != match {
			return errors.Errorf("URL's fragment does not match %s, got %s", match, u.url.Fragment)
		}
		return nil
	})
}

func (s *Sequence) URL() *URLMatch {
	return &URLMatch{
		s: s,
	}
}

// Forward moves forward in the browser's history
func (s *Sequence) Forward() *Sequence {
	s.last = func() *Sequence {
		if s.err != nil {
			return s
		}

		err := s.driver.Forward()
		if err != nil {
			s.err = &Error{
				Stage: "Forward",
				Err:   err,
			}
		}
		return s
	}
	return s.last()
}

// Back moves back in the browser's history
func (s *Sequence) Back() *Sequence {
	s.last = func() *Sequence {
		if s.err != nil {
			return s
		}

		err := s.driver.Back()
		if err != nil {
			s.err = &Error{
				Stage: "Back",
				Err:   err,
			}
		}
		return s
	}
	return s.last()
}

// Refresh refreshes the page
func (s *Sequence) Refresh() *Sequence {
	s.last = func() *Sequence {
		if s.err != nil {
			return s
		}

		err := s.driver.Refresh()
		if err != nil {
			s.err = &Error{
				Stage: "Refresh",
				Err:   err,
			}
		}
		return s
	}
	return s.last()
}

// Find returns a selection of one or more elements to apply a set of actions against
// If .Any or.All are not specified, then it is assumed that the selection will contain a single element
// and the tests will fail if more than one element is found
func (s *Sequence) Find(selector string) *Elements {
	e := &Elements{
		s:        s,
		selector: selector,
		selectFunc: func(selector string) ([]selenium.WebElement, error) {
			return s.driver.FindElements(selenium.ByCSSSelector, selector)
		},
	}
	if s.err != nil {
		return e
	}
	e.e, s.err = e.selectFunc(selector)

	if s.err != nil {
		s.err = &Error{
			Stage: "Elements",
			Err:   s.err,
		}
		return e
	}
	return e
}

// Wait will wait for the given duration before continuing in the sequence
func (s *Sequence) Wait(duration time.Duration) *Sequence {
	if s.err != nil {
		return s
	}
	time.Sleep(duration)
	return s
}

// Debug will print the current page's title and source
// For use with debugging issues mostly
func (s *Sequence) Debug() *Sequence {
	if s.err != nil {
		return s
	}
	src, err := s.driver.PageSource()
	if err != nil {
		s.err = &Error{
			Stage: "Debug Source",
			Err:   err,
		}
		return s
	}

	title, err := s.driver.Title()
	if err != nil {
		s.err = &Error{
			Stage: "Debug Title",
			Err:   err,
		}
		return s
	}

	uri, err := s.driver.CurrentURL()
	if err != nil {
		s.err = &Error{
			Stage: "Debug URL",
			Err:   err,
		}
		return s
	}

	fmt.Println("-----------------------------------------------")
	fmt.Printf("%s - (%s)\n", title, uri)
	fmt.Println("-----------------------------------------------")
	fmt.Println(src)
	return s
}

// Screenshot takes a screenshot
func (s *Sequence) Screenshot(filename string) *Sequence {
	if s.err != nil {
		return s
	}

	buff, err := s.driver.Screenshot()
	if err != nil {
		s.err = &Error{
			Stage: "Screenshot",
			Err:   err,
		}
		return s
	}

	err = ioutil.WriteFile(filename, buff, 0622)
	if err != nil {
		s.err = &Error{
			Stage: "Screenshot Writing File",
			Err:   err,
		}
		return s
	}
	return s
}

// End Completes a sequence and returns any errors
func (e *Elements) End() error {
	return e.s.End()
}

func (e *Elements) Wait(duration time.Duration) *Elements {
	time.Sleep(duration)
	return e
}

// Any means the following tests will pass if they pass for ANY of the selected elements
func (e *Elements) Any() *Elements {
	e.all = false
	e.any = true
	return e
}

// All means the following tests will pass if they pass only if pass for ALL of the selected elements
func (e *Elements) All() *Elements {
	e.any = false
	e.all = true
	return e
}

// Count verifies that the number of elements in the selection matches the argument
func (e *Elements) Count(count int) *Elements {
	e.last = func() *Elements {
		if e.s.err != nil {
			return e
		}

		if count != len(e.e) {
			e.s.err = &Error{
				Stage: "Count",
				Err: errors.Errorf("Invalid count for selector %s wanted %d got %d", e.selector, count,
					len(e.e)),
			}

			return e
		}
		return e
	}
	return e.last()
}

// And allows you chain additional sequences
func (e *Elements) And() *Sequence {
	return e.s
}

// Find finds a new element
func (e *Elements) Find(selector string) *Elements {
	return e.s.Find(selector)
}

// FindChildren returns a new Elements object for all the elements that match the selector
func (e *Elements) FindChildren(selector string) *Elements {
	newE := &Elements{
		s:        e.s,
		selector: selector,
		selectFunc: func(selector string) ([]selenium.WebElement, error) {
			var found []selenium.WebElement
			success := false
			var lastErr error
			var lastElement selenium.WebElement

			for i := range e.e {
				elements, err := e.e[i].FindElements(selenium.ByCSSSelector, selector)
				if err != nil {
					lastElement = e.e[i]
					lastErr = err
					continue
				}
				found = append(found, elements...)
				success = true
			}
			if !success {
				// all find elements calls failed
				return nil, &Error{
					Stage:   "Find Children",
					Element: lastElement,
					Err:     lastErr,
				}
			}
			return found, nil
		},
	}
	if e.s.err != nil {
		return e
	}

	newE.e, newE.s.err = newE.selectFunc(selector)

	return newE
}

// Test tests an arbitrary function against all the elements in this sequence
// if the function returns an error then the test fails
func (e *Elements) Test(testName string, fn func(e selenium.WebElement) error) *Elements {
	stage := testName + " Test"
	e.last = func() *Elements {
		if e.s.err != nil {
			return e
		}

		if len(e.e) == 0 {
			e.s.err = &Error{
				Stage: stage,
				Err:   errors.Errorf("No elements exist for the selector '%s'", e.selector),
			}
		}
		if len(e.e) == 1 {
			err := fn(e.e[0])
			if err != nil {
				e.s.err = &Error{
					Stage:   stage,
					Element: e.e[0],
					Err:     err,
				}
			}
			return e
		}

		if !e.any && !e.all {
			e.s.err = &Error{
				Stage: stage,
				Err: errors.Errorf("Selector '%s' returned multiple elements but .Any() or .All() weren't specified",
					e.selector),
			}
			return e
		}

		var errs Errors

		for i := range e.e {
			err := fn(e.e[i])
			if err != nil {
				if e.all {
					e.s.err = &Error{
						Stage:   stage,
						Element: e.e[i],
						Err:     errors.Wrap(err, "Not All elements passed"),
					}
					return e
				}
				errs = append(errs, &Error{
					Stage:   stage,
					Element: e.e[i],
					Err:     err,
				})
			} else if e.any {
				return e
			}
		}
		if len(errs) != 0 {
			e.s.err = errors.Wrap(errs, "None of the elements passed")
		}
		return e
	}
	return e.last()
}

// Visible tests if the elements are visible
func (e *Elements) Visible() *Elements {
	return e.Test("Visible", func(we selenium.WebElement) error {
		ok, err := we.IsDisplayed()
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("Element was not visible")
		}
		return nil
	})
}

// Hidden tests if the elements are hidden
func (e *Elements) Hidden() *Elements {
	return e.Test("Hidden", func(we selenium.WebElement) error {
		ok, err := we.IsDisplayed()
		if err != nil {
			return err
		}
		if ok {
			return errors.New("Element was not visible")
		}
		return nil
	})
}

// Enabled tests if the elements are hidden
func (e *Elements) Enabled() *Elements {
	return e.Test("Enabled", func(we selenium.WebElement) error {
		ok, err := we.IsEnabled()
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("Element was not enabled")
		}
		return nil
	})
}

// Disabled tests if the elements are hidden
func (e *Elements) Disabled() *Elements {
	return e.Test("Disabled", func(we selenium.WebElement) error {
		ok, err := we.IsEnabled()
		if err != nil {
			return err
		}
		if ok {
			return errors.New("Element was not disabled")
		}
		return nil
	})
}

// Selected tests if the elements are selected
func (e *Elements) Selected() *Elements {
	return e.Test("Selected", func(we selenium.WebElement) error {
		ok, err := we.IsSelected()
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("Element was not selected")
		}
		return nil
	})
}

// Unselected tests if the elements aren't selected
func (e *Elements) Unselected() *Elements {
	return e.Test("Selected", func(we selenium.WebElement) error {
		ok, err := we.IsSelected()
		if err != nil {
			return err
		}
		if ok {
			return errors.New("Element was selected")
		}
		return nil
	})
}

// StringMatch is for testing the value of strings in elements
type StringMatch struct {
	testName string
	value    func(selenium.WebElement) (string, error)
	e        *Elements
}

// Equals tests if the string value matches the passed in value exactly
func (s *StringMatch) Equals(match string) *Elements {
	return s.e.Test(fmt.Sprintf("%s Equals", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if val != match {
			return errors.Errorf("The element's %s does not equal '%s'. Got '%s'", s.testName, match, val)
		}
		return nil
	})
}

// Contains tests if the string value contains the passed in value
func (s *StringMatch) Contains(match string) *Elements {
	return s.e.Test(fmt.Sprintf("%s Contains", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !strings.Contains(val, match) {
			return errors.Errorf("The Element's %s does not contain '%s'. Got '%s'", s.testName, match, val)
		}
		return nil
	})
}

// StartsWith tests if the string value starts with the passed in value
func (s *StringMatch) StartsWith(match string) *Elements {
	return s.e.Test(fmt.Sprintf("%s Starts With", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(val, match) {
			return errors.Errorf("The Element's %s does not start with '%s'. Got '%s'", s.testName, match, val)
		}
		return nil
	})
}

// EndsWith tests if the string value end with the passed in value
func (s *StringMatch) EndsWith(match string) *Elements {
	return s.e.Test(fmt.Sprintf("%s Ends With", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(val, match) {
			return errors.Errorf("The Element's %s does not end with '%s'. Got '%s'", s.testName, match, val)
		}
		return nil
	})
}

// Regexp tests if the string value matches the regular expression
func (s *StringMatch) Regexp(exp *regexp.Regexp) *Elements {
	return s.e.Test(fmt.Sprintf("%s Matches RegExp", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !exp.MatchString(val) {
			return errors.Errorf("The Element's %s does not match the regex '%s'.", s.testName, exp)
		}
		return nil
	})
}

// TagName tests if the elements match the given tag name
func (e *Elements) TagName() *StringMatch {
	return &StringMatch{
		testName: "TagName",
		value: func(we selenium.WebElement) (string, error) {
			return we.TagName()
		},
		e: e,
	}
}

// Text tests if the elements matches
func (e *Elements) Text() *StringMatch {
	return &StringMatch{
		testName: "Text",
		value: func(we selenium.WebElement) (string, error) {
			return we.Text()
		},
		e: e,
	}
}

// Attribute tests if the elements attribute matches
func (e *Elements) Attribute(attribute string) *StringMatch {
	return &StringMatch{
		testName: fmt.Sprintf("%s Attribute", attribute),
		value: func(we selenium.WebElement) (string, error) {
			return we.GetAttribute(attribute)
		},
		e: e,
	}
}

// CSSProperty tests if the elements attribute matches
func (e *Elements) CSSProperty(property string) *StringMatch {
	return &StringMatch{
		testName: fmt.Sprintf("%s CSS Property", property),
		value: func(we selenium.WebElement) (string, error) {
			return we.CSSProperty(property)
		},
		e: e,
	}
}

// Click sends a click to all of the elements
func (e *Elements) Click() *Elements {
	return e.Test("Click", func(we selenium.WebElement) error {
		return we.Click()
	})
}

// SendKeys sends a string of key to the elements
func (e *Elements) SendKeys(keys string) *Elements {
	return e.Test("SendKeys", func(we selenium.WebElement) error {
		return we.SendKeys(keys)
	})
}

// Submit sends a submit event to the elements
func (e *Elements) Submit() *Elements {
	return e.Test("Submit", func(we selenium.WebElement) error {
		return we.Submit()
	})
}

// Clear clears the elements
func (e *Elements) Clear() *Elements {
	return e.Test("Clear", func(we selenium.WebElement) error {
		return we.Clear()
	})
}
