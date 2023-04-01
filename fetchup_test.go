package fetchup_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ysmood/fetchup"
	"github.com/ysmood/got"
)

func TestDownload(t *testing.T) {
	g, s, data := setup(t)

	{
		logger := &bufLogger{}

		u := s.URL("/tar-gz/t.tar.gz")
		d := getTmpDir(g)

		fu := fetchup.Fetchup{
			Logger: logger,
			To:     d,
		}
		g.E(fu.Download(u))

		g.Eq(g.Read(filepath.Join(d, "a/t.txt")).Bytes(), data)
		g.Eq(logger.buf, fmt.Sprintf("Download: %s\nProgress: 19%%\nProgress: 39%%\nProgress: 60%%\nProgress: 80%%\nProgress: 100%%\nDownloaded: %s\n", u, d))
	}

	{
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

	{
		logger := &bufLogger{}
		u := s.URL("/zip/t.zip")
		d := getTmpDir(g)
		fu := fetchup.Fetchup{
			Logger: logger,
			To:     d,
		}
		g.E(fu.Download(u))
		g.Eq(g.Read(filepath.Join(d, "to/file.txt")).Bytes(), data)
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

	{
		d := g.RandStr(8)
		fu := fetchup.New([]string{s.URL("/slow/t.zip"), s.URL("/tar-gz/t.tar.gz")}, d)
		g.Cleanup(func() {
			_ = os.RemoveAll(fu.To)
		})

		g.E(fu.Fetch())
		g.PathExists(filepath.Join(d, "a/t.txt"))
	}
}

type bufLogger struct {
	buf string
}

var _ fetchup.Logger = (*bufLogger)(nil)

func (b *bufLogger) Println(msg ...interface{}) {
	b.buf += fmt.Sprintln(msg...)
}

func getTmpDir(g got.G) string {
	return filepath.Join("tmp", g.RandStr(8))
}

func setup(t *testing.T) (got.G, *got.Router, []byte) {
	g := got.T(t)

	data := make([]byte, 20000)
	g.E(rand.Read(data))

	s := g.Serve()

	s.Mux.HandleFunc("/tar-gz/", func(rw http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(nil)
		gz := gzip.NewWriter(buf)

		tw := tar.NewWriter(gz)
		g.E(tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "a",
			Mode:     0755,
		}))
		g.E(tw.Write(nil))

		g.E(tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "a/t.txt",
			Mode:     0644,
			Size:     int64(len(data)),
		}))
		g.E(tw.Write(data))
		g.E(tw.Close())
		g.E(gz.Close())

		rw.Header().Add("Content-Length", fmt.Sprintf("%d", buf.Len()))
		g.E(io.Copy(rw, buf))
	})

	s.Mux.HandleFunc("/zip/", func(rw http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(nil)
		zw := zip.NewWriter(buf)

		// folder "to"
		h := &zip.FileHeader{Name: "to/"}
		h.SetMode(0755)
		_, err := zw.CreateHeader(h)
		g.E(err)

		// file "file.txt"
		w, err := zw.CreateHeader(&zip.FileHeader{Name: "to/file.txt"})
		g.E(err)
		g.E(w.Write(data))

		// file "file2.txt"
		w, err = zw.CreateHeader(&zip.FileHeader{Name: "to/file2.txt"})
		g.E(err)
		g.E(w.Write([]byte("ok")))

		g.E(zw.Close())

		rw.Header().Add("Content-Length", fmt.Sprintf("%d", buf.Len()))
		_, _ = io.Copy(rw, buf)
	})

	s.Mux.HandleFunc("/slow/", func(rw http.ResponseWriter, r *http.Request) {
		t := time.NewTimer(3 * time.Second)
		select {
		case <-t.C:
		case <-r.Context().Done():
			t.Stop()
		}
	})

	return g, s, data
}
