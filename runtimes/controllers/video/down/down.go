package down

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	// "sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/proxys"
	"tools/runtimes/downloader"

	// "tools/runtimes/eventbus"
	"tools/runtimes/file"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/listens/ws"
	"tools/runtimes/proxy"
	"tools/runtimes/response"
	"tools/runtimes/videos/downloader/parser"

	"github.com/gin-gonic/gin"
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
	Path  string `json:"path"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Tp    string `json:"tp"`
	Ext   string `json:"ext"`
	Mime  string `json:"mime"`
}

type Pms struct {
	Name       string  `json:"name"`         // 文件名称
	Path       string  `json:"path"`         // 路径名称
	FullName   string  `json:"full_name"`    // 运行目录下的相对路径
	Url        string  `json:"url"`          // 链接地址
	Timer      int64   `json:"timer"`        // 最后更新时间
	Dir        bool    `json:"dir"`          // 是否是目录
	Ext        string  `json:"ext"`          // 文件后缀
	Size       string  `json:"size"`         // 文件大小
	Mime       string  `json:"mime"`         // 文件类型
	Fmt        string  `json:"fmt"`          // 下载百分比字符串
	Num        float64 `json:"num"`          // 下载进度数字,100为下载完成
	DownFile   string  `json:"down_file"`    // md5内容,下载地址的md5
	Status     int     `json:"status"`       // 下载状态
	DownErrMsg string  `json:"down_err_msg"` // 下载错误信息
	Platform   string  `json:"platform"`     // 下载的平台
	Cover      string  `json:"cover"`        // 封面
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

	findPath := filepath.Join(config.MEDIAROOT, ddt.Path)
	fn, err := os.Stat(findPath)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	if fn.IsDir() == false {
		response.Error(c, http.StatusNotFound, i18n.T("%s 不是有效目录", ddt.Path), nil)
		return
	}

	fls, err := ioutil.ReadDir(findPath)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if ddt.Page < 1 {
		ddt.Page = 1
	}
	if ddt.Limit < 1 {
		ddt.Limit = 10
	}
	limits := ddt.Limit

	var dirs []*Pms
	var files []*Pms
	for _, v := range fls {
		nm := v.Name()
		if nm == "." || nm == ".." {
			continue
		}
		if nm[0:1] == "." {
			continue
		}

		t := new(Pms)
		t.Dir = v.IsDir()
		if t.Dir {
			t.Path = filepath.Join(ddt.Path, nm)
		} else {
			t.Path = ddt.Path
		}
		t.Name = nm

		t.Timer = v.ModTime().Unix()
		t.FullName = fmt.Sprintf("%s/%s", ddt.Path, nm)

		tms := strings.Split(nm, ".")
		if len(tms) > 1 {
			t.Ext = strings.ToLower(tms[len(tms)-1])
		}
		t.Mime = getMime(filepath.Join(config.MEDIAROOT, t.FullName))
		// fmt.Println(config.MediaUrl, "------------")
		t.Url = fmt.Sprintf("%s%s", config.MediaUrl, t.FullName)

		if ddt.Ext != "" && ddt.Ext != t.Ext {
			continue
		}
		if t.Ext == "yaml" {
			continue
		}
		if v.IsDir() {
			if ddt.Tp != "file" {
				dirs = append(dirs, t)
			}
		} else {
			if ddt.Mime != "" {
				if !checkMime(ddt.Mime, filepath.Join(findPath, nm)) {
					continue
				}
			}
			if ddt.Tp != "dir" {
				t.Size = funcs.FormatFileSize(v.Size())
				// t.Size = fmt.Sprintf("%.2f M", float64(v.Size())/1048576.0)
				files = append(files, t)
			}
		}
	}
	start := (ddt.Page - 1) * ddt.Limit

	var lists []*Pms
	sort.Sort(ByTimerDesc(dirs))
	for k, v := range dirs {
		if k >= start {
			lists = append(lists, v)
			ddt.Limit--
			if ddt.Limit <= 0 {
				break
			}
		}
	}

	dirlen := len(dirs)
	sort.Sort(ByTimerDesc(files))
	for k, v := range files {
		if dirlen+k >= start {
			lists = append(lists, v)
			ddt.Limit--
			if ddt.Limit <= 0 {
				break
			}
		}
	}

	flen := len(files)
	total := flen + dirlen
	tf, _ := strconv.ParseFloat(fmt.Sprintf("%d", total), 64)
	lf, _ := strconv.ParseFloat(fmt.Sprintf("%d", ddt.Limit), 64)
	pages := int(math.Ceil(tf / lf))

	rp := map[string]any{"pages": pages, "limit": limits, "list": lists, "total": total, "dirs": dirlen, "fils": flen, "baseurl": config.FullPath(config.MEDIAROOT), "prevpath": ddt.Path}
	response.Success(c, rp, "Success")
}

func getMime(fn string) string {
	f, _ := os.Open(fn)
	defer f.Close()

	// 读取文件前几字节用于 MIME 类型检测
	buf := make([]byte, 512)
	_, err := f.Read(buf)
	if err != nil {
		return ""
	}

	return strings.ToLower(http.DetectContentType(buf))
}

// 检查mime
func checkMime(mime, fn string) bool {
	if mime == "" {
		return true
	}
	fileinfo, err := os.Stat(fn)
	if err != nil {
		return true
	}
	if fileinfo.IsDir() == true {
		return true
	}
	mime = strings.ToLower(strings.ReplaceAll(mime, "*", ""))

	f, _ := os.Open(fn)
	defer f.Close()

	// 读取文件前几字节用于 MIME 类型检测
	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil {
		return true
	}

	// 检测 MIME 类型
	mimeType := strings.ToLower(http.DetectContentType(buf))
	if strings.HasPrefix(mimeType, mime) {
		return true
	}
	return false
}

// type downWsBack struct {
// 	Fmt    string  `json:"fmt"`
// 	Num    float64 `json:"num"`
// 	File   string  `json:"file"`
// 	Status int     `json:"status"`
// }

func Download(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusBadRequest, "Login first", nil)
		return
	}

	user := u.(*admins.Admin)

	dt := new(downDt)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	dt.Urls = strings.ReplaceAll(dt.Urls, "\r\n", "\n")
	dt.Urls = strings.ReplaceAll(dt.Urls, "\r", "\n")
	urls := strings.Split(dt.Urls, "\n")

	if len(urls) < 1 {
		response.Error(c, http.StatusBadRequest, i18n.T("Please enter the download address"), nil)
		return
	}

	// var transport *http.Transport
	var errs []string
	// var downs []*parser.VideoParseInfo
	var wg sync.WaitGroup
	// var proxyUrlStr string

	// 设置代理
	// if len(dt.Proxys) > 0 {
	// 	rand.Seed(time.Now().UnixNano())
	// 	proxyid := dt.Proxys[rand.Intn(len(dt.Proxys))]
	// 	px := proxys.GetById(proxyid)
	// 	if px != nil && px.Id > 0 {
	// 		if pc, err := proxy.Client(px.GetConfig(), "", px.Port, px.GetTransfer()); err == nil {
	// 			if _, err := pc.Run(false); err == nil {
	// 				proxyUrlStr = pc.Listened()
	// 				defer pc.Close(false)
	// 			}
	// 		}
	// 	}
	// }
	// if proxyUrlStr != "" {
	// 	if proxyURL, err := url.Parse(proxyUrlStr); err == nil {
	// 		transport = &http.Transport{
	// 			Proxy: http.ProxyURL(proxyURL),
	// 		}
	// 	} // 你的代理地址
	// }

	// 开启选择的代理
	rand.Seed(time.Now().UnixNano())
	proxyObjs := make(map[int64]*proxy.ProxyConfig)
	if len(dt.Proxys) > 0 {
		proxyid := dt.Proxys[rand.Intn(len(dt.Proxys))]
		if _, ok := proxyObjs[proxyid]; !ok {
			px := proxys.GetById(proxyid)
			if px != nil && px.Id > 0 {
				if pc, err := proxy.Client(px.GetConfig(), "", px.Port, px.GetTransfer()); err == nil {
					if _, err := pc.Run(false); err == nil {
						proxyObjs[px.Id] = pc
						defer pc.Close(false)
					}
				}
			}
		}
	}

	// rps := make(map[string]any)
	var rps []*Pms
	re := regexp.MustCompile(`https?://[^\s]+`)
	for _, u := range urls {
		wg.Add(1)
		ul := re.FindString(u)
		urlmd5 := funcs.Md5String(ul)

		rsrow := new(Pms)
		rsrow.DownFile = urlmd5

		var proxyUrl string
		if len(dt.Proxys) > 0 {
			proxyid := dt.Proxys[rand.Intn(len(dt.Proxys))]
			if pc, ok := proxyObjs[proxyid]; ok {
				proxyUrl = pc.Listened()
			}
		}

		go func(proxy, urlmd5 string) {
			defer wg.Done()
			var transport *http.Transport
			if proxy != "" {
				if proxyURL, err := url.Parse(proxy); err == nil {
					transport = &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					}
				}
			}
			parseRes, err := parser.ParseVideoShareUrlByRegexp(u, transport)
			if err != nil {
				errs = append(errs, err.Error()) //i18n.T("%s download failed", u)
				return
			}
			rsrow.Cover = parseRes.CoverUrl

			if dt.AutoDown == true {
				go requestDown(proxy, parseRes, urlmd5, dt.Path, user.Id)
			}

			// downs = append(downs, parseRes)
		}(proxyUrl, urlmd5)
		rps = append(rps, rsrow)
	}
	wg.Wait()
	response.Success(c, rps, strings.Join(errs, "\n"))
}

func requestDown(proxy string, parseRes *parser.VideoParseInfo, urlmd5, path string, uid int64) {
	if parseRes.VideoUrl != "" {
		var fullFn string
		fn := urlmd5
		d := downloader.NewDownloader(proxy, func(percent float64) {
			fmt.Printf("\r下载进度: %.2f%%", percent)
			dbk := new(Pms)
			dbk.DownFile = urlmd5
			dbk.Fmt = fmt.Sprintf("%.2f%%", percent)
			dbk.Num = percent
			dbk.Dir = false

			ws.SentBus(uid, "video-download", dbk, "")
		})

		ext, err := d.GetUrlFileExt(parseRes.VideoUrl)
		if err != nil {
			dbk := new(Pms)
			dbk.DownFile = urlmd5
			dbk.Fmt = ""
			dbk.Num = 0
			dbk.Dir = false
			dbk.Status = -1
			dbk.DownErrMsg = err.Error()

			ws.SentBus(uid, "video-download", dbk, "")
			return
		}
		fullFn = fmt.Sprintf("%s.%s", fn, ext)

		saveto := filepath.Join(path, fullFn)
		fullSaveTo := filepath.Join(config.MEDIAROOT, saveto)

		err = d.Download(parseRes.VideoUrl, fullSaveTo)
		if err != nil {
			dbk := new(Pms)
			dbk.DownFile = urlmd5
			dbk.Fmt = ""
			dbk.Num = 0
			dbk.Dir = false
			dbk.Status = -1
			dbk.DownErrMsg = err.Error()

			ws.SentBus(uid, "video-download", dbk, "")
			return
		}

		md := new(medias.Media)
		fl, err := file.NewFileInfo(fullSaveTo)
		if err != nil {
			return
		}
		md.Mime = fl.GetMime()
		md.Size = fl.Size()
		md.Filetime = fl.Time().Unix()
		md.Md5 = fl.Md5()
		md.Addtime = time.Now()
		md.Name = fullFn
		md.Path = path
		md.Platform = parseRes.Platform
		md.Url = parseRes.VideoUrl
		md.Title = parseRes.Title
		md.Save(nil)

		dbk := new(Pms)
		dbk.DownFile = urlmd5
		dbk.Fmt = "100%"
		dbk.Num = 100
		dbk.Dir = false
		dbk.Status = 1
		dbk.Mime = md.Mime
		dbk.Size = funcs.FormatFileSize(md.Size)
		dbk.Name = fullFn
		dbk.Platform = md.Platform

		tms := strings.Split(fullFn, ".")
		if len(tms) > 1 {
			dbk.Ext = strings.ToLower(tms[len(tms)-1])
		}

		dbk.FullName = fmt.Sprintf("%s/%s", md.Path, md.Name)
		dbk.Timer = md.Filetime

		ws.SentBus(uid, "video-download", dbk, "")
	} else if len(parseRes.Images) > 0 { // 下载图片

	}
}
