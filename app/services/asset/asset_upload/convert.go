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

const jpegQuality = 85

// convertToJPEGMIMETypes are image formats that browsers cannot reliably
// display and which are transcoded to JPEG on upload. Mirrors the set of
// non-web-friendly types handled by traya-api-server's upload flow.
var convertToJPEGMIMETypes = map[string]struct{}{
	"image/heic":          {},
	"image/heif":          {},
	"image/heic-sequence": {},
	"image/heif-sequence": {},
	"image/avif":          {},
	"image/tiff":          {},
}

func needsJPEGConversion(mt *mime.Type) bool {
	_, ok := convertToJPEGMIMETypes[mt.String()]
	return ok
}

func maybeConvertToJPEG(mt *mime.Type, r io.Reader) (io.Reader, *mime.Type, int64, bool, error) {
	if !needsJPEGConversion(mt) {
		return r, mt, -1, false, nil
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, nil, 0, false, fault.Wrap(err)
	}

	buf := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, nil, 0, false, fault.Wrap(err)
	}

	jpegMIME := mime.New("image/jpeg")

	return buf, &jpegMIME, int64(buf.Len()), true, nil
}
