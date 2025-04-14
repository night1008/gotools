package media

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"time"
)

const (
	ImageTypeJpg  = ".jpg"
	ImageTypeJpeg = ".jpeg"
	ImageTypePng  = ".png"
	ImageTypeGif  = ".gif"
	ImageTypeBmp  = ".bmp"
	ImageTypeTif  = ".tif"
	ImageTypeWebp = ".webp"
	ImageTypeCr2  = ".cr2"
	ImageTypeHeif = ".heif"
	ImageTypeJxr  = ".jxr"
	ImageTypeIco  = ".ico"
	ImageType     = ".dwg"
	ImageTypeAvif = ".avif"
)

const MediaImageThumbnailJpegQuality = 10 // jpeg 缩略图质量

// 获取图片信息
func GetMediaImageInfo(file *os.File) (*MediaInfo, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	img, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	ext := GetMediaExtension(file.Name())
	var duration float64
	switch ext {
	case ImageTypeGif:
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to reset file pointer: %v", err)
		}

		gifImg, err := gif.DecodeAll(file)
		if err != nil {
			return nil, err
		}

		var totalDuration time.Duration
		for _, delay := range gifImg.Delay {
			totalDuration += time.Duration(delay) * 10 * time.Millisecond
		}
		duration = totalDuration.Seconds()
	}

	return &MediaInfo{
		Width:    img.Width,
		Height:   img.Height,
		Duration: duration,
		Format:   format,
	}, nil
}

// 获取图片缩略图
func GetMediaImageThumbnail(file *os.File) ([]byte, error) {
	ext := GetMediaExtension(file.Name())

	var buf bytes.Buffer
	switch ext {
	case ImageTypeJpg, ImageTypeJpeg:
		img, err := jpeg.Decode(file)
		if err != nil {
			return nil, err
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: MediaImageThumbnailJpegQuality,
		}); err != nil {
			return nil, err
		}
	case ImageTypePng:
		img, err := png.Decode(file)
		if err != nil {
			return nil, err
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: MediaImageThumbnailJpegQuality,
		}); err != nil {
			return nil, err
		}
	case ImageTypeGif:
		return ExtractMediaGifImageFirstFrame(file)
	default:
		return nil, fmt.Errorf("unsupport get image type %s thumbnail", "")
	}

	return buf.Bytes(), nil
}

// 获取动态图首帧
func ExtractMediaGifImageFirstFrame(file *os.File) ([]byte, error) {
	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, gifImg.Image[0], &jpeg.Options{
		Quality: MediaImageThumbnailJpegQuality,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
