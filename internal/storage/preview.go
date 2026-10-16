package storage

import (
	"errors"
	"fmt"
	"gocms/internal/errs"
	"image"
	"io"

	"github.com/chai2010/webp"

	internalImage "gocms/internal/image"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func GeneratePreview(file io.Reader, width int, height int) ([]byte, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return nil, fmt.Errorf("%w: file is not an image", errs.ErrInval)
		}
		return nil, err
	}

	preview := internalImage.Contain(img, width, height)

	data, err := webp.EncodeRGBA(preview, 0.7)
	if err != nil {
		return nil, err
	}

	return data, nil
}
