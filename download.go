package fetchup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (fu *Fetchup) Downloader(u string) (io.Reader, func(), error) {
	fu.Logger.Println(EventDownload, u)

	q, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	res, err := (&http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			IdleConnTimeout:   fu.IdleConnTimeout,
		},
	}).Do(q)
	if err != nil {
		return nil, nil, err
	}

	size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, nil, err
	}

	return NewProgress(res.Body, int(size), fu.MinReportSpan, fu.Logger), func() { _ = res.Body.Close() }, nil
}

func (fu *Fetchup) Download(u string) error {
	r, close, err := fu.Downloader(u)
	if err != nil {
		return err
	}
	defer close()

	if strings.HasSuffix(u, ".gz") {
		u = strings.TrimSuffix(u, ".gz")
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(fu.To, 0755)
	if err != nil {
		return err
	}

	if strings.HasSuffix(u, ".tar") {
		err := fu.UnTar(r)
		if err != nil {
			return err
		}
	}

	if strings.HasSuffix(u, ".zip") {
		err := fu.UnZip(r)
		if err != nil {
			return err
		}
	}

	fu.Logger.Println(EventDownloaded, fu.To)

	return nil
}

func (fu *Fetchup) UnZip(r io.Reader) error {
	// Because zip format does not streaming, we need to read the whole file into memory.
	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}

	size := 0
	for _, f := range zr.File {
		size += int(f.FileInfo().Size())
	}

	fu.Logger.Println(EventUnzip, fu.To)

	progress := NewProgress(r, size, fu.MinReportSpan, fu.Logger)

	for _, f := range zr.File {
		name := strings.ReplaceAll(f.Name, "\\", string(filepath.Separator))
		name = strings.ReplaceAll(name, "/", string(filepath.Separator))
		p := filepath.Join(fu.To, name)

		if f.FileInfo().IsDir() {
			err := os.Mkdir(p, f.Mode())
			if err != nil {
				return err
			}
			continue
		}

		r, err := f.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(io.MultiWriter(dst, progress), r)
		if err != nil {
			return err
		}

		err = dst.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (fu *Fetchup) UnTar(r io.Reader) error {
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		info := hdr.FileInfo()
		p := filepath.Join(fu.To, hdr.Name)

		if info.IsDir() {
			err = os.Mkdir(p, info.Mode())
			if err != nil {
				return err
			}

			continue
		}

		dst, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(dst, tr)
		if err != nil {
			return err
		}

		err = dst.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
