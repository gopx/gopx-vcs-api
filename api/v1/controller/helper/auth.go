package helper

import (
	"os"

	"github.com/pkg/errors"
	"gopx.io/gopx-vcs-api/api/v1/auth"
	"gopx.io/gopx-vcs-api/pkg/config"
)

// AuthRequest validates authentication of the incoming http request.
func AuthRequest(authValue string) (ok bool, err error) {
	authType, err := auth.Parse(authValue)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse the auth value")
		return
	}

	switch v := authType.(type) {
	case *auth.AuthenticationTypeAuthKey:
		if v.AuthKey() == os.Getenv(config.Env.GoPxVCSAPIAuthKey) {
			ok = true
		}
	default:
		err = errors.Errorf("Auth type %s is not supported", v.Name())
	}

	return
}
