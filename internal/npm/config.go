package npm

import "github.com/jacobarthurs/shipbin/internal/config"

type Config struct {
	Name      string
	Version   string
	License   string
	Artifacts []config.Artifact
	DryRun    bool
	Org        string
	Tag        string
	Provenance bool
	Readme     string
}
