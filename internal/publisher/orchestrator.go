package publisher

import (
	"errors"
	"sync"

	"github.com/jacobarthurs/shipbin/internal/config"
)

func Publish(cfg *config.Config) error {
	var wg sync.WaitGroup
	errs := make([]error, 2)

	if cfg.Target == config.TargetAll || cfg.Target == config.TargetNpm {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: implement npm.Publish
			// errs[0] = npm.Publish(cfg)
		}()
	}

	if cfg.Target == config.TargetAll || cfg.Target == config.TargetPyPI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: implement pypi.Publish
			// errs[1] = pypi.Publish(cfg)
		}()
	}

	wg.Wait()
	return errors.Join(errs...)
}
