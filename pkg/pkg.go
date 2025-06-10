package pkg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/ysmood/fetchup"
)

type Options struct {
	Ctx    context.Context
	Logger fetchup.Logger

	// InstallToDir is the directory to install the binary to.
	InstallToDir string

	// Version is a shortcut to set Version argument in the TemplateArgs.
	Version string

	// URLs is a list of URLs to download the bundle from.
	URLs []Template

	// BundleBin is the path to the binary inside the bundle, the bundle is a tar or zip file.
	// Each item will be joined with the OS-specific path separator.
	BundleBin []Template

	// ExecutableName is the name of the executable after installation,
	// by default it's the same as the last part of BundleBin.
	ExecutableName Template

	// TemplateArgs are the arguments to render the any templates in the options.
	// It will set some default values like OS, Arch, BundleExt, and ExecutableExt,
	// check the code of [SetDefaultTemplateArgs] for more details.
	TemplateArgs map[string]any
}

var DefaultOptions = Options{
	Ctx:          context.Background(),
	InstallToDir: "./bin",
	Logger:       log.New(os.Stderr, "", log.LstdFlags),
}

func Defaults(opts Options) Options {
	if opts.Ctx == nil {
		opts.Ctx = DefaultOptions.Ctx
	}

	if opts.InstallToDir == "" {
		opts.InstallToDir = DefaultOptions.InstallToDir
	}

	if opts.Logger == nil {
		opts.Logger = DefaultOptions.Logger
	}

	if opts.TemplateArgs == nil {
		opts.TemplateArgs = map[string]any{}
	}

	SetDefaultTemplateArgs(opts)

	return opts
}

func InstallWithOptions(opts Options) error {
	opts = Defaults(opts)

	urls := []string{}
	for _, urlTpl := range opts.URLs {
		url, err := urlTpl.Render(opts.TemplateArgs)
		if err != nil {
			return fmt.Errorf("failed to render URL template: %w", err)
		}
		urls = append(urls, url)
	}

	if len(opts.BundleBin) == 0 {
		return fmt.Errorf("no bundle binary specified, please set BundleBin option")
	}

	bundleBin := []string{}
	for _, item := range opts.BundleBin {
		sect, err := item.Render(opts.TemplateArgs)
		if err != nil {
			return fmt.Errorf("failed to render bundle binary path section: %w", err)
		}

		bundleBin = append(bundleBin, sect)
	}

	exeName := filepath.Base(bundleBin[len(bundleBin)-1])
	if !opts.ExecutableName.IsZero() {
		var err error
		exeName, err = opts.ExecutableName.Render(opts.TemplateArgs)
		if err != nil {
			return fmt.Errorf("failed to render executable name: %w", err)
		}
	}

	f := fetchup.New(urls...).WithContext(opts.Ctx).WithLogger(opts.Logger)
	f = f.WithSaveTo(f.SaveTo + "-" + stripExt(exeName))

	err := f.Fetch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(f.SaveTo) }()

	bin := filepath.Join(f.SaveTo, filepath.Join(bundleBin...))

	err = os.MkdirAll(opts.InstallToDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", opts.InstallToDir, err)
	}

	err = os.Rename(bin, filepath.Join(opts.InstallToDir, exeName))
	if err != nil {
		return fmt.Errorf("failed to move binary to %s: %w", opts.InstallToDir, err)
	}

	return nil
}

type Template struct {
	tpl *template.Template
}

func NewTemplate(tpl string) Template {
	return Template{
		tpl: template.Must(template.New("tpl").Parse(tpl)),
	}
}

func NewTemplates(list ...string) []Template {
	templates := make([]Template, len(list))
	for i, tpl := range list {
		templates[i] = NewTemplate(tpl)
	}
	return templates
}

func (t Template) IsZero() bool {
	return t.tpl == nil
}

func (t Template) Render(data map[string]any) (string, error) {
	buf := bytes.NewBuffer(nil)

	err := t.tpl.Execute(buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func SetDefaultTemplateArgs(opts Options) {
	opts.TemplateArgs["Version"] = opts.Version
	opts.TemplateArgs["OS"] = runtime.GOOS
	opts.TemplateArgs["Arch"] = runtime.GOARCH
	opts.TemplateArgs["BundleExt"] = BundleExt()
	opts.TemplateArgs["ExecutableExt"] = ExecutableExt()
}

// ExecutableExt returns ".exe" for Windows and an empty string for Unix-like systems.
func ExecutableExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}

	return ""
}

// BundleExt returns ".tar.gz" for Unix-like systems and ".zip" for Windows.
func BundleExt() string {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	return ext
}

func stripExt(name string) string {
	if ext := filepath.Ext(name); ext != "" {
		return name[:len(name)-len(ext)]
	}
	return name
}
