package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

// decodeWithOrientation decodes an image and applies EXIF orientation so the
// pixel data matches the intended viewing orientation before any processing.
func decodeWithOrientation(data []byte) (image.Image, string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid image format",
			ServiceMessage:    fmt.Sprintf("unable to decode image config: %s", err),
		}
	}
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid image format",
			ServiceMessage:    fmt.Sprintf("unable to decode image: %s", err),
		}
	}
	return img, format, nil
}

func encode(img image.Image, format string) ([]byte, string, error) {
	var buf bytes.Buffer
	var mime string
	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("failed to encode JPEG: %s", err),
			}
		}
		mime = "image/jpeg"
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("failed to encode PNG: %s", err),
			}
		}
		mime = "image/png"
	case "webp":
		// No CGO-free WebP encoder in the toolchain; transcode resized WebP to JPEG.
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Failed to process image",
				ServiceMessage:    fmt.Sprintf("failed to encode WebP as JPEG: %s", err),
			}
		}
		mime = "image/jpeg"
	default:
		return nil, "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Unsupported image format",
			ServiceMessage:    fmt.Sprintf("format %s not supported", format),
		}
	}
	return buf.Bytes(), mime, nil
}

// ProcessImage resizes and optimizes the image, respecting EXIF orientation.
func ProcessImage(decoded []byte) ([]byte, string, error) {
	img, format, err := decodeWithOrientation(decoded)
	if err != nil {
		return nil, "", err
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	if format == "webp" && w <= constants.MaxImageResolution && h <= constants.MaxImageResolution {
		return decoded, "image/webp", nil
	}
	if w > constants.MaxImageResolution || h > constants.MaxImageResolution {
		img = resize.Thumbnail(constants.MaxImageResolution, constants.MaxImageResolution, img, resize.Lanczos3)
	}

	return encode(img, format)
}

// ImageMimeType validates the image format and returns its MIME type without resizing.
func ImageMimeType(decoded []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(decoded))
	if err != nil {
		return "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid image format",
			ServiceMessage:    fmt.Sprintf("unable to decode image: %s", err),
		}
	}
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg", nil
	case "png":
		return "image/png", nil
	case "webp":
		return "image/webp", nil
	default:
		return "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Unsupported image format",
			ServiceMessage:    fmt.Sprintf("format %s not supported", format),
		}
	}
}

// CompressToMaxDim resizes img to fit within maxDim x maxDim, respecting EXIF orientation.
func CompressToMaxDim(decoded []byte, maxDim uint) ([]byte, string, error) {
	img, format, err := decodeWithOrientation(decoded)
	if err != nil {
		return nil, "", err
	}
	resized := resize.Thumbnail(maxDim, maxDim, img, resize.Lanczos3)
	return encode(resized, format)
}
