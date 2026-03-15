package audios

import (
	"encoding/json"
	"errors"
	"mime"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
)

type Audio struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"index;not null"`      // 存储的文件名
	Ext         string    `json:"ext"`                             // 后缀
	MimeType    string    `json:"mime_type" gorm:"index"`          // 类型
	Title       string    `json:"title"`                           // 名称
	StorageType string    `json:"storage_type"`                    // 存储格式
	Size        int64     `json:"size" gorm:"default:0"`           // 大小
	Addtime     time.Time `json:"addtime"`                         // 添加时间
	Duration    int64     `json:"duration" gorm:"index;default:0"` // 时长,秒
	Bitrate     string    `json:"bitrate"`                         // 比特率
	SampleRate  string    `json:"sample_rate"`                     // 采样率
	Channels    string    `json:"channels"`                        // 声道
	Codec       string    `json:"codec"`                           // 编码
	Language    string    `json:"language" gorm:"index"`           // 语音
}

type AudioTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index"`
}

var dbs = db.MEDIADB

func init() {
	dbs.DB().AutoMigrate(&Audio{})
}

func AddAudio(src, title string) (*Audio, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "format=duration,bit_rate",
		"-show_entries", "stream=codec_name,sample_rate,channels",
		"-analyzeduration", "1000000",
		"-probesize", "1000000",
		src,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		Streams []struct {
			CodecName  string `json:"codec_name"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
		} `json:"streams"`

		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}

	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Streams) == 0 {
		return nil, errors.New("no audio stream")
	}

	durationFloat, _ := strconv.ParseFloat(result.Format.Duration, 64)

	audio := &Audio{
		Name:        filepath.Base(src),
		Title:       title,
		Ext:         filepath.Ext(src),
		MimeType:    mime.TypeByExtension(filepath.Ext(src)),
		StorageType: config.DefStorage,
		Addtime:     time.Now(),
		Duration:    int64(durationFloat),
		Bitrate:     result.Format.BitRate,
		SampleRate:  result.Streams[0].SampleRate,
		Channels:    strconv.Itoa(result.Streams[0].Channels),
		Codec:       result.Streams[0].CodecName,
	}

	return audio, nil
}
