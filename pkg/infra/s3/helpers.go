package s3

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)
	// image.RegisterFormat("tiff", "tiff", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
}

func GetImageDimensions(body io.Reader) (*image.Config, error) {
	imgConf, _, err := image.DecodeConfig(body)
	if err != nil {
		return nil, err
	}
	return &imgConf, nil
}

func GenerateThumbnail(body io.Reader, width, height int, ext string) ([]byte, *image.Config, error) {
	img, err := imaging.Decode(body)
	if err != nil {
		return nil, nil, err
	}
	thumbnail := imaging.Resize(img, width, height, imaging.Lanczos)
	imagingFormat, err := imaging.FormatFromExtension(ext)
	if err == imaging.ErrUnsupportedFormat {
		imagingFormat = imaging.JPEG
	}
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumbnail, imagingFormat); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), &image.Config{
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}, nil
}
