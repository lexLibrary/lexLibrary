// +build dev

package files

import "net/http"

func init() {
	FileSystem = http.Dir("./assets/")
}
