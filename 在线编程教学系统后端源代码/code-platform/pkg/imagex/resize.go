package imagex

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
)

type ImageType uint8

const (
	JPEG ImageType = iota
	PNG
	GIF
)

// Resize 调整图像大小；支持GIF/PNG/JPEG；width 或 height 为 0 时，保持原图像比例
func Resize(reader io.Reader, width int, height int, imageType ImageType) (io.Reader, int64, error) {
	var (
		img image.Image
		err error
	)
	switch imageType {
	case JPEG:
		img, err = jpeg.Decode(reader)
	case PNG:
		img, err = png.Decode(reader)
	case GIF:
		img, err = gif.Decode(reader)
	}

	if err != nil {
		return nil, 0, err
	}

	dst := imaging.Resize(img, width, height, imaging.Lanczos)
	buf := bytes.NewBuffer(nil)

	switch imageType {
	case JPEG:
		err = jpeg.Encode(buf, dst, nil)
	case PNG:
		err = png.Encode(buf, dst)
	case GIF:
		err = gif.Encode(buf, dst, nil)
	}

	if err != nil {
		return nil, 0, err
	}
	return buf, int64(buf.Len()), nil
}
