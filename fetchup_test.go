package fetchup_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ysmood/fetchup"
	"github.com/ysmood/got"
)

func TestUnTar(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}

	u := s.URL("/tar-gz/t.tar.gz")
	d := getTmpDir(g)

	fu := fetchup.New(d)
	fu.Logger = logger
	g.E(fu.Download(u))

	g.Eq(g.Read(filepath.Join(d, "a", "t.txt")).Bytes(), data)
	g.Eq(logger.buf, g.Render(`Download: {{.U}}
Progress: 19%
Progress: 100%
Downloaded: {{.D}}
`, struct {
		U string
		D string
	}{u, d}).String())
}

func TestZip(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}
	u := s.URL("/zip/t.zip")
	d := getTmpDir(g)
	fu := fetchup.New(d)
	fu.Logger = logger
	fu.MinReportSpan = 0
	g.E(fu.Download(u))
	g.Eq(g.Read(filepath.Join(d, "to", "file.txt")).Bytes(), data)
	g.Eq(logger.buf, g.Render(`Download: {{.U}}
Progress: 02%
Progress: 05%
Progress: 10%
Progress: 19%
Progress: 40%
Progress: 80%
Progress: 100%
Unzip: {{.D}}
99%
100%
Downloaded: {{.D}}
`, struct {
		U string
		D string
	}{u, d}).String())
}

func TestNew(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}

	u := s.URL("/tar-gz/t.tar.gz")
	d := getTmpDir(g)

	fu := fetchup.New(d, s.URL("/slow/"), u)
	fu.Logger = logger
	fu.SpeedPacketSize = 100
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

func TestUnTarSymbolLink(t *testing.T) {
	g := got.T(t)

	g.E(os.RemoveAll("tmp/t/t"))
	g.E(os.MkdirAll("tmp/t/t", 0755))

	fu := fetchup.New("tmp/t/t", "")

	g.E(fu.UnTar(g.Open(false, "features/test.tar")))

	g.Eq(g.Read("tmp/t/t/test/b.txt").String(), "test test")
}

func TestUnZipSymbolLink(t *testing.T) {
	g := got.T(t)

	g.E(os.RemoveAll("tmp/t/t"))
	g.E(os.MkdirAll("tmp/t/t", 0755))

	fu := fetchup.New("tmp/t/t", "")

	g.E(fu.UnZip(g.Open(false, "features/test.zip")))

	g.Eq(g.Read("tmp/t/t/test/b.txt").String(), "test test")
}

func TestURLErr(t *testing.T) {
	g, s, _ := setup(t)

	fu := fetchup.New(getTmpDir(g), s.URL("/err/"))
	e := &fetchup.ErrNoURLs{}
	g.True(errors.As(fu.Fetch(), &e))
}
