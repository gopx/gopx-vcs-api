package vcs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopx.io/gopx-common/fs"
	"gopx.io/gopx-vcs-api/api/v1/controller/helper"
	"gopx.io/gopx-vcs-api/api/v1/types"
)

// RegisterPublicPackage registers a package to the vcs registry as public.
// Note: Assume the package meta and owner info provided in
// http request (from gopx-api service) is valid,
// so doesn't need to sanitize it again.
func RegisterPublicPackage(meta *types.PackageMeta, data io.Reader) (err error) {
	pkgName := meta.Name
	pkgVersion := meta.Version
	owner := &meta.Owner

	err = resolvePackageRepo(pkgName)
	if err != nil {
		err = errors.Wrapf(err, "Failed to resolve the package [%s]", pkgName)
		return
	}

	verExists, err := packageVersionExists(pkgName, pkgVersion)
	if err != nil {
		err = errors.Wrapf(err, "Failed to check version exists or not [%s]", pkgName)
		return
	}

	if verExists {
		err = errors.Errorf("Package version already exists %s [%s]", meta.Version, pkgName)
		return
	}

	opsDir, err := tempPackageRepoOpsDir(pkgName)
	if err != nil {
		err = errors.Wrapf(err, "Unable to create new temp dir for repo operations [%s]", pkgName)
		return
	}
	defer os.RemoveAll(opsDir)

	cloneOpt, err := packageRepoCloneOptions(pkgName)
	if err != nil {
		err = errors.Wrapf(err, "Failed to create package repo clone options [%s]", pkgName)
		return
	}

	repo, err := git.PlainClone(opsDir, false, cloneOpt)
	if err != nil && err != transport.ErrEmptyRemoteRepository {
		err = errors.Wrapf(err, "Failed to clone the package repo [%s]", pkgName)
		return
	}

	files, err := ioutil.ReadDir(opsDir)
	if err != nil {
		err = errors.Wrapf(err, "Unable to read package temp operaion dir [%s]", pkgName)
		return
	}

	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}
		err = os.RemoveAll(filepath.Join(opsDir, f.Name()))
		if err != nil {
			err = errors.Wrapf(err, "Unable to delete package files from temp operation dir [%s]", pkgName)
			return
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		err = errors.Wrapf(err, "Unable to access repo worktree [%s]", pkgName)
		return
	}

	_, err = fs.DecompressTarGZ(opsDir, data)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't extract package data into worktree [%s]", pkgName)
		return
	}

	files, err = ioutil.ReadDir(opsDir)
	if err != nil {
		err = errors.Wrapf(err, "Unable to read package temp operaion dir [%s]", pkgName)
		return
	}

	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}
		_, err = wt.Add(f.Name())
		if err != nil {
			err = errors.Wrapf(err, "Unable to stage files into git index [%s]", pkgName)
			return
		}
	}

	pkgTagName, err := tagNameFromVersion(pkgVersion)
	if err != nil {
		err = errors.Wrapf(err, "Failed to generate tagname [%s]", pkgName)
		return
	}

	_, err = wt.Commit(
		vcsRepoCommitMessage(pkgTagName),
		&git.CommitOptions{
			All:       true,
			Author:    vcsRepoAuthorSignature(owner),
			Committer: vcsRepoCommitterSignature(),
		},
	)
	if err != nil {
		err = errors.Wrapf(err, "Failed to commit package updated data [%s]", pkgName)
		return
	}

	err = vcsRepoCreateTag(
		repo,
		pkgTagName,
		vcsRepoTaggerSignature(),
		vcsRepoTagMessage(pkgTagName),
	)
	if err != nil {
		err = errors.Wrapf(err, "Unable to create tag [%s]", pkgName)
		return
	}

	tagRefSpec := config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", pkgTagName, pkgTagName))
	masterRefSpec := config.RefSpec("+refs/heads/master:refs/heads/master")
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{tagRefSpec, masterRefSpec},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		err = errors.Wrapf(err, "Unable to push package data to the repo [%s]", pkgName)
		return
	}

	err = exportPackageRepo(pkgName)
	if err != nil {
		err = errors.Wrapf(err, "Unable to export package repo [%s]", pkgName)
		return
	}

	return nil
}

// RegisterPrivatePackage registers a package to the vcs registry as private.
// Note: Assume the package meta and owner info provided in
// http request (from gopx-api service) is valid,
// so doesn't need to sanitize it again.
func RegisterPrivatePackage(meta *types.PackageMeta, data io.Reader) (err error) {
	return errors.New("Private package is not supported")
}

func packageVersionExists(pkgName, version string) (ok bool, err error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return
	}

	rPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	repo, err := git.PlainOpen(rPath)
	if err != nil {
		err = errors.Wrapf(err, "Package repo couldn't be opened [%s]", pkgName)
		return
	}

	tagIter, err := repo.TagObjects()
	if err != nil {
		err = errors.Wrapf(err, "Couldn't access tag Objects [%s]", pkgName)
		return
	}

	for {
		tag, err := tagIter.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				err = errors.Wrapf(err, "Couldn't iterate over tag Objects [%s]", pkgName)
				return false, err
			}
		}

		tagVer, err := semver.NewVersion(tag.Name)
		if err != nil {
			return false, err
		}

		if v.Equal(tagVer) {
			return true, nil
		}

	}

	return false, nil
}

// PackageExists checks whether the package exists or not.
func PackageExists(pkgName string) (ok bool, err error) {
	repoPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	pathExists, err := fs.Exists(repoPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to check package path existence [%s]", pkgName)
		return
	}

	if !pathExists {
		ok = false
		return
	}

	visible, err := isVisibleRepo(repoPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to check whether the package is visible or not [%s]", pkgName)
		return
	}

	if !visible {
		ok = false
		return
	}

	ok = true

	return
}

// DeletePackage removes package data from vcs storage.
func DeletePackage(pkgName string) (err error) {
	repoPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	repoDelPath := fmt.Sprintf("%s-%s.deleted", repoPath, helper.FormatTime(time.Now()))

	err = os.Rename(repoPath, repoDelPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to rename repo to deleted repo [%s]", pkgName)
		return
	}

	err = hideRepo(repoDelPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to hide deleted repo [%s]", pkgName)
		return
	}

	return
}
