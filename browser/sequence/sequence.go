// Copyright (c) 2017-2018 Townsourced Inc.

package sequence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
)

// Sequence is a helper structs of chaining selecting elements and testing them
// if any part of the sequence fails the sequence ends and returns the error
// built to make writing tests easier
type Sequence struct {
	driver selenium.WebDriver
	err    error
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
	s        *Sequence
	e        []selenium.WebElement
	selector string
}

// NewSequence starts a new sequence of tests
func Start(driver selenium.WebDriver) *Sequence {
	return &Sequence{driver: driver}
}

// End ends a sequence and returns any errors
func (s *Sequence) End() error {
	return s.err
}

// Driver returns the underlying WebDriver
func (s *Sequence) Driver() selenium.WebDriver {
	return s.driver
}

// Test runs an arbitrary test against the entire page
func (s *Sequence) Test(testName string, fn func(d selenium.WebDriver) error) *Sequence {
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

// Title checks the match against the page's title
func (s *Sequence) Title(match string) *Sequence {
	return s.Test("Title", func(d selenium.WebDriver) error {
		title, err := s.driver.Title()
		if err != nil {
			return err
		}
		if title != match {
			return errors.Errorf("Title does not match. Expected %s, got %s", match, title)
		}
		return nil
	})
}

// Get navigates to the passed in URI
func (s *Sequence) Get(uri string) *Sequence {
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

// Forward moves forward in the browser's history
func (s *Sequence) Forward() *Sequence {
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

// Back moves back in the browser's history
func (s *Sequence) Back() *Sequence {
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

// Refresh refreshes the page
func (s *Sequence) Refresh() *Sequence {
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

// Find returns a selection of one or more elements to apply a set of actions against
func (s *Sequence) Find(selector string) *Elements {
	e := &Elements{
		s:        s,
		selector: selector,
	}
	if s.err != nil {
		return e
	}

	e.e, s.err = s.driver.FindElements(selenium.ByCSSSelector, selector)
	if s.err != nil {
		s.err = &Error{
			Stage: "Elements",
			Err:   s.err,
		}
		return e
	}
	return e
}

// And allows you chain additional sequences
func (e *Elements) And() *Sequence {
	return e.s
}

// End Completes a sequence and returns any errors
func (e *Elements) End() error {
	return e.s.End()
}

// Count verifies that the number of elements in the selection matches the argument
func (e *Elements) Count(count int) *Elements {
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

// FindChildren returns a new Elements object for all the elements that match the selector
func (e *Elements) FindChildren(selector string) *Elements {
	newE := &Elements{
		s:        e.s,
		selector: selector,
	}
	if e.s.err != nil {
		return e
	}

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
		newE.e = append(newE.e, elements...)
		success = true
	}

	if !success {
		// all find elements calls failed
		e.s.err = &Error{
			Stage:   "Element.Elements",
			Element: lastElement,
			Err:     lastErr,
		}
		return newE
	}

	return newE
}

// Test tests an arbitrary function against all the elements in this sequence
// if the function returns an error then the test fails
func (e *Elements) Test(testName string, fn func(e selenium.WebElement) error) *Elements {
	if e.s.err != nil {
		return e
	}
	if len(e.e) == 0 {
		e.s.err = &Error{
			Stage: testName,
			Err:   errors.Errorf("No elements exist for the selector '%s'", e.selector),
		}
	}

	for i := range e.e {
		err := fn(e.e[i])
		if err != nil {
			e.s.err = &Error{
				Stage:   testName,
				Element: e.e[i],
				Err:     err,
			}
			return e
		}
	}
	return e
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
			return errors.Errorf("The element's %s does not equal %s. Got %s", s.testName, match, val)
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
			return errors.Errorf("The Element's %s does not contain %s. Got %s", s.testName, match, val)
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
			return errors.Errorf("The Element's %s does not start with %s. Got %s", s.testName, match, val)
		}
		return nil
	})
}

// EndsWith tests if the string value end with the passed in value
func (s *StringMatch) EndsWith(match string) *Elements {
	return s.e.Test(fmt.Sprintf("%s Starts With", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(val, match) {
			return errors.Errorf("The Element's %s does not end with %s. Got %s", s.testName, match, val)
		}
		return nil
	})
}

// Regexp tests if the string value end with the passed in value
func (s *StringMatch) Regexp(exp *regexp.Regexp) *Elements {
	return s.e.Test(fmt.Sprintf("%s Starts With", s.testName), func(we selenium.WebElement) error {
		val, err := s.value(we)
		if err != nil {
			return err
		}
		if !exp.MatchString(val) {
			return errors.Errorf("The Element's %s does not match the regex %s.", s.testName, exp)
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
