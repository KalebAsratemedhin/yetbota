package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

// ProcessImage resizes and optimizes the image.
func ProcessImage(decoded []byte) ([]byte, string, error) {
	img, format, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid image format",
			ServiceMessage:    fmt.Sprintf("Unable to decode image: %s", err),
		}
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if format == "webp" && width <= constants.MaxImageResolution && height <= constants.MaxImageResolution {
		return decoded, "image/webp", nil
	}
	if width > constants.MaxImageResolution || height > constants.MaxImageResolution {
		img = resize.Thumbnail(constants.MaxImageResolution, constants.MaxImageResolution, img, resize.Lanczos3)
	}

	var buf bytes.Buffer
	var mime string
	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("Failed to encode JPEG: %s", err),
			}
		}
		mime = "image/jpeg"
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("Failed to encode PNG: %s", err),
			}
		}
		mime = "image/png"
	case "webp":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("Failed to encode WebP as JPEG: %s", err),
			}
		}
		mime = "image/jpeg"
	default:
		return nil, "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Unsupported image format",
			ServiceMessage:    fmt.Sprintf("Format %s not supported", format),
		}
	}

	return buf.Bytes(), mime, nil
}
