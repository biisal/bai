package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
)

func binaryFileName() string {
	if runtime.GOOS == "windows" {
		return binaryName + ".exe"
	}
	return binaryName
}

func extractBinary(archivePath, archiveName, destPath string) error {
	switch {
	case strings.HasSuffix(archiveName, ".tar.gz"):
		return extractTarGz(archivePath, destPath)
	case strings.HasSuffix(archiveName, ".zip"):
		return extractZip(archivePath, destPath)
	default:
		return fmt.Errorf("unsupported archive format: %s", archiveName)
	}
}

func extractTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s not found in archive", binaryFileName())
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg && path.Base(hdr.Name) == binaryFileName() {
			return writeExecutable(destPath, tr)
		}
	}
}

func extractZip(archivePath, destPath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()

	for _, zf := range zr.File {
		if zf.Mode().IsRegular() && path.Base(zf.Name) == binaryFileName() {
			return extractZipFile(zf, destPath)
		}
	}
	return fmt.Errorf("%s not found in archive", binaryFileName())
}

func extractZipFile(zf *zip.File, destPath string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	return writeExecutable(destPath, rc)
}

func writeExecutable(destPath string, r io.Reader) (err error) {
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, r)
	return err
}
