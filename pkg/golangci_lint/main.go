package golangci_lint

import (
	"github.com/ysmood/fetchup/pkg"
)

var DefaultOptions = pkg.Options{
	Version:   "2.1.6",
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

	return pkg.InstallWithOptions(opts)
}
