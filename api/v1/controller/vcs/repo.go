package vcs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopx.io/gopx-common/fs"
	"gopx.io/gopx-vcs-api/api/v1/constants"
	"gopx.io/gopx-vcs-api/api/v1/types"
	"gopx.io/gopx-vcs-api/pkg/config"
)

func repoName(pkgName string) string {
	return fmt.Sprintf("%s%s", pkgName, config.VCS.RepoExt)
}

func vcsRepoAuthorSignature(pkgOwner *types.PackageOwner) *object.Signature {
	return &object.Signature{
		Name:  fmt.Sprintf("%s(%s)", pkgOwner.Name, pkgOwner.Username),
		Email: pkgOwner.PublicEmail,
		When:  time.Now(),
	}
}

func vcsRepoCommitterSignature() *object.Signature {
	return &object.Signature{
		Name:  constants.RepoAutoCommitterName,
		Email: constants.RepoAutoCommitterEmail,
		When:  time.Now(),
	}
}

func vcsRepoTaggerSignature() *object.Signature {
	return &object.Signature{
		Name:  constants.RepoAutoTaggerName,
		Email: constants.RepoAutoTaggerEmail,
		When:  time.Now(),
	}
}

func vcsRepoCommitMessage(tagName string) string {
	return fmt.Sprintf("Update package to version %s", tagName)
}

func vcsRepoTagMessage(tagName string) string {
	return fmt.Sprintf("Released %s", tagName)
}

func vcsRepoCreateTag(repo *git.Repository, tagName string, tagger *object.Signature, message string) (err error) {
	wt, err := repo.Worktree()
	if err != nil {
		err = errors.Wrapf(err, "Failed to access the worktree")
		return
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Master,
	})
	if err != nil {
		err = errors.Wrapf(err, "Failed to checkout master branch")
		return
	}

	headRef, err := repo.Head()
	if err != nil {
		err = errors.Wrapf(err, "Failed to access HEAD reference")
		return
	}

	tag := object.Tag{
		Name:       tagName,
		Tagger:     *tagger,
		Message:    message,
		TargetType: plumbing.CommitObject,
		Target:     headRef.Hash(),
	}

	enObj := repo.Storer.NewEncodedObject()
	tag.Encode(enObj)

	hash, err := repo.Storer.SetEncodedObject(enObj)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't set encoded object for tag")
		return
	}

	tagRefName := fmt.Sprintf("refs/tags/%s", tagName)
	ref := plumbing.NewReferenceFromStrings(tagRefName, hash.String())

	err = repo.Storer.SetReference(ref)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't set reference for tag")
		return
	}

	return
}

func tagNameFromVersion(pkgVersion string) (tagName string, err error) {
	v, err := semver.NewVersion(pkgVersion)
	if err != nil {
		return
	}

	tagName = fmt.Sprintf(
		"v%d.%d.%d",
		v.Major(),
		v.Minor(),
		v.Patch(),
	)

	if v.Prerelease() != "" {
		tagName = fmt.Sprintf("%s-%s", tagName, v.Prerelease())
	}

	return
}

func packageRepoPath(pkgName string) (rPath string, err error) {
	rPath = filepath.Join(config.VCS.RepoRoot, repoName(pkgName))
	rPath, err = filepath.Abs(rPath)
	return
}

func resolvePackageRepo(pkgName string) (err error) {
	rPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	if exists, err := fs.Exists(rPath); err != nil {
		return err
	} else if exists {
		visible, err := isVisibleRepo(rPath)
		if err != nil {
			return err
		}

		if visible {
			_, err = git.PlainOpen(rPath)
			if err == nil || err != git.ErrRepositoryNotExists {
				return err
			}
		}

		err = careCorruptedRepo(rPath)
		if err != nil {
			return err
		}
	}

	_, err = git.PlainInit(rPath, true)
	if err != nil {
		err = errors.Wrapf(err, "Failed to initialize the package repo [%s]", pkgName)
		return
	}

	return
}

func careCorruptedRepo(rPath string) (err error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		err = errors.Wrapf(err, "Failed to generate a UUID [%s]", extractPkgName(rPath))
		return
	}

	repoCorrPath := fmt.Sprintf("%s-%s.corrupted", rPath, uuid.String())

	err = os.Rename(rPath, repoCorrPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to rename repo to corrupted repo [%s]", extractPkgName(rPath))
		return
	}

	err = hideRepo(repoCorrPath)
	if err != nil {
		err = errors.Wrapf(err, "Failed to hide corrupted repo [%s]", extractPkgName(rPath))
		return
	}

	return
}

func makeVisibleRepo(rPath string) (err error) {
	exportFile := filepath.Join(rPath, constants.GitExportRepoFileName)

	exists, err := fs.Exists(exportFile)
	if err != nil || exists {
		return
	}

	file, err := os.Create(exportFile)
	if err != nil {
		err = errors.Wrapf(err, "Unable to create %s file [%s]", constants.GitExportRepoFileName, extractPkgName(rPath))
		return
	}
	file.Close()

	return
}

func hideRepo(rPath string) (err error) {
	exportFile := filepath.Join(rPath, constants.GitExportRepoFileName)
	err = os.RemoveAll(exportFile)
	return
}

func isVisibleRepo(rPath string) (ok bool, err error) {
	exportFile := filepath.Join(rPath, constants.GitExportRepoFileName)
	ok, err = fs.Exists(exportFile)
	return
}

func extractPkgName(rPath string) string {
	repoName := filepath.Base(rPath)
	ext := filepath.Ext(repoName)
	return repoName[:(len(repoName) - len(ext))]
}

func packageRepoCloneOptions(pkgName string) (cloneOpt *git.CloneOptions, err error) {
	rPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	cloneOpt = &git.CloneOptions{
		URL:           rPath,
		RemoteName:    "origin",
		ReferenceName: plumbing.Master,
		SingleBranch:  true,
		Tags:          git.NoTags,
	}

	return
}

func exportPackageRepo(pkgName string) (err error) {
	rPath, err := packageRepoPath(pkgName)
	if err != nil {
		return
	}

	err = makeVisibleRepo(rPath)

	return err
}

func tempPackageRepoOpsDir(pkgName string) (string, error) {
	prefix := fmt.Sprintf("gopx-package-repo-%s-", pkgName)
	return ioutil.TempDir("", prefix)
}
