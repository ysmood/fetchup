package fetchup_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
Unzip: {{.D}}
Progress: 99%
Progress: 100%
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
Downloaded: {{.D}}
`, struct {
		U string
		D string
	}{u, d}).String())

	g.E(fetchup.StripFirstDir(d))
	g.True(g.PathExists(filepath.Join(d, "t.txt")))
}

func TestUnTarSymbolLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	g := got.T(t)

	p := filepath.FromSlash("tmp/t/t")

	g.E(os.RemoveAll(p))
	g.E(os.MkdirAll(p, 0755))

	fu := fetchup.New(p, "")

	g.E(fu.UnTar(g.Open(false, filepath.FromSlash("fixtures/test.tar"))))

	g.Eq(g.Read(filepath.FromSlash("tmp/t/t/test/b.txt")).String(), "test test")
}

func TestUnZipSymbolLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	g := got.T(t)

	p := filepath.FromSlash("tmp/t/t")

	g.E(os.RemoveAll(p))
	g.E(os.MkdirAll(p, 0755))

	fu := fetchup.New(p, "")

	g.E(fu.UnZip(g.Open(false, filepath.FromSlash("fixtures/test.zip"))))

	g.Eq(g.Read(filepath.FromSlash("tmp/t/t/test/b.txt")).String(), "test test")
}

func TestURLErr(t *testing.T) {
	g, s, _ := setup(t)

	fu := fetchup.New(getTmpDir(g), s.URL("/err/"))
	e := &fetchup.ErrNoURLs{}
	g.True(errors.As(fu.Fetch(), &e))
}

func TestGzipHttpBody(t *testing.T) {
	g, s, data := setup(t)

	p := filepath.Join(getTmpDir(g), "t.out")

	fu := fetchup.New(p, s.URL("/file/"))
	fu.SpeedPacketSize = 100
	g.E(fu.Fetch())

	g.Eq(g.Read(p).Bytes(), data)
}

func TestNoContentLength(t *testing.T) {
	g, s, data := setup(t)

	p := filepath.Join(getTmpDir(g), "t.out")

	fu := fetchup.New(p, s.URL("/no-content-length/"))
	fu.SpeedPacketSize = 100
	g.E(fu.Fetch())

	g.Eq(g.Read(p).Bytes(), data)
}

func TestContext(t *testing.T) {
	g, s, _ := setup(t)

	p := getTmpDir(g)

	ctx := g.Context()
	ctx.Cancel()

	fu := fetchup.New(p, s.URL("/slow/"))
	fu.Ctx = ctx
	g.Err(fu.Fetch())

	fu = fetchup.New(p, s.URL("/slow/"))
	fu.Ctx = ctx
	g.Err(fu.Download(s.URL("/slow/")))
}
