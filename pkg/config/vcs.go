package config

import (
	"encoding/json"
	"io/ioutil"

	"gopx.io/gopx-common/log"
)

// VCSConfigPath holds vcs related configuration file path.
const VCSConfigPath = "./config/vcs.json"

// VCSConfig represents vcs related configurations.
type VCSConfig struct {
	RepoRoot string `json:"repoRoot"`
	RepoExt  string `json:"repoExt"`
}

// VCS holds loaded VCS related configurations.
var VCS = new(VCSConfig)

func init() {
	bytes, err := ioutil.ReadFile(VCSConfigPath)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
	err = json.Unmarshal(bytes, VCS)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
}
