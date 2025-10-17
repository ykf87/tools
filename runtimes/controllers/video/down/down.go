package down

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"tools/runtimes/config"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/proxy"
	"tools/runtimes/response"
	"tools/runtimes/videos/downloader/parser"

	"github.com/gin-gonic/gin"
)

type downDt struct {
	Urls     string `json:"urls" form:"urls"`
	AutoDown int    `json:"auto_down" form:"auto_down"` // 是否自动下载
	Dest     string `json:"dest" form:"dest"`           // 自动下载保存的目录
	Proxy    string `json:"proxy" form:"proxy"`         // 使用的代理
	ProxyId  int64  `json:"proxy_id" form:"proxy_id"`   // proxys表的id
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
	Name     string `json:"name"`      //文件名称
	Path     string `json:"path"`      //路径名称
	FullName string `json:"full_name"` //运行目录下的相对路径
	Timer    int64  `json:"timer"`     //最后更新时间
	Dir      bool   `json:"dir"`       //是否是目录
	Ext      string `json:"ext"`       //文件后缀
	Size     string `json:"size"`      //文件大小
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
				t.Size = fmt.Sprintf("%.2f M", float64(v.Size())/1048576.0)
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

func Download(c *gin.Context) {
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

	var transport *http.Transport
	var errs []string
	var downs []*parser.VideoParseInfo
	var wg sync.WaitGroup
	var proxyUrlStr string

	// 设置代理
	if dt.Proxy != "" {
		proxyUrlStr = dt.Proxy

	} else if dt.ProxyId > 0 {
		px := proxys.GetById(dt.ProxyId)
		if px != nil && px.Id > 0 {
			if pc, err := proxy.Client(px.GetConfig(), "", px.Port, px.GetTransfer()); err == nil {
				if _, err := pc.Run(false); err == nil {
					proxyUrlStr = pc.Listened()
					defer pc.Close(false)
				}
			}
		}
	}
	if proxyUrlStr != "" {
		if proxyURL, err := url.Parse(proxyUrlStr); err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		} // 你的代理地址
	}

	for _, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			parseRes, err := parser.ParseVideoShareUrlByRegexp(u, transport)
			if err != nil {
				errs = append(errs, err.Error()) //i18n.T("%s download failed", u)
				return
			}
			downs = append(downs, parseRes)
		}()

	}
	wg.Wait()
	response.Success(c, downs, strings.Join(errs, "\n"))
}
