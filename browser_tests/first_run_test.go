package browser_test

import (
	"testing"
)

func TestFirstRun(t *testing.T) {
	if err := page.Navigate(uri); err != nil {
		t.Fatal("Failed to navigate:", err)
	}
}
