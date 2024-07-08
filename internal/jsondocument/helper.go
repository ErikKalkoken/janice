package jsondocument

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// Creates and returns an URIReadCloser from a reader.
func MakeURIReadCloser(r io.Reader, name string) uriReadCloser {
	return uriReadCloser{io.NopCloser(r), name}
}

type uriReadCloser struct {
	io.ReadCloser
	name string
}

func (u uriReadCloser) URI() fyne.URI {
	uri, _ := storage.ParseURI(fmt.Sprintf("dummy://dummy/%s", u.name))
	return uri
}
