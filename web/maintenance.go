// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/data"
)

func init() {
	data.MaintenanceTrigger(
		func() { addInterrupt(maintenanceInterrupt) },
		func() { removeInterrupt(maintenanceInterrupt) },
	)

}

var maintenanceTemplate = templateHandler{
	templateFiles: []string{"maintenance.template.html"},
}

var maintenanceInterrupt = &interrupt{
	name: "maintenance",
	fn: func(w http.ResponseWriter, r *http.Request) {
		emptyTemplate(w, r, ctx{})
	},
}
