package down

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
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
