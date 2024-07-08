package ui

import (
	"io"

	"fyne.io/fyne/v2"
)

func makeURIReadCloser(r io.Reader) uriReadCloser {
	return uriReadCloser{io.NopCloser(r)}
}

type uriReadCloser struct {
	io.ReadCloser
}

func (uriReadCloser) URI() fyne.URI {
	return nil
}
