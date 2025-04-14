package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	VideoTypeMp4  = ".mp4"
	VideoTypeMov  = ".mov"
	VideoTypeAvi  = ".avi"
	VideoTypeMkv  = ".mkv"
	VideoTypeFlv  = ".flv"
	VideoTypeWmv  = ".wmv"
	VideoTypeWebm = ".webm"
	VideoTypeM4v  = ".m4v"
	VideoTypeMpg  = ".mpg"
	VideoType3gp  = ".3gp"
)

type VideoInfoFromFfprobe struct {
	Streams []struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Duration string `json:"duration"`
	} `json:"streams"`
}

// 获取视频信息
func GetMediaVideoInfo(file *os.File) (*MediaInfo, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-of", "json",
		"-", // Output to stdout
	)
	cmd.Stdin = file
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var videoInfoFromFfprobe VideoInfoFromFfprobe
	if err := json.Unmarshal(out.Bytes(), &videoInfoFromFfprobe); err != nil {
		return nil, err
	}
	if len(videoInfoFromFfprobe.Streams) == 0 {
		return nil, fmt.Errorf("vedio streams length is zero")
	}

	duration, err := strconv.ParseFloat(videoInfoFromFfprobe.Streams[0].Duration, 64)
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		Width:    videoInfoFromFfprobe.Streams[0].Width,
		Height:   videoInfoFromFfprobe.Streams[0].Height,
		Duration: duration,
	}, nil
}

// 获取视频首帧
func GetMediaVideoFirstFrame(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	baseFileName := filepath.Base(file.Name())
	ext := GetMediaExtension(baseFileName)
	tmpInputFileName := fmt.Sprintf("%s_*%s", strings.TrimRight(baseFileName, ext), ext)

	tmpInputFile, err := os.CreateTemp("", tmpInputFileName)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpInputFile.Name())

	// Copy content
	if _, err := io.Copy(tmpInputFile, file); err != nil {
		return nil, err
	}
	tmpInputFile.Close()

	// 生成临时文件路径
	tmpOutPath := fmt.Sprintf("%s_%d.tmp", tmpInputFile.Name(), time.Now().UnixNano())
	cmd := exec.Command(
		"ffmpeg",
		"-i", tmpInputFile.Name(), // input file
		"-loglevel", "info", // 日志等级
		"-f", "image2", // 指定输出为图像格式
		"-c:v", "mjpeg", // 使用JPEG编码
		"-vframes", "1", // only capture 1 frame
		"-q:v", "5", // quality (1=best, 31=worst)
		tmpOutPath, // output to temp file
	)

	// Capture ffmpeg output to stderr
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %v", err)
	}
	defer os.Remove(tmpOutPath)

	imgBuffer, err := os.ReadFile(tmpOutPath)
	if err != nil {
		return nil, err
	}
	if len(imgBuffer) == 0 {
		return nil, fmt.Errorf("video first frame is empty")
	}

	return imgBuffer, nil
}
