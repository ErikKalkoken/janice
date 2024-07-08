package jsondocument_test

import (
	"strings"
	"testing"

	"github.com/ErikKalkoken/jsonviewer/internal/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestHelper(t *testing.T) {
	r := strings.NewReader("test")
	r2 := jsondocument.MakeURIReadCloser(r, "alpha")
	assert.Equal(t, "alpha", r2.URI().Name())
}
