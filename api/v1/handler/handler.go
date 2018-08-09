package handler

import (
	"io"
	"net/http"
)

// PackagesPOST registers a new package or a new version of an
// existing package to the vcs registry.
// Request: POST /packages
func PackagesPOST(w http.ResponseWriter, r *http.Request) {
	ur, err := authUser(r.Header.Get("Authorization"))
}

// SinglePackageReadmeGET returns the content of README file from vcs registry.
// Request: GET /packages/:packageName/readme
// For a specific version: GET /packages/:packageName/readme?v=1.0.2
func SinglePackageReadmeGET(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "Hey\n")
}

// SinglePackageDELETE deletes a whole package from vcs registry.
// Request: DELETE /packages/:packageName
func SinglePackageDELETE(w http.ResponseWriter, r *http.Request) {
}
