package fetchup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
