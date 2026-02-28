package npm

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	registryPollInterval = 5 * time.Second
	registryPollTimeout  = 2 * time.Minute
)

func Publish(cfg *Config) error {
	verb := "publishing"
	if cfg.DryRun {
		verb = "would publish"
	}

	fmt.Printf("npm: %s version %s\n", verb, cfg.Version)

	platforms, cleanup, err := buildPlatformPackages(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	for _, pkg := range platforms {
		fmt.Printf("npm: %s %s...\n", verb, pkg.name)
		if err := npmPublish(pkg.dir, cfg.Tag, cfg.Provenance, cfg.DryRun); err != nil {
			return fmt.Errorf("npm: failed to publish %s: %w", pkg.name, err)
		}
	}

	if !cfg.DryRun {
		fmt.Println("npm: waiting for registry propagation...")
		for _, pkg := range platforms {
			if err := pollUntilVisible(pkg.name, cfg.Version); err != nil {
				return fmt.Errorf("npm: registry propagation timed out for %s: %w", pkg.name, err)
			}
		}
	}

	root, rootCleanup, err := buildRootPackage(cfg)
	if err != nil {
		return err
	}
	defer rootCleanup()

	fmt.Printf("npm: %s %s...\n", verb, root.name)
	if err := npmPublish(root.dir, cfg.Tag, cfg.Provenance, cfg.DryRun); err != nil {
		return fmt.Errorf("npm: failed to publish root package %s: %w", root.name, err)
	}

	fmt.Println("npm: done")
	return nil
}

func npmPublish(dir, tag string, provenance, dryRun bool) error {
	args := []string{"publish", "--access", "public", "--tag", tag}
	if provenance {
		args = append(args, "--provenance")
	}
	if dryRun {
		fmt.Printf("npm: [dry run] npm %s (in %s)\n", strings.Join(args, " "), dir)
		return nil
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", npmError(out))
	}
	return nil
}

func npmError(out []byte) string {
	outStr := string(out)
	for line := range strings.SplitSeq(outStr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "npm ERR! code ") {
			continue
		}
		switch strings.TrimPrefix(line, "npm ERR! code ") {
		case "EOTP":
			return "2FA is blocking publish: generate and use a token with 2FA disabled"
		case "ENEEDAUTH", "E401":
			return "not authenticated: generate a token with npm and ensure it's configured correctly"
		case "E403":
			return "permission denied: ensure the token has write access to this package or org"
		case "E409", "EPUBLISHCONFLICT":
			return "version already exists: this version has already been published"
		case "ENOTFOUND", "ETIMEDOUT", "ECONNREFUSED":
			return "network error: unable to reach the npm registry, check your connection"
		case "EUSAGE":
			if strings.Contains(outStr, "provenance") {
				return "provenance is not supported outside of CI: use --provenance=false when publishing locally"
			}
		}
	}
	return "publish failed: " + strings.TrimSpace(outStr)
}

func pollUntilVisible(pkgName, version string) error {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", pkgName, version)
	deadline := time.Now().Add(registryPollTimeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(registryPollInterval)
	}

	return fmt.Errorf("package %s@%s not visible after %s", pkgName, version, registryPollTimeout)
}
