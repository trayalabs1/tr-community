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

func assertConvertsToJPEG(t *testing.T, raw []byte) {
	t.Helper()
	r := require.New(t)
	a := assert.New(t)

	out, size, err := ConvertToJPEG(bytes.NewReader(raw))
	r.NoError(err)

	outBytes, err := io.ReadAll(out)
	r.NoError(err)
	a.Equal(int64(len(outBytes)), size)

	img, format, err := image.Decode(bytes.NewReader(outBytes))
	r.NoError(err)
	a.Equal("jpeg", format)
	a.Positive(img.Bounds().Dx())
	a.Positive(img.Bounds().Dy())
}

func TestConvertToJPEG_HEIC(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.heic")
	require.NoError(t, err)
	assertConvertsToJPEG(t, raw)
}

func TestConvertToJPEG_AVIF(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.avif")
	require.NoError(t, err)
	assertConvertsToJPEG(t, raw)
}

func TestConvertToJPEG_TIFF(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 8, 8))
	buf := bytes.NewBuffer(nil)
	require.NoError(t, tiff.Encode(buf, src, nil))
	assertConvertsToJPEG(t, buf.Bytes())
}

func TestNeedsJPEGConversion(t *testing.T) {
	a := assert.New(t)

	for _, mt := range []string{"image/heic", "image/heif", "image/avif", "image/tiff"} {
		a.True(NeedsJPEGConversion(mime.New(mt)), mt)
	}
	for _, mt := range []string{"image/jpeg", "image/png", "image/webp", "image/gif"} {
		a.False(NeedsJPEGConversion(mime.New(mt)), mt)
	}
}

func TestConvertToJPEG_RejectsNonImage(t *testing.T) {
	_, _, err := ConvertToJPEG(bytes.NewReader([]byte("not an image")))
	assert.Error(t, err)
}

func encodeImage(t *testing.T, encode func(io.Writer, image.Image) error) []byte {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	require.NoError(t, encode(buf, image.NewRGBA(image.Rect(0, 0, 4, 4))))
	return buf.Bytes()
}

func TestConvertToJPEG_DecodesWebFriendly(t *testing.T) {
	a := assert.New(t)

	jpegBytes := encodeImage(t, func(w io.Writer, m image.Image) error { return jpeg.Encode(w, m, nil) })
	out, _, err := ConvertToJPEG(bytes.NewReader(jpegBytes))
	a.NoError(err)
	a.NotNil(out)

	pngBytes := encodeImage(t, png.Encode)
	out, _, err = ConvertToJPEG(bytes.NewReader(pngBytes))
	a.NoError(err)
	a.NotNil(out)
}
