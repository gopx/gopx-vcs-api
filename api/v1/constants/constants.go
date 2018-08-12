package constants

import (
	"github.com/pkg/errors"
)

// Error constants
var (
	ErrInternalServer = errors.New("Internal server error occurred")
)

// MultiPartReaderMaxMemorySize is the maximum memory size used while reading
// multipart/form-data content.
const MultiPartReaderMaxMemorySize = 10 * 1024 * 1024

const (
	// RepoAutoCommitterName represents the commiter name for auto generated
	// commits in package repo.
	RepoAutoCommitterName = "GoPx"

	// RepoAutoCommitterEmail represents the commiter email for auto generated
	// commits in package repo.
	RepoAutoCommitterEmail = "gopx@gopx.io"

	// RepoAutoTaggerName represents the tagger name for auto generated
	// tags in package repo.
	RepoAutoTaggerName = "GoPx"

	// RepoAutoTaggerEmail represents the tagger email for auto generated
	// tags in package repo.
	RepoAutoTaggerEmail = "gopx@gopx.io"
)

// GitExportRepoFileName is the file name which existence is responsible
// for package exporting status.
const GitExportRepoFileName = "git-daemon-export-ok"
