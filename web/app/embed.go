package web

import (
	"compress/gzip"
	"embed"
	"io/fs"
)

//go:embed all:dist
var assets embed.FS

// Dist returns the built asset tree. The build stores files gzipped
// (dist/x.js.gz); anything uncompressed is served as-is.
func Dist() (fs.FS, error) {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil, err
	}
	return gzFS{sub}, nil
}

type gzFS struct{ fs.FS }

func (f gzFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name + ".gz")
	if err != nil {
		return f.FS.Open(name)
	}
	return &gzFile{File: file}, nil
}

type gzFile struct {
	fs.File
	zr *gzip.Reader
}

func (f *gzFile) Read(p []byte) (int, error) {
	if f.zr == nil {
		zr, err := gzip.NewReader(f.File)
		if err != nil {
			return 0, err
		}
		f.zr = zr
	}
	return f.zr.Read(p)
}

func (f *gzFile) Close() error {
	if f.zr != nil {
		_ = f.zr.Close()
	}
	return f.File.Close()
}
