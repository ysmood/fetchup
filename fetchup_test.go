package fetchup_test

import (
	"errors"
	"io"
	"log"
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

	fu := fetchup.New().WithLogger(logger)
	defer func() { g.E(os.RemoveAll(fu.SaveTo)) }()

	g.E(fu.Download(u))

	g.Eq(g.Read(filepath.Join(fu.SaveTo, "a", "t.txt")).Bytes(), data)
	g.Eq(logger.buf, g.Render(`Download: {{.U}}
Progress: 19%
Downloaded: {{.D}}
`, struct {
		U string
		D string
	}{u, fu.SaveTo}).String())
}

func TestZip(t *testing.T) {
	g, s, data := setup(t)

	logger := &bufLogger{}
	u := s.URL("/zip/t.zip")
	d := getTmpDir(g)
	fu := fetchup.New().WithSaveTo(d)
	fu.Logger = logger
	fu.MinReportSpan = 0
	g.E(fu.Download(u))
	g.Eq(g.Read(filepath.Join(d, "to", "file.txt")).Bytes(), data)
	g.Eq(logger.buf, g.Render(`Download: {{.U}}
Progress: 19%
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

	fu := fetchup.New(s.URL("/slow/"), u).WithSaveTo(d)
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

	fu := fetchup.New("").WithSaveTo(p)

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

	fu := fetchup.New("").WithSaveTo(p)
	fu.Logger = log.New(io.Discard, "", 0)

	g.E(fu.UnZip(g.Open(false, filepath.FromSlash("fixtures/test.zip"))))

	g.Eq(g.Read(filepath.FromSlash("tmp/t/t/test/b.txt")).String(), "test test")
}

func TestURLErr(t *testing.T) {
	g, s, _ := setup(t)

	fu := fetchup.New(s.URL("/err/")).WithSaveTo(getTmpDir(g))
	e := &fetchup.ErrNoURLs{}
	g.True(errors.As(fu.Fetch(), &e))
}

func TestGzipHttpBody(t *testing.T) {
	g, s, data := setup(t)

	p := filepath.Join(getTmpDir(g), "t.out")

	fu := fetchup.New(s.URL("/file/")).WithSaveTo(p)
	fu.Logger = log.New(io.Discard, "", 0)
	fu.SpeedPacketSize = 100
	g.E(fu.Fetch())

	g.Eq(g.Read(p).Bytes(), data)
}

func TestNoContentLength(t *testing.T) {
	g, s, data := setup(t)

	p := filepath.Join(getTmpDir(g), "t.out")

	fu := fetchup.New(s.URL("/no-content-length/")).WithSaveTo(p)
	fu.Logger = log.New(io.Discard, "", 0)
	fu.SpeedPacketSize = 100
	g.E(fu.Fetch())

	g.Eq(g.Read(p).Bytes(), data)
}

func TestContext(t *testing.T) {
	g, s, _ := setup(t)

	p := getTmpDir(g)

	ctx := g.Context()
	ctx.Cancel()

	fu := fetchup.New(s.URL("/slow/")).WithSaveTo(p)
	fu.Logger = log.New(io.Discard, "", 0)
	fu.Ctx = ctx
	g.Err(fu.Fetch())

	fu = fetchup.New(s.URL("/slow/")).WithSaveTo(p)
	fu.Logger = log.New(io.Discard, "", 0)
	fu.Ctx = ctx
	g.Err(fu.Download(s.URL("/slow/")))
}
