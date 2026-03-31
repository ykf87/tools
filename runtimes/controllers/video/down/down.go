package down

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// "sync"

	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/proxys"
	"tools/runtimes/storage"

	// "tools/runtimes/eventbus"

	"tools/runtimes/funcs"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	_          = iota
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

type downDt struct {
	Urls     string  `json:"urls" form:"urls"`
	AutoDown bool    `json:"auto_down" form:"auto_down"` // 是否自动下载
	Dest     string  `json:"dest" form:"dest"`           // 自动下载保存的目录
	Proxys   []int64 `json:"proxys" form:"proxys"`       // proxys表的id
	Path     string  `json:"path" form:"path"`           // 下载路径
}

type ListDataStruct struct {
	PathID int64  `json:"path_id"`
	Path   string `json:"path"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Tp     string `json:"tp"`
	Ext    string `json:"ext"`
	Mime   string `json:"mime"`
	Search string `json:"search"`
}

type Pms struct {
	Name       string            `json:"name"`         // 文件名称
	Path       string            `json:"path"`         // 路径名称
	FullName   string            `json:"full_name"`    // 运行目录下的相对路径
	Url        string            `json:"url"`          // 链接地址
	Timer      int64             `json:"timer"`        // 最后更新时间
	Dir        bool              `json:"dir"`          // 是否是目录
	Ext        string            `json:"ext"`          // 文件后缀
	Size       string            `json:"size"`         // 文件大小
	Mime       string            `json:"mime"`         // 文件类型
	Fmt        string            `json:"fmt"`          // 下载百分比字符串
	Num        float64           `json:"num"`          // 下载进度数字,100为下载完成
	DownFile   string            `json:"down_file"`    // md5内容,下载地址的md5
	Status     int               `json:"status"`       // 下载状态
	DownErrMsg string            `json:"down_err_msg"` // 下载错误信息
	Platform   string            `json:"platform"`     // 下载的平台
	Cover      string            `json:"cover"`        // 封面
	User       *medias.MediaUser `json:"user" gorm:"-"`
}
type ByTimerDesc []*Pms

func (a ByTimerDesc) Len() int           { return len(a) }
func (a ByTimerDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimerDesc) Less(i, j int) bool { return a[i].Timer > a[j].Timer }

func List(c *gin.Context) {
	ddt := new(ListDataStruct)
	if err := c.ShouldBindJSON(ddt); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	page := ddt.Page
	limit := ddt.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	var lists []*medias.MediaResponseMessage

	mps, total := medias.GetChilds(ddt.PathID, page, limit, ddt.Search)
	dirlen := len(mps)
	var flen int64

	if dirlen > 0 {
		for _, v := range mps {
			mmmmp := &medias.MediaResponseMessage{
				Type:    1,
				Title:   v.Name,
				ID:      v.ID,
				Addtime: v.Addtime,
				Sizes:   0,
			}

			var mpids []int64
			medias.GetDb().DB().Model(&medias.MediaPath{}).Select("id").
				Where("chain like ? AND chain <> ?", fmt.Sprintf("%s%%", v.Chain), v.Chain).
				Scan(&mpids)
			mpids = append(mpids, v.ID)

			medias.GetDb().DB().Model(&medias.Media{}).Select("SUM(sizes)").Where("removed = ? and path_id in ?", 0, mpids).Scan(&mmmmp.Sizes)

			lists = append(lists, mmmmp)
		}
	}

	if dirlen < limit {
		var mds []medias.Media
		mlimit := limit - dirlen
		model := medias.GetDb().DB().Model(&medias.Media{}).
			Where("removed = 0 and path_id = ?", ddt.PathID)
		if ddt.Search != "" {
			model = model.Where("title like ?", fmt.Sprintf("%%%s%%", ddt.Search))
		}

		model.Count(&flen)
		total = total + flen
		model.Order("media.id DESC").
			Preload("User").Preload("Files").
			Offset((page - 1) * mlimit).Limit(mlimit).
			Find(&mds)

		for _, v := range mds {
			lists = append(lists, &medias.MediaResponseMessage{
				Type:         0,
				Title:        v.Title,
				ID:           v.Id,
				Addtime:      v.Addtime,
				Cover:        v.Cover,
				CoverStorage: v.CoverStorage,
				Sizes:        v.Sizes,
				Numbers:      int64(v.Numbers),
				Files:        v.Files,
				User:         v.User,
				Platform:     v.Platform,
			})
		}
	}

	tf, _ := strconv.ParseFloat(fmt.Sprintf("%d", total), 64)
	lf, _ := strconv.ParseFloat(fmt.Sprintf("%d", limit), 64)
	rs := map[string]any{
		"pages":    int(math.Ceil(tf / lf)),
		"limit":    limit,
		"list":     lists,
		"total":    total,
		"dirs":     dirlen,
		"fils":     flen,
		"baseurl":  config.FullPath(config.MEDIAROOT),
		"prevpath": ddt.Path,
		"prevurl":  config.MediaUrl,
	}
	response.Success(c, rs, "")
}

func Download(c *gin.Context) {
	user, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	dt := new(downDt)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var proxysC []*proxy.ProxyConfig
	for _, proxyid := range dt.Proxys {
		px := proxys.GetById(proxyid)
		if pc, err := proxy.Client(px.GetConfig(), "", px.Port, px.GetTransfer()); err == nil {
			proxysC = append(proxysC, pc)
		}
	}
	go medias.GetPlatformVideos(dt.Urls, proxysC, dt.Path, user.Id, "", false, false)
	// var errmsg []string
	// for _, err := range errs {
	// 	errmsg = append(errmsg, err.Error())
	// }
	response.Success(c, nil, "")
}

func ReDownload(c *gin.Context) {
	user, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusNonAuthoritativeInfo, err.Error(), nil)
		return
	}

	type idsobj struct {
		Ids    []int64 `json:"ids"`
		Proxys []int64 `json:"proxys"`
	}
	var ids idsobj
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var pxobj []*proxys.Proxy
	db.DB.DB().Model(&proxys.Proxy{}).Where("id in ?", ids.Proxys).Find(&pxobj)
	var pcs []*proxy.ProxyConfig
	for _, v := range pxobj {
		if pc, err := v.Start(false); err == nil {
			pcs = append(pcs, pc)
		}
	}

	var mds []*medias.Media
	medias.GetDb().DB().Model(&medias.Media{}).Where("id in ?", ids.Ids).Find(&mds)

	var errs []string
	for _, v := range mds {
		go func() {
			rerrs := v.ReDown(pcs, user.Id)
			for _, ev := range rerrs {
				errs = append(errs, ev.Error())
			}
		}()
	}

	msg := "重新下载已提交"
	if len(errs) > 0 {
		msg = strings.Join(errs, "\n")
	}
	response.Success(c, nil, msg)
}

type mkdirStruct struct {
	Name string `json:"name" form:"name"` // 目录名称
	Path string `json:"path" form:"path"` // 下载路径
}

func Mkdir(c *gin.Context) {
	admin, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	dt := new(mkdirStruct)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_, err = medias.MKDBNameID(fmt.Sprintf("%s/%s", dt.Path, dt.Name), admin.Id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "Success")
}

func OpenDir(c *gin.Context) {
	dt := new(mkdirStruct)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	fullPath := filepath.Join(config.MEDIAROOT, dt.Path, dt.Name)
	fmt.Println(fullPath, "----")
	f, err := os.Stat(fullPath)
	if err != nil {
		response.Error(c, 404, "目录不存在", nil)
		return
	}
	if f.IsDir() == false {
		response.Error(c, 500, "不是有效的目录", nil)
		return
	}

	funcs.OpenDir(fullPath)
	response.Success(c, nil, "Success")
}

func Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusNotModified, err.Error(), nil)
		return
	}
	defer file.Close()
	storageType := c.PostForm("storage")
	if storageType == "" {
		storageType = config.DefStorage
	}

	contentType := header.Header.Get("Content-Type")

	// 1️⃣ 读取头部用于 MIME 判断
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)

	mimeType := http.DetectContentType(buffer[:n])

	fullReader := io.MultiReader(
		bytes.NewReader(buffer[:n]),
		file,
	)

	h := sha256.New()
	tr := io.TeeReader(fullReader, h)
	fm := &storage.FileMeta{
		Ext:         filepath.Ext(header.Filename),
		ObjectKey:   fmt.Sprintf("tmp/%s", uuid.New().String()),
		Reader:      tr,
		H:           h,
		Size:        header.Size,
		ContentType: contentType,
	}

	s, err := storage.Load(storageType).Put(file, fm)
	if err != nil {
		response.Error(c, http.StatusNotModified, err.Error(), nil)
		return
	}

	dir := c.PostForm("dir")
	mp, _ := medias.MKDBNameID(c.PostForm("dir"), 0)
	// if err != nil {
	// 	response.Error(c, http.StatusNotModified, err.Error(), nil)
	// 	return
	// }

	var pid int64
	if mp != nil && mp.ID > 0 {
		pid = mp.ID
	}
	nsps := strings.Split(filepath.Base(s), ".")
	hash := nsps[0]
	if err := medias.GetDb().Write(func(tx *gorm.DB) error {
		md := &medias.Media{
			Title:   header.Filename,
			PathID:  pid,
			Path:    dir,
			Sizes:   fm.Size,
			Numbers: 1,
			Addtime: time.Now(),
		}
		if err := tx.Create(md).Error; err != nil {
			return err
		}

		mf := &medias.MediaFile{
			MID:        md.Id,
			FileName:   s,
			FileSystem: storageType,
			Size:       fm.Size,
			MimeType:   mimeType,
			CreatedAt:  time.Now(),
			Ext:        filepath.Ext(header.Filename),
			Hash:       hash,
		}
		return tx.Create(mf).Error
	}); err != nil {
		response.Error(c, http.StatusNotModified, err.Error(), nil)
		return
	}
	response.Success(c, gin.H{
		"filename": header.Filename,
		"path":     filepath.Base(s),
		"size":     fm.Size,
	}, "上传成功")

	// 4. 返回结果
	// c.JSON(http.StatusOK, gin.H{
	// 	"code": 0,
	// 	"msg":  "上传成功",
	// 	"data": gin.H{
	// 		"filename": header.Filename,
	// 		"path":     filepath.Base(s),
	// 		"size":     fm.Size,
	// 	},
	// })
}
