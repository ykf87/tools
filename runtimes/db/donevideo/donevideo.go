package donevideo

import (
	"tools/runtimes/db"
	"tools/runtimes/videoproc"
)

type DoneVideo struct {
	ID       int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID   int64   `json:"user_id" gorm:"index;default:0"`
	Filename string  `json:"filename" gorm:"not null"`
	Cover    string  `json:"cover"`
	Format   string  `json:"format"`
	Duration float64 `json:"duration"`
	Size     int64   `json:"size"`
	Bitrate  int64   `json:"bitrate"`
	FPS      float64 `json:"fps"`
	Codec    string  `json:"codec"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	db.BaseModel
}

var Dbs *db.SQLiteWriter = db.MEDIADB

func init() {
	Dbs.DB().AutoMigrate(&DoneVideo{})
}

func AddRow(src, cover string, mif *videoproc.MediaInfo, userID int64) (*DoneVideo, error) {
	dv := &DoneVideo{
		UserID:   userID,
		Filename: src,
		Cover:    cover,
		Format:   mif.Format,
		Duration: mif.Duration,
		Size:     mif.Size,
		Bitrate:  mif.Video.Bitrate,
		FPS:      mif.Video.FPS,
		Codec:    mif.Video.Codec,
		Width:    mif.Video.Width,
		Height:   mif.Video.Height,
	}

	if err := dv.Save(dv, Dbs.DB()); err != nil {
		return nil, err
	}
	return dv, nil
}
