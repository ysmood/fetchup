package golang_migrate

import (
	"github.com/ysmood/fetchup/pkg"
)

var DefaultOptions = pkg.Options{
	Version:   "4.18.3",
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

	return pkg.InstallWithOptions(opts)
}
