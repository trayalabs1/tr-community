package asset_upload

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/tiff"

	"github.com/Southclaws/storyden/internal/mime"
)

func assertConvertsToJPEG(t *testing.T, mimeType string, raw []byte) {
	t.Helper()
	r := require.New(t)
	a := assert.New(t)

	mt := mime.New(mimeType)

	out, outMT, size, converted, err := maybeConvertToJPEG(&mt, bytes.NewReader(raw))
	r.NoError(err)
	a.True(converted)
	a.Equal("image/jpeg", outMT.String())

	outBytes, err := io.ReadAll(out)
	r.NoError(err)
	a.Equal(int64(len(outBytes)), size)

	img, format, err := image.Decode(bytes.NewReader(outBytes))
	r.NoError(err)
	a.Equal("jpeg", format)
	a.Positive(img.Bounds().Dx())
	a.Positive(img.Bounds().Dy())
}

func TestMaybeConvertToJPEG_HEIC(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.heic")
	require.NoError(t, err)
	assertConvertsToJPEG(t, "image/heic", raw)
}

func TestMaybeConvertToJPEG_AVIF(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.avif")
	require.NoError(t, err)
	assertConvertsToJPEG(t, "image/avif", raw)
}

func TestMaybeConvertToJPEG_TIFF(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 8, 8))
	buf := bytes.NewBuffer(nil)
	require.NoError(t, tiff.Encode(buf, src, nil))
	assertConvertsToJPEG(t, "image/tiff", buf.Bytes())
}

func TestMaybeConvertToJPEG_PassesThroughJPEG(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	buf := bytes.NewBuffer(nil)
	r.NoError(jpeg.Encode(buf, src, nil))
	original := buf.Bytes()

	mt := mime.New("image/jpeg")

	out, outMT, size, converted, err := maybeConvertToJPEG(&mt, bytes.NewReader(original))
	r.NoError(err)
	a.False(converted)
	a.Equal("image/jpeg", outMT.String())
	a.Equal(int64(-1), size)

	passed, err := io.ReadAll(out)
	r.NoError(err)
	a.Equal(original, passed)
}

func TestMaybeConvertToJPEG_PassesThroughPNG(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	buf := bytes.NewBuffer(nil)
	r.NoError(png.Encode(buf, src))
	original := buf.Bytes()

	mt := mime.New("image/png")

	out, outMT, _, converted, err := maybeConvertToJPEG(&mt, bytes.NewReader(original))
	r.NoError(err)
	a.False(converted)
	a.Equal("image/png", outMT.String())

	passed, err := io.ReadAll(out)
	r.NoError(err)
	a.Equal(original, passed)
}
