package pypi

import "github.com/jacobarthurs/shipbin/internal/config"

type Config struct {
	Name      string
	Version   string
	Artifacts []config.Artifact
	Summary   string
	License   string
	Readme    string
	DryRun    bool
}
