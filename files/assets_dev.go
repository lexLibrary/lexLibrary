// +build dev

package files

import "net/http"

var Assets = http.Dir("./assets/")
