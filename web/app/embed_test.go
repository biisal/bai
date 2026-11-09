package web

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestGzFSServesGzippedAssets(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	fsys := gzFS{fstest.MapFS{"index.html.gz": &fstest.MapFile{Data: buf.Bytes()}}}

	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q, want %q", data, "hello")
	}
}

func TestGzFSFallsBackToPlainFiles(t *testing.T) {
	fsys := gzFS{fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("plain")}}}

	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "plain" {
		t.Fatalf("got %q, want %q", data, "plain")
	}
}

func TestDistServesBuiltIndex(t *testing.T) {
	fsys, err := Dist()
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		t.Skip("web bundle not built")
	}
	if !bytes.Contains(data, []byte(`<div id="root">`)) {
		t.Fatalf("index.html does not look like the built app: %q", data)
	}
}
