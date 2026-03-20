package audios

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/funcs"
	"tools/runtimes/storage"
)

type Audio struct {
	ID           int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string      `json:"name" gorm:"index;not null"`                 // 存储的文件名
	Ext          string      `json:"ext"`                                        // 后缀
	MimeType     string      `json:"mime_type" gorm:"index"`                     // 类型
	Title        string      `json:"title"`                                      // 名称
	StorageType  string      `json:"storage_type"`                               // 存储格式
	Size         int64       `json:"size" gorm:"default:0"`                      // 大小
	Addtime      time.Time   `json:"addtime"`                                    // 添加时间
	Duration     int64       `json:"duration" gorm:"index;default:0"`            // 时长,秒
	Bitrate      string      `json:"bitrate"`                                    // 比特率
	SampleRate   string      `json:"sample_rate"`                                // 采样率
	Channels     string      `json:"channels"`                                   // 声道
	Codec        string      `json:"codec"`                                      // 编码
	MID          int64       `json:"mid" gorm:"index;default:0"`                 // 来自哪个视频
	UserID       int64       `json:"user_id" gorm:"index;default:0"`             // 来自哪个用户
	Removed      int         `json:"-" gorm:"index;default:0"`                   // 软删除标记
	RemoveTime   int64       `json:"-" gorm:"index;default:0"`                   // 软删除时间
	Language     string      `json:"language" gorm:"index"`                      // 语音
	Tags         []*AudioTag `json:"tags" gorm:"many2many:audio_tag_relations;"` // 标签列表
	db.BaseModel `json:"-" gorm:"-"`
}

type AudioTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index"`
}

var Dbs = db.MEDIADB

func init() {
	Dbs.DB().AutoMigrate(&Audio{}, &AudioTag{})
}

func (m Audio) MarshalJSON() ([]byte, error) {
	type Alias Audio
	a := Alias(m)
	if a.Name != "" {
		if a.StorageType == "" {
			a.StorageType = config.DefStorage
		}
		a.Name = storage.Load(a.Name).URL(a.Name)
	}

	return config.Json.Marshal(a)
}

func detectMimeFromURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Range", "bytes=0-4095")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)

	return http.DetectContentType(buf[:n]), nil
}

// 添加音频到数据库
func AddAudio(src, title string) (*Audio, error) {
	src = storage.Load("").URL(src)

	var mimeType string
	if strings.HasPrefix(src, "http") {
		s, err := detectMimeFromURL(src)
		if err != nil {
			return nil, err
		}
		mimeType = s
	} else {
		s, err := funcs.FileMimeType(src)
		if err != nil {
			return nil, err
		}
		mimeType = s
	}

	if strings.Contains(mimeType, "video") { // 如果是视频,则分离音频
		audioSrc := config.FullPath(config.MEDIAROOT, filepath.Base(src)+".mp3")
		if _, err := os.Stat(audioSrc); err != nil {
			if err := ffmpeg.SeparateAudio(src, audioSrc); err != nil {
				return nil, err
			}
		}
		s, err := storage.Load("").PutStr(audioSrc)
		if err != nil {
			return nil, err
		}
		src = storage.Load("").URL(s)
	}

	info, err := ffmpeg.GetAudioInfo(src)
	if err != nil {
		return nil, err
	}

	name := storage.Load("").Base(src)
	audio := new(Audio)
	Dbs.DB().Model(&Audio{}).Where("name = ?", name).Find(audio)

	if audio.ID < 1 {
		audio = &Audio{
			Name:        name,
			Title:       title,
			Ext:         strings.Trim(filepath.Ext(src), ","),
			MimeType:    mime.TypeByExtension(filepath.Ext(src)),
			StorageType: config.DefStorage,
			Addtime:     time.Now(),
			Duration:    info.Duration,
			Bitrate:     info.Bitrate,
			SampleRate:  info.SampleRate,
			Channels:    info.Channels,
			Codec:       info.Codec,
			Size:        funcs.GetFileSize(src),
		}
	} else {
		audio.Removed = 0
	}

	return audio, nil
}
