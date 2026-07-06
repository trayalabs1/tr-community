package asset_upload

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/Southclaws/fault"

	_ "github.com/gen2brain/avif"
	_ "github.com/gen2brain/heic"
	_ "golang.org/x/image/tiff"

	"github.com/Southclaws/storyden/internal/mime"
)

const JPEGQuality = 85

// ConvertToJPEGMIMETypes are image formats that browsers cannot reliably
// display and which are transcoded to JPEG after upload. Mirrors the set of
// non-web-friendly types handled by traya-api-server's upload flow.
var ConvertToJPEGMIMETypes = map[string]struct{}{
	"image/heic":          {},
	"image/heif":          {},
	"image/heic-sequence": {},
	"image/heif-sequence": {},
	"image/avif":          {},
	"image/tiff":          {},
}

func NeedsJPEGConversion(mt mime.Type) bool {
	_, ok := ConvertToJPEGMIMETypes[mt.String()]
	return ok
}

// ConvertToJPEG decodes any supported non-web-friendly image and re-encodes it
// as JPEG. Decoding is memory-heavy (the decoders run in a WASM runtime), so
// this must run off the request path.
func ConvertToJPEG(r io.Reader) (io.Reader, int64, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, 0, fault.Wrap(err)
	}

	buf := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, 0, fault.Wrap(err)
	}

	return buf, int64(buf.Len()), nil
}
