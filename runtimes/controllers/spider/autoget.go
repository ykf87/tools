package spider

import (
	"fmt"
	"math/rand"
	"tools/runtimes/bs"
	"tools/runtimes/db"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/products"
	"tools/runtimes/db/proxys"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/chromedp/cdproto/runtime"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type ProxyC struct {
	IDs  []int64  `json:"ids" form:"ids"`
	Tags []string `json:"tags" form:"tags"`
}

type ReqObj struct {
	ProxyConfig *ProxyC `json:"proxy_config" form:"proxy_config"`
	SpiderHisID int64   `json:"spider_his_id" form:"spider_his_id"`
	SpiderJSID  int64   `json:"spider_js_id" form:"spider_js_id"`
	Show        bool    `json:"show" form:"show"`
}

func (pc *ProxyC) GetPC() []int64 {
	model := db.DB.DB().Model(&proxys.Proxy{}).Select("proxies.id")
	if len(pc.IDs) > 0 {
		model = model.Where("id in ?", pc.IDs)
	}
	if len(pc.Tags) > 0 {
		model = model.Joins("RIGHT JOIN proxy_tags as pt on pt.proxy_id = proxies.id").Where("pt.tag_id in (?)", db.DB.DB().Table("tags as t").Select("id").Where("name in ?", pc.Tags))
	}

	var proxyID []int64
	model.Scan(&proxyID)

	return proxyID
}

func AutoContent(c *gin.Context) {
	var req ReqObj
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}

	if req.SpiderJSID < 1 {
		response.Error(c, 500, "请选择js", nil)
		return
	}
	var jsstr string
	jsobj := jses.GetJsById(req.SpiderJSID)
	if jsobj != nil {
		jsstr = jsobj.GetContent(nil)
	}
	if jsstr == "" {
		response.Error(c, 500, "无法获取js内容", nil)
		return
	}

	proModel := products.DB.DB().Where(&products.Product{}).Where("spidered = 0")
	if req.SpiderHisID > 0 {
		proModel = proModel.Where("spider_his_id = ?", req.SpiderHisID)
	}

	proxyIDs := req.ProxyConfig.GetPC()
	var maxIdx int
	if len(proxyIDs) > 0 {
		maxIdx = len(proxyIDs) - 1
	}
	fmt.Println("----- 代理", proxyIDs)

	var pros []*products.Product
	proModel.Find(&pros)

	go func() {
		chs := make(chan bool, 1)
		for _, pro := range pros {
			if pro.Url == "" {
				continue
			}
			var pxyid int64
			if maxIdx > 0 {
				pxyid = proxyIDs[rand.Intn(maxIdx)]
			}

			if err := getFromBrowse(pxyid, "", pro, req.Show, chs); err != nil {
				response.Error(c, 500, err.Error(), nil)
				return
			}
			<-chs
		}
	}()
	response.Success(c, nil, "请求已提交")
}

func getFromBrowse(proxyID int64, jsstr string, pro *products.Product, isshow bool, ch chan bool) error {
	var pc *proxy.ProxyConfig

	if proxyID > 0 {
		if pcc, err := proxys.GetProxyConfigByID(proxyID); err == nil {
			fmt.Println(pcc.Name)
			pc = pcc
		}
	}

	bs, err := bs.BsManager.New(0, &bs.Options{
		Pc:    pc,
		Url:   pro.Url,
		Show:  isshow,
		JsStr: jsstr,
	}, false)
	if err != nil {
		return err
	}

	bs.OpenBrowser()
	bs.OnClosed(func() {
		if pc != nil {
			pc.Close(false)
		}
		ch <- true
	})
	bs.OnConsole(func(args []*runtime.RemoteObject) {
		for _, arg := range args {
			if arg.Value != nil {
				gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
				fmt.Println(gs.String(), "-----")
				if gs.Get("type").String() == "tools" {
					fmt.Println(gs.Get("data").String(), "---------")
				}
			}
		}
	})
	// bs.RunJs(jsstr)
	return nil
}
