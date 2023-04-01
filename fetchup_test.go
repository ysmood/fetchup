package fetchup_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/ysmood/fetchup"
)

func TestUnTar(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}

	u := s.URL("/tar-gz/t.tar.gz")
	d := getTmpDir(g)

	fu := fetchup.Fetchup{
		Logger: logger,
		To:     d,
	}
	g.E(fu.Download(u))

	g.Eq(g.Read(filepath.Join(d, "a", "t.txt")).Bytes(), data)
	g.Eq(logger.buf, fmt.Sprintf("Download: %s\nProgress: 19%%\nProgress: 39%%\nProgress: 60%%\nProgress: 80%%\nProgress: 100%%\nDownloaded: %s\n", u, d))
}

func TestMinReportSpan(t *testing.T) {
	g, s, _ := setup(t)

	logger := &bufLogger{}

	u := s.URL("/tar-gz/t.tar.gz")
	d := getTmpDir(g)
	fu := fetchup.Fetchup{
		Logger:        logger,
		MinReportSpan: time.Second,
		To:            d,
	}
	g.E(fu.Download(u))
	g.Eq(logger.buf, fmt.Sprintf("Download: %s\nProgress: 19%%\nProgress: 100%%\nDownloaded: %s\n", u, d))
}

func TestZip(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}
	u := s.URL("/zip/t.zip")
	d := getTmpDir(g)
	fu := fetchup.Fetchup{
		Logger: logger,
		To:     d,
	}
	g.E(fu.Download(u))
	g.Eq(g.Read(filepath.Join(d, "to", "file.txt")).Bytes(), data)
	g.Eq(logger.buf, fmt.Sprintf(`Download: %s
Progress: 02%%
Progress: 05%%
Progress: 10%%
Progress: 19%%
Progress: 40%%
Progress: 80%%
Progress: 100%%
Unzip: %s
99%%
100%%
Downloaded: %s
`, u, d, d))
}

func TestNew(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}

	u := s.URL("/tar-gz/t.tar.gz")
	d := getTmpDir(g)

	fu := fetchup.New(d, s.URL("/slow/"), u)
	fu.Logger = logger
	g.E(fu.Fetch())

	g.Eq(g.Read(filepath.Join(d, "a/t.txt")).Bytes(), data)
	g.Eq(logger.buf, g.Render(`Download: {{.U}}
Progress: 19%
Progress: 100%
Downloaded: {{.D}}
`, struct {
		U string
		D string
	}{u, d}).String())

	g.E(fetchup.StripFirstDir(d))
	g.True(g.PathExists(filepath.Join(d, "t.txt")))
}
