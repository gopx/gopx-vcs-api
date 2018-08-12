package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gopx.io/gopx-common/log"
	"gopx.io/gopx-vcs-api/api/v1/constants"
	"gopx.io/gopx-vcs-api/api/v1/controller/helper"
	"gopx.io/gopx-vcs-api/api/v1/controller/vcs"
	"gopx.io/gopx-vcs-api/api/v1/types"
	errorCtrl "gopx.io/gopx-vcs-api/pkg/controller/error"
)

// PackagesPOST registers a new package or a new version of an
// existing package to the vcs registry.
// Request: POST /packages
func PackagesPOST(w http.ResponseWriter, r *http.Request) {
	ok, err := helper.AuthRequest(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if !ok {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	err = r.ParseMultipartForm(constants.MultiPartReaderMaxMemorySize)
	if err != nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Content-Type must be multipart/form-data")
		return
	}

	mForm := r.MultipartForm
	mfMeta := mForm.Value["meta"]
	mfData := mForm.File["data"]

	if mfMeta == nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Package meta not found with param name meta")
		return
	}

	if mfData == nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Package data not found with param name data as a file")
		return
	}

	pkgMeta := mfMeta[0]
	pkgData, err := mfData[0].Open()
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	meta := types.PackageMeta{}
	err = json.NewDecoder(strings.NewReader(pkgMeta)).Decode(&meta)
	if err != nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Problems parsing JSON meta data")
		return
	}

	switch meta.Type {
	case types.PackageTypePublic:
		err = vcs.RegisterPublicPackage(&meta, pkgData)
	case types.PackageTypePrivate:
		err = vcs.RegisterPrivatePackage(&meta, pkgData)
	default:
		errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Unknown package type %d", int(meta.Type)))
		return
	}
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	helper.WriteResponse(w, r, []byte{}, http.StatusCreated)
}

// SinglePackageDELETE deletes a whole package from vcs registry.
// Request: DELETE /packages/:packageName
func SinglePackageDELETE(w http.ResponseWriter, r *http.Request) {
	inputPkgName := mux.Vars(r)["packageName"]

	ok, err := vcs.PackageExists(inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if !ok {
		errorCtrl.Error404(w, r)
		return
	}

	err = vcs.DeletePackage(inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	helper.WriteResponse(w, r, nil, http.StatusNoContent)
}
