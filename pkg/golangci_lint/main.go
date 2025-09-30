package golangci_lint

import (
	"os/exec"
	"strings"

	"github.com/ysmood/fetchup/pkg"
)

var DefaultOptions = pkg.Options{
	Version:   "2.5.0",
	URLs:      pkg.NewTemplates("https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}{{.BundleExt}}"),
	BundleBin: pkg.NewTemplates("golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}", "golangci-lint{{.ExecutableExt}}"),
}

func Install() error {
	return InstallWithOptions(DefaultOptions)
}

func InstallWithOptions(opts pkg.Options) error {
	if opts.Version == "" {
		opts.Version = DefaultOptions.Version
	}

	if opts.URLs == nil {
		opts.URLs = DefaultOptions.URLs
	}

	if opts.BundleBin == nil {
		opts.BundleBin = DefaultOptions.BundleBin
	}

	opts.Exists = func(path string) bool {
		return exists(path, opts.Version)
	}

	return pkg.InstallWithOptions(opts)
}

func exists(path, version string) bool {
	if !pkg.ExecExists(path) {
		return false
	}

	b, err := exec.Command(path, "version", "--short").CombinedOutput()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(b)) == version
}
