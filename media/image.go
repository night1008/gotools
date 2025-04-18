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

const MediaImageThumbnailJpegQuality = 20 // jpeg 缩略图质量

// 获取图片信息
func GetMediaImageInfo(file *os.File) (*MediaInfo, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	img, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	var duration float64
	switch format {
	case "gif":
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

	if img.Width == 0 || img.Height == 0 {
		return nil, fmt.Errorf("image height or width is zero")
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
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	var buf bytes.Buffer
	switch format {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(file)
		if err != nil {
			return nil, err
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: MediaImageThumbnailJpegQuality,
		}); err != nil {
			return nil, err
		}
	case "png":
		img, err := png.Decode(file)
		if err != nil {
			return nil, err
		}
		encoder := png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		if err := encoder.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "gif":
		return ExtractMediaGifImageFirstFrame(file)
	default:
		return nil, fmt.Errorf("unsupport get image type %s thumbnail", format)
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
