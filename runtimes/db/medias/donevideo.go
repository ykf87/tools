package medias

import (
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/storage"
	"tools/runtimes/videoproc"
)

type DoneVideo struct {
	ID       int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID   int64           `json:"user_id" gorm:"index;default:0"`
	Origin   string          `json:"origin" gorm:"default:null"`
	Filename string          `json:"filename" gorm:"not null"`
	Title    string          `json:"title" gorm:"index;"`
	Cover    string          `json:"cover"`
	Format   string          `json:"format"`
	Duration float64         `json:"duration"`
	Size     int64           `json:"size"`
	Bitrate  int64           `json:"bitrate"`
	FPS      float64         `json:"fps"`
	Codec    string          `json:"codec"`
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Tags     []*DoneVideoTag `json:"tags" gorm:"many2many:done_video_tag_relations;"`      // 标签列表
	UserTags []*MediaUserTag `json:"user_tags" gorm:"many2many:media_user_tag_relations;"` // 标签列表
	Addtime  int64           `json:"addtime" gorm:"default:0;index"`                       // 添加时间
	Removed  int             `json:"removed" gorm:"type:tinyint(1);default:0;index"`       //软删除
	db.BaseModel
}

type DoneVideoTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

func (m DoneVideo) MarshalJSON() ([]byte, error) {
	type Alias DoneVideo
	a := Alias(m)
	if a.Cover != "" {
		a.Cover = storage.Load("").URL(a.Cover)
	}
	if a.Filename != "" {
		a.Filename = storage.Load("").URL(a.Filename)
	}
	if a.Origin != "" {
		a.Origin = storage.Load("").URL(a.Origin)
	}

	return config.Json.Marshal(a)
}

func AddRow(src, origin, cover, title string, mif *videoproc.MediaInfo, userID int64) (*DoneVideo, error) {
	dv := &DoneVideo{
		UserID:   userID,
		Filename: src,
		Origin:   origin,
		Cover:    cover,
		Title:    title,
		Format:   mif.Format,
		Duration: mif.Duration,
		Size:     mif.Size,
		Bitrate:  mif.Video.Bitrate,
		FPS:      mif.Video.FPS,
		Codec:    mif.Video.Codec,
		Width:    mif.Video.Width,
		Height:   mif.Video.Height,
		Addtime:  time.Now().Unix(),
	}

	if err := dv.Save(dv, dbs.DB()); err != nil {
		return nil, err
	}
	return dv, nil
}
