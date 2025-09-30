package golang_migrate

import (
	"os/exec"
	"strings"

	"github.com/ysmood/fetchup/pkg"
)

var DefaultOptions = pkg.Options{
	Version:   "4.19.0",
	URLs:      pkg.NewTemplates("https://github.com/golang-migrate/migrate/releases/download/v{{.Version}}/migrate.{{.OS}}-{{.Arch}}{{.BundleExt}}"),
	BundleBin: pkg.NewTemplates("migrate{{.ExecutableExt}}"),
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

	b, err := exec.Command(path, "-version").CombinedOutput()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(b)) == version
}
