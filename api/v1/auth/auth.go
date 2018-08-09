package auth

import (
	"strings"

	"github.com/pkg/errors"
	"gopx.io/gopx-common/str"
)

const (
	authTypeAuthKey = "AuthKey"
)

// AuthenticationType represents the http request auth type.
type AuthenticationType interface {
	Name() string
}

// AuthenticationTypeAuthKey represents the Auth Key http auth type.
type AuthenticationTypeAuthKey struct {
	name    string
	authKey string
}

// AuthKey returns the Auth Key value.
func (ata *AuthenticationTypeAuthKey) AuthKey() string {
	return ata.authKey
}

// Name returns auth type name.
func (ata *AuthenticationTypeAuthKey) Name() string {
	return ata.name
}

// AuthenticationTypeUnknown represents an unrecognized http auth type.
type AuthenticationTypeUnknown struct {
	name string
}

// Name returns auth type name.
func (atu *AuthenticationTypeUnknown) Name() string {
	return atu.name
}

// Parse parses the http Authorization header value and returns
// the corresponding auth type.
func Parse(auth string) (authType AuthenticationType, err error) {
	auth = strings.TrimSpace(auth)
	parts := str.SplitSpace(auth)

	if len(parts) < 2 {
		err = errors.New("Invalid auth data")
		return
	}

	switch aType, aVal := parts[0], parts[1]; aType {
	case authTypeAuthKey:
		authType = &AuthenticationTypeAuthKey{
			name:    aType,
			authKey: aVal,
		}
	default:
		authType = &AuthenticationTypeUnknown{
			name: aType,
		}
	}

	return
}
