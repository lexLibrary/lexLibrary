package web

import (
	"fmt"
	"strings"
)

type csp struct {
	defaultSrc []string
	scriptSrc  []string
	styleSrc   []string
	imgSrc     []string
}

var cspDefault = csp{
	defaultSrc: []string{"'self'"},
	scriptSrc:  []string{"'self'", "'unsafe-eval'"},
	styleSrc:   []string{"'self'"},
	imgSrc:     []string{"'self'", "data:"},
}

// func (c csp) addDefault(src ...string) csp {
// 	c.defaultSrc = append(c.defaultSrc, src...)
// 	return c
// }

// func (c csp) addScript(src ...string) csp {
// 	c.scriptSrc = append(c.scriptSrc, src...)
// 	return c
// }

func (c csp) addStyle(src ...string) csp {
	c.styleSrc = append(c.styleSrc, src...)
	return c
}

// func (c csp) addImg(src ...string) csp {
// 	c.imgSrc = append(c.imgSrc, src...)
// 	return c
// }

// String returns the CSP header string based on the definition
func (c csp) String() string {
	if len(c.defaultSrc) == 0 {
		c = cspDefault
	}
	return fmt.Sprintf("default-src %s; script-src %s; style-src %s; img-src %s",
		strings.Join(c.defaultSrc, " "),
		strings.Join(c.scriptSrc, " "),
		strings.Join(c.styleSrc, " "),
		strings.Join(c.imgSrc, " "),
	)
}
