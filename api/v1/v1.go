package v1

import (
	"github.com/gorilla/mux"
	"gopx.io/gopx-vcs-api/api/v1/handler"
)

// RegisterRoutes registers the routes for API version v1.
func RegisterRoutes(r *mux.Router) {
	r.Path("/packages").
		Methods("POST").
		HandlerFunc(handler.PackagesPOST)

	r.Path("/packages/{packageName}").
		Methods("DELETE").
		HandlerFunc(handler.SinglePackageDELETE)
}
