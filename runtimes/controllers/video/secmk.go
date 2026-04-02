package video

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/db/donevideo"
	"tools/runtimes/db/medias"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/imager"
	"tools/runtimes/mainsignal"
	"tools/runtimes/response"
	"tools/runtimes/storage"
	"tools/runtimes/videoproc"

	"github.com/gin-gonic/gin"
)

// 二创的数据
func SecmkData(c *gin.Context) {
	// _, audioSlice := audios.GetList(&db.ListFinder{
	// 	Limit: 12,
	// 	Page:  1,
	// })
	_, audioTagSlice := audios.GetAudioTags(&db.ListFinder{
		Page:  1,
		Limit: 500,
	})

	response.Success(c, gin.H{
		// "audios":     audioSlice,
		"audio_tags": audioTagSlice,
		"user_tags":  medias.GetTags(),
	}, "")
}

// 按条件查询视频
// type smData struct {
// 	UserIDs          []int64 `json:"user_ids"`
// 	UserTags         []int64 `json:"user_tags"`
// 	SearchMediaTitle string  `json:"search_media_title"`
// 	Page             int     `json:"page"`
// 	Limit            int     `json:"limit"`
// }

func SearchMedias(c *gin.Context) {
	var req videos
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	mds, total, err := req.getVideos()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, gin.H{
		"total": total,
		"lists": mds,
	}, "")
}

// 视频制作
type factory struct {
	AmixAduio    []int64        `json:"amix_aduio"`
	AmixAduioTag []int64        `json:"amix_aduio_tag"`
	AmixVolume   float64        `json:"amix_volume"`
	Audio        []int64        `json:"audio"`
	AudioTag     []int64        `json:"audio_tag"`
	AudioVolume  float64        `json:"audio_volume"`
	Clearer      bool           `json:"clearer"` // ai清晰
	Crop         *imager.Crop   `json:"crop"`
	Cropable     bool           `json:"cropable"`
	Flip         int            `json:"flip"`
	Height       float64        `json:"height"`
	Linear       *imager.Linear `json:"linear"`
	Resize       float64        `json:"resize"`
	Rotation     float64        `json:"rotation"`
	Width        float64        `json:"width"`
}
type minmax struct {
	Min int `json:"min"`
	Max int `json:"max"`
}
type videos struct {
	IncludeChildPath bool    `json:"include_child_path"`
	Path             int64   `json:"path"`
	SearchMediaTitle string  `json:"search_media_title"`
	Size             *minmax `json:"size"`
	UserFans         *minmax `json:"user_fans"`
	UserIDs          []int64 `json:"user_ids"`
	UserTags         []int64 `json:"user_tags"`
	UserWorks        *minmax `json:"user_works"`
	Page             int     `json:"page"`
	Limit            int     `json:"limit"`
}
type mkvideoData struct {
	Factory *factory `json:"factory"`
	Videos  *videos  `json:"videos"`
}

type Maker struct {
	ctx     context.Context
	cancle  context.CancelFunc
	StartAt int64    `json:"start_at"`
	Errs    []string `json:"errs"`
}

var makers []*Maker

func (v *videos) getVideos() (mds []*medias.Media, total int64, err error) {
	if len(v.UserTags) > 0 {
		uids := medias.GetMUIDsByTagIDs(v.UserTags)
		v.UserIDs = append(v.UserIDs, uids...)
	}

	model := medias.GetDb().DB().Model(&medias.Media{}).
		Where("removed = 0").Where(`
    EXISTS (
        SELECT 1 FROM media_files
        WHERE media_files.m_id = media.id
        AND mime_type LIKE ?
    )
`, "video/%").Preload("Files", "mime_type LIKE ?", "video/%")

	if len(v.UserIDs) > 0 {
		model = model.Where("user_id in ?", v.UserIDs)
	}
	if v.SearchMediaTitle != "" {
		model = model.Where("title like ?", fmt.Sprintf("%%%s%%", v.SearchMediaTitle))
	}

	if v.Limit > 0 {
		if v.Page < 1 {
			v.Page = 1
		}
		model = model.Offset((v.Page - 1) * v.Limit).Limit(v.Limit).Order("id DESC")
	} else {
		if len(v.UserIDs) < 1 && v.SearchMediaTitle == "" {
			return nil, 0, errors.New("未设置条件")
		}
	}

	model.Count(&total)
	model.Find(&mds)

	return
}

func Stop(c *gin.Context) {
	idx, err := strconv.Atoi(c.Query("idx"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(makers) < idx+1 {
		response.Error(c, http.StatusBadRequest, i18n.T("找不到"), nil)
		return
	}

	makers[idx].cancle()
	response.Success(c, nil, "")
}

func MakerVideos(c *gin.Context) {
	var req mkvideoData
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	req.Videos.Limit = 0
	mds, total, err := req.Videos.getVideos()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if total < 1 {
		response.Error(c, http.StatusBadRequest, i18n.T("未设置视频"), nil)
		return
	}

	fac, ads, axds, err := req.Factory.build()
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	ctx, cancle := context.WithCancel(mainsignal.MainCtx)
	mk := &Maker{
		ctx:     ctx,
		cancle:  cancle,
		StartAt: time.Now().Unix(),
	}
	makers = append(makers, mk)

	go func() {
		for k, v := range mds {
			if len(v.Files) < 1 {
				// fmt.Println("找不到文件")
				continue
			}

			var videos []string
			for _, row := range v.Files {
				videos = append(videos, storage.Load("").URL(row.FileName))
			}

			mker, err := videoproc.SecMaker(videos, fac, ctx, func(idx, itotal int) {
				if itotal > 0 {
					inner := float64(idx) * 100 / float64(itotal)
					percent := (float64(k)*100 + inner) / float64(total)
					msg := fmt.Sprintf("%.2f%%", percent)
					// messages.SuccessMsg(msg)
					fmt.Printf("\r%d/%d: %s", idx, itotal, msg)
				}

			})
			if err != nil {
				mk.Errs = append(mk.Errs, err.Error())
				return
			}

			var ad *videoproc.AudioInpter
			if len(ads) > 0 {
				adidx := funcs.RandomNumber(0, len(ads)-1)
				// fmt.Println("音频下标：", adidx)
				ad = ads[adidx]
			}
			var axd *videoproc.AudioInpter
			if len(axds) > 0 {
				axd = axds[funcs.RandomNumber(0, len(axds)-1)]
			}
			mker.Audio = ad
			mker.AmixAudio = axd
			mker.Width = int(req.Factory.Width)
			mker.Height = int(req.Factory.Height)

			go func() {
				outfile := config.FullPath(config.MEDIAROOT, ".tmp", filepath.Base(videos[0]))
				err := mker.Output(outfile)
				if err != nil {
					// fmt.Println("生成失败:", err)
					mk.Errs = append(mk.Errs, err.Error())
				}
				fn, err := storage.Load("").PutStr(outfile)
				if err != nil {
					mk.Errs = append(mk.Errs, err.Error())
				}
				if md, err := videoproc.ProbeMedia(storage.Load("").URL(fn)); err == nil {
					donevideo.AddRow(fn, v.Cover, md, v.UserId)
				}
				v.SecMaker++
				v.Save()
				os.Remove(outfile)
			}()
		}
	}()

	response.Success(c, makers, "已加入处理队列")
}

func (f *factory) build() (*videoproc.Factory, []*videoproc.AudioInpter, []*videoproc.AudioInpter, error) {
	fac := new(videoproc.Factory)
	fac.Clearer = f.Clearer
	fac.Crop = f.Crop
	fac.Linear = f.Linear
	fac.Mirror = f.Flip
	fac.Resize = &imager.Resize{Scale: f.Resize}
	fac.Rotation = &imager.Rotation{Angle: f.Rotation}

	if f.AudioVolume <= 0 {
		f.AudioVolume = 1
	}
	if f.AmixVolume <= 0 {
		f.AmixVolume = 1
	}

	var ais []*videoproc.AudioInpter
	for _, v := range audios.GetByIDsOrTagIDs(f.Audio, f.AudioTag) {
		ais = append(ais, &videoproc.AudioInpter{
			Url:    storage.Load(v.StorageType).URL(v.Name),
			Volume: f.AudioVolume,
		})
	}
	// fmt.Println(len(ais), "音频数量----")

	var axis []*videoproc.AudioInpter
	for _, v := range audios.GetByIDsOrTagIDs(f.AmixAduio, f.AmixAduioTag) {
		axis = append(axis, &videoproc.AudioInpter{
			Url:    storage.Load(v.StorageType).URL(v.Name),
			Volume: f.AmixVolume,
		})
	}
	return fac, ais, axis, nil
}
