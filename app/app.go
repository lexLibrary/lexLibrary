// Package app handles the application logic for Lex Library
// All rules and logic that apply to application structures should happen in this library
// Transactions should all be self contained in this library, and not be initiated in the Web layer
// No web structures or packages (http, cookies, etc) should show up in this package
package app

const maxRows = 10000
