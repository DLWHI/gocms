package storage

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"

	"gocms/internal/domain"
	internalImage "gocms/internal/image"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/HugoSmits86/nativewebp"
)

func GeneratePreview(file io.Reader, width int, height int) ([]byte, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return nil, fmt.Errorf("%w: file is not an image", domain.ErrInval)
		}
		return nil, err
	}

	preview := internalImage.Contain(img, width, height)

	writer := &bytes.Buffer{}
	options := nativewebp.Options{
		CompressionLevel: nativewebp.DefaultCompression,
	}
	if err := nativewebp.Encode(writer, preview, &options); err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}
