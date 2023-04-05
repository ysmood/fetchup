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
	"path/filepath"
	"testing"
	"time"

	"github.com/ysmood/fetchup"
	"github.com/ysmood/got"
)

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

	s.Mux.HandleFunc("/file/", func(rw http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(nil)
		gz := gzip.NewWriter(buf)
		g.E(gz.Write(data))
		g.E(gz.Close())

		rw.Header().Add("Content-Length", fmt.Sprintf("%d", buf.Len()))
		rw.Header().Add("Content-Encoding", "gzip")
		g.E(io.Copy(rw, buf))
	})

	s.Mux.HandleFunc("/no-content-length/", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		g.E(rw.Write(data))
	})

	s.Mux.HandleFunc("/err/", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})

	return g, s, data
}
