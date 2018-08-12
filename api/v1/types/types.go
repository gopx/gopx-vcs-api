package types

// PackageType represents type of package in vcs registry.
type PackageType int

func (p PackageType) String() string {
	switch p {
	case 0:
		return "public"
	case 1:
		return "private"
	default:
		return "unknown"
	}
}

const (
	// PackageTypePublic indicates the package is public.
	PackageTypePublic PackageType = iota
	// PackageTypePrivate indicates the package is private.
	PackageTypePrivate
)

// PackageMeta represents the package meta data required on vcs registry request.
type PackageMeta struct {
	Type    PackageType  `json:"type"`
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Owner   PackageOwner `json:"owner"`
}

// PackageOwner represents the owner data required on vcs registry request.
type PackageOwner struct {
	Name        string `json:"name"`
	PublicEmail string `json:"publicEmail"`
	Username    string `json:"username"`
}
