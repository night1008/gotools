package media

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MediaTypeImage = "image" // 图片
	MediaTypeVideo = "video" // 视频
)

type MediaInfo struct {
	Name      string  // 文件名
	Type      string  // 文件类型
	Size      int64   // 文件大小
	Width     int     // 宽度
	Height    int     // 长度
	Duration  float64 // 时长
	Extension string  // 文件扩展名
	Format    string
}

// 获取文件基础信息
func GetMediaBaseInfo(file *os.File) (*MediaInfo, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	fileStatInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %v", err)
	}

	return &MediaInfo{
		Name: fileStatInfo.Name(),
		Size: fileStatInfo.Size(),
	}, nil
}

// 获取文件完整信息
func GetMediaInfo(file *os.File) (*MediaInfo, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	fileStatInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %v", err)
	}

	var fileInfo *MediaInfo
	var mediaType string
	ext := GetMediaExtension(fileStatInfo.Name())
	switch ext {
	case ImageTypeJpg, ImageTypeJpeg, ImageTypePng,
		ImageTypeGif, ImageTypeBmp, ImageTypeTif,
		ImageTypeWebp:
		fileInfo, err = GetMediaImageInfo(file)
		if err != nil {
			return nil, err
		}
		mediaType = MediaTypeImage
	case VideoTypeMp4, VideoTypeMov, VideoTypeAvi,
		VideoTypeMkv, VideoTypeFlv, VideoTypeWmv,
		VideoTypeWebm, VideoType3gp:
		fileInfo, err = GetMediaVideoInfo(file)
		if err != nil {
			return nil, err
		}
		mediaType = MediaTypeVideo
	default:
		return nil, fmt.Errorf("unknown media extension %s", ext)
	}

	fileInfo.Type = mediaType
	fileInfo.Name = fileStatInfo.Name()
	fileInfo.Size = fileStatInfo.Size()
	fileInfo.Extension = ext
	return fileInfo, nil
}

// 获取文件 MD5
func GetMediaMD5Hash(file *os.File) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %v", err)
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5 hash: %v", err)
	}
	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))
	return md5Hash, nil
}

// 获取文件扩展名
func GetMediaExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// 获取文件缩略图
func GetMediaThumbnail(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	fileStatInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %v", err)
	}

	ext := GetMediaExtension(fileStatInfo.Name())
	switch ext {
	case ImageTypeJpg, ImageTypeJpeg, ImageTypePng, ImageTypeGif:
		fileInfo, err := GetMediaImageThumbnail(file)
		if err != nil {
			return nil, err
		}
		return fileInfo, nil
	case VideoTypeMp4, VideoTypeMov, VideoTypeAvi, VideoType3gp:
		fileInfo, err := GetMediaVideoFirstFrame(file)
		if err != nil {
			return nil, err
		}
		return fileInfo, nil
	default:
		return nil, fmt.Errorf("unknown media extension %s", ext)
	}
}
