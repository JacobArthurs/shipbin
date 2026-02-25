package publisher

import (
	"errors"
	"slices"
	"sync"

	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/npm"
)

func Publish(cfg *config.Config) error {
	var wg sync.WaitGroup
	errs := make([]error, 2)

	if slices.Contains(cfg.Targets, config.TargetNpm) {
		wg.Go(func() {
			errs[0] = npm.Publish(cfg)
		})
	}

	if slices.Contains(cfg.Targets, config.TargetPyPI) {
		wg.Go(func() {
			// TODO: implement pypi.Publish
			// errs[1] = pypi.Publish(cfg)
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}
