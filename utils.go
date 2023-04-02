package fetchup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Logger interface
type Logger interface {
	// Same as fmt.Printf
	Println(...interface{})
}

// Log type for Println
type Log func(msg ...interface{})

// Println interface
func (l Log) Println(msg ...interface{}) {
	l(msg...)
}

// LoggerQuiet does nothing
var LoggerQuiet Logger = Log(func(_ ...interface{}) {})

// MultiLogger is similar to https://golang.org/pkg/io/#MultiWriter
func MultiLogger(list ...Logger) Log {
	return Log(func(msg ...interface{}) {
		for _, lg := range list {
			lg.Println(msg...)
		}
	})
}

type progress struct {
	s       io.Reader
	total   int
	count   int
	logger  Logger
	last    time.Time
	minSpan time.Duration
}

var _ io.ReadWriter = &progress{}

// NewProgress returns a new progress reader.
func NewProgress(s io.Reader, total int, minSpan time.Duration, logger Logger) *progress {
	return &progress{
		s:       s,
		total:   total,
		logger:  logger,
		minSpan: minSpan,
	}
}

func (p *progress) Read(b []byte) (n int, err error) {
	n, err = p.s.Read(b)

	p.count += n

	if p.count == p.total {
		p.logger.Println(EventProgress, "100%")
		return
	}

	if time.Since(p.last) < p.minSpan {
		return
	}

	p.last = time.Now()
	p.logger.Println(EventProgress, fmt.Sprintf("%02d%%", p.count*100/p.total))

	return
}

func (p *progress) Write(b []byte) (n int, err error) {
	n = len(b)

	p.count += n

	if p.count == p.total {
		p.logger.Println("100%")
		return
	}

	if time.Since(p.last) < p.minSpan {
		return
	}

	p.last = time.Now()
	p.logger.Println(fmt.Sprintf("%02d%%", p.count*100/p.total))

	return
}

func CacheDir() string {
	return filepath.Join(map[string]string{
		"windows": filepath.Join(os.Getenv("APPDATA")),
		"darwin":  filepath.Join(os.Getenv("HOME"), ".cache"),
		"linux":   filepath.Join(os.Getenv("HOME"), ".cache"),
	}[runtime.GOOS])
}

// StripFirstDir removes the first dir but keep all its children.
func StripFirstDir(dir string) error {
	list, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	name := ""
	for _, f := range list {
		if f.IsDir() {
			if name != "" {
				return fmt.Errorf("expected only one dir in %s", dir)
			}

			name = f.Name()
			continue
		}
	}

	if name == "" {
		return fmt.Errorf("no dir found under %s", dir)
	}

	root := filepath.Join(dir, name)

	children, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, child := range children {
		err = os.Rename(filepath.Join(root, child.Name()), filepath.Join(dir, child.Name()))
		if err != nil {
			return err
		}
	}

	return os.Remove(root)
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", string(filepath.Separator))
	return strings.ReplaceAll(p, "/", string(filepath.Separator))
}
