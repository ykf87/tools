package browsers

import (
	"fmt"
	"net/http"
	"tools/runtimes/bs"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagList struct {
	Q     string `json:"q" form:"q"`
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
}

func BrowserTags(c *gin.Context) {
	dt := new(TagList)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model := db.DB.DB().Model(&browserdb.BrowserTag{})
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		model = model.Where("name LIKE ?", qs)
	}
	var total int64
	model.Count(&total)
	model = model.Order("id DESC")

	if dt.Limit != -1 {
		if dt.Page < 1 {
			dt.Page = 1
		}
		if dt.Limit < 1 {
			dt.Limit = 20
		}
		model = model.Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit)
	}

	var list []*browserdb.BrowserTag
	model.Find(&list)

	rrs, _ := parses.Marshal(list, c)
	response.Success(c, gin.H{"list": rrs, "total": total}, "")
}

type ListStruct struct {
	Page    int     `json:"page" form:"page"`
	Limit   int     `json:"limit" form:"limit"`
	Q       string  `json:"q" form:"q"`
	SortCol string  `json:"scol" form:"scol"`
	By      string  `json:"by" form:"by"`
	Col     string  `json:"col" form:"col"`
	Cval    string  `json:"cval" form:"cval"`
	Tags    []int64 `json:"tags" form:"tags"`
}

func List(c *gin.Context) {
	var l ListStruct
	if err := c.ShouldBindQuery(&l); err != nil {
		response.Error(c, http.StatusNotFound, i18n.T("Error"), nil)
		return
	}

	model := db.DB.DB().Model(&browserdb.Browser{})
	if l.Q != "" {
		qs := fmt.Sprintf("%%%s%%", l.Q)
		model = model.Where("name LIKE ? OR lang local = ? or lang = ?", qs, qs, l.Q)
	}

	if l.Col != "" && l.Cval != "" {
		model = model.Where(fmt.Sprintf("%s = ?", l.Col), l.Cval)
	}

	if len(l.Tags) > 0 {
		var tagIds []int64
		for _, v := range l.Tags {
			tagIds = append(tagIds, v)
		}
		if len(tagIds) > 0 {
			model = model.Joins("right join browser_to_tags on browser_id = id").Where("browser_to_tags.tag_id in ?", tagIds)
		}
	}

	var total int64
	model.Count(&total)

	sortCol := "id"
	sortBy := "DESC"

	if l.SortCol != "" {
		sortCol = l.SortCol
	}
	if l.By != "" {
		if l.By == "asc" {
			sortBy = "ASC"
		} else {
			sortBy = "DESC"
		}
	}
	if l.Page < 1 {
		l.Page = 1
	}
	if l.Limit < 1 {
		l.Limit = 10
	}

	var ps []*browserdb.Browser
	model.Order(fmt.Sprintf("%s %s", sortCol, sortBy)).Offset((l.Page - 1) * l.Limit).Limit(l.Limit).Debug().Find(&ps)
	for _, v := range ps {
		v.Opend = bs.BsManager.IsArride(v.Id)
	}

	// 处理代理标签
	if len(ps) > 0 {
		browserdb.SetBrowserTags(ps)
	}

	rs := gin.H{"list": ps, "total": total}
	rsp, _ := parses.Marshal(rs, c)
	response.Success(c, rsp, "")
}

// 编辑
type BrowserAddData struct {
	Id          int64    `json:"id" form:"id"`                    // 编辑的时候才有
	Width       int      `json:"width" form:"width"`              // 屏幕宽度
	Height      int      `json:"height" form:"height"`            // 屏幕高度
	Name        string   `json:"name" form:"name"`                // 名称
	Proxy       int64    `json:"proxy" form:"proxy"`              // 代理的id
	ProxyConfig string   `json:"proxy_config" form:"ProxyConfig"` // 具体的配置
	Tags        []string `json:"tags" form:"tags"`                // 标签
	Lang        string   `json:"lang" form:"lang"`                // 语言
	Timezone    string   `json:"timezone" form:"timezone"`        // 时区
}

func Editer(c *gin.Context) {
	id := c.Param("id")
	var l BrowserAddData
	if err := c.ShouldBind(&l); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	u, ok := c.Get("_user")
	if ok == false {
		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
		return
	}
	user, ok := u.(*admins.Admin)
	if ok != true {
		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
		return
	}

	if l.Name == "" {
		response.Error(c, http.StatusNotFound, i18n.T("Name can not be empty"), nil)
		return
	}

	var browserObj *browserdb.Browser
	if id != "" {
		bb, err := browserdb.GetBrowserById(id)
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		browserObj = bb
	} else {
		browserObj = new(browserdb.Browser)
	}

	changeed := false
	if browserObj.Name != l.Name {
		browserObj.Name = l.Name
		changeed = true
	}
	if browserObj.Width != l.Width {
		browserObj.Width = l.Width
		changeed = true
	}
	if browserObj.Height != l.Height {
		browserObj.Height = l.Height
		changeed = true
	}
	if browserObj.Proxy != l.Proxy {
		if l.Proxy > 0 {
			px := proxys.GetById(l.Proxy)
			if px == nil || px.Id < 1 {
				response.Error(c, http.StatusNotFound, i18n.T("Proxy id not found"), nil)
				return
			}
			browserObj.Proxy = l.Proxy
			browserObj.Local = px.Local
			browserObj.Lang = px.Lang
			browserObj.Timezone = px.Timezone
			browserObj.Ip = px.Ip
			browserObj.ProxyName = px.Name
		} else {
			browserObj.Proxy = 0
		}
		changeed = true
	}
	if browserObj.ProxyConfig != l.ProxyConfig {
		if browserObj.Proxy < 1 {
			loc, err := proxy.GetLocal(l.ProxyConfig)
			if err != nil {
				response.Error(c, http.StatusNotFound, err.Error(), nil)
				return
			}
			browserObj.Ip = loc.Ip
			browserObj.Timezone = loc.Timezone
			browserObj.Local = loc.Iso
			browserObj.Lang = loc.Lang
		}
		browserObj.ProxyConfig = l.ProxyConfig
		changeed = true
	}
	if browserObj.Lang != l.Lang {
		browserObj.Lang = l.Lang
		changeed = true
	}
	if browserObj.Timezone != l.Timezone {
		browserObj.Timezone = l.Timezone
		changeed = true
	}

	if len(l.Tags) > 0 {
		if err := browserObj.CoverTgs(l.Tags, nil); err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
	}
	if changeed {
		browserObj.AdminID = user.Id
		if err := db.DB.Write(func(tx *gorm.DB) error {
			return browserObj.Save(browserObj, tx)
		}); err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
	}
	response.Success(c, browserObj, "Success")
}

// 获取语言列表
func GetLangs(c *gin.Context) {
	response.Success(c, bs.LangMap, "Success")
}

// 获取语言列表
func GetTimezones(c *gin.Context) {
	response.Success(c, bs.Timezones, "Success")
}

// 启动浏览器
func Start(c *gin.Context) {
	id := c.Param("id")
	bs, err := browserdb.GetBrowserById(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	u, ok := c.Get("_user")
	if ok == false {
		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
		return
	}
	user, ok := u.(*admins.Admin)
	if ok != true {
		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
		return
	}

	if bs.AdminID != user.Id {
		response.Error(c, http.StatusNotFound, i18n.T("无法启动"), nil)
		return
	}
	if err := bs.Open(nil); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	// u, ok := c.Get("_user")
	// if ok == false {
	// 	response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
	// 	return
	// }
	// user, ok := u.(*admins.Admin)
	// if ok != true {
	// 	response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
	// 	return
	// }

	// bs.AdminID = user.Id
	response.Success(c, nil, "Success")
}

// 关闭浏览器
func Stop(c *gin.Context) {
	id := c.Param("id")
	bs, err := browserdb.GetBrowserById(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if err := bs.Close(false); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, nil, "Success")
}

// 删除浏览器
func Delete(c *gin.Context) {
	id := c.Param("id")
	bs, err := browserdb.GetBrowserById(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if err := bs.Delete(); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, nil, "Success")
}
