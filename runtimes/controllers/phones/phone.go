package phones

import (
	"fmt"
	"net/http"
	"tools/runtimes/browser"
	"tools/runtimes/db"
	"tools/runtimes/db/clients"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type TagList struct {
	Q     string `json:"q" form:"q"`
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
}

func PhoneTags(c *gin.Context) {
	dt := new(TagList)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model := db.DB.Model(&clients.PhoneTag{})
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		model = model.Where("name LIKE ?", qs)
	}
	var total int64
	model.Count(&total)
	model = model.Order("id DESC")

	if dt.Limit > 0 {
		if dt.Page < 1 {
			dt.Page = 1
		}
		model = model.Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit)
	}

	var list []*clients.PhoneTag
	model.Find(&list)

	rrs, _ := parses.Marshal(list, c)
	response.Success(c, gin.H{"list": rrs, "total": total}, "")
}

type ListStruct struct {
	Page    int      `json:"page" form:"page"`
	Limit   int      `json:"limit" form:"limit"`
	Q       string   `json:"q" form:"q"`
	SortCol string   `json:"scol" form:"scol"`
	By      string   `json:"by" form:"by"`
	Col     string   `json:"col" form:"col"`
	Cval    string   `json:"cval" form:"cval"`
	Tags    []string `json:"tags" form:"tags"`
}

// 获取总数
func Total(c *gin.Context) {
	response.Success(c, clients.PhoneTotal(), "success")
}

func List(c *gin.Context) {
	var l ListStruct
	if err := c.ShouldBindJSON(&l); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	fmt.Println(l.Tags, l.Limit, "----", l)

	model := db.DB.Model(&clients.Phone{})
	if l.Q != "" {
		qs := fmt.Sprintf("%%%s%%", l.Q)
		model = model.Where("name LIKE ? or lang like ? OR local like ? or num = ? or os like ? or brand like ?", qs, qs, qs, l.Q, qs, qs)
	}

	if l.Col != "" && l.Cval != "" {
		model = model.Where(fmt.Sprintf("%s = ?", l.Col), l.Cval)
	}

	if len(l.Tags) > 0 {
		mtgs := clients.GetPhoneTagsByNames(l.Tags, nil)
		var tagIds []int64
		for _, v := range mtgs {
			tagIds = append(tagIds, v)
		}
		if len(tagIds) > 0 {
			model = model.Joins("right join phone_to_tags on phone_id = id").Where("phone_to_tags.tag_id in ?", tagIds)
		}
	}

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

	var ps []*clients.Phone
	model.Order(fmt.Sprintf("%s %s", sortCol, sortBy)).Offset((l.Page - 1) * l.Limit).Limit(l.Limit).Debug().Find(&ps)

	// 处理代理标签
	if len(ps) > 0 {
		clients.SetPhoneTags(ps)
	}

	rsp, _ := parses.Marshal(ps, c)
	response.Success(c, rsp, "")
}

// 编辑
type PhoneAddData struct {
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
	var l PhoneAddData
	if err := c.ShouldBind(&l); err != nil {
		response.Error(c, http.StatusNotFound, i18n.T("Error"), nil)
		return
	}

	if l.Name == "" {
		response.Error(c, http.StatusNotFound, i18n.T("Name can not be empty"), nil)
		return
	}

	var phoneObj *clients.Phone
	if id != "" {
		bb, err := clients.GetPhoneById(id)
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		phoneObj = bb
	} else {
		phoneObj = new(clients.Phone)
	}

	changeed := false
	if phoneObj.Name != l.Name {
		phoneObj.Name = l.Name
		changeed = true
	}
	// if phoneObj.Width != l.Width {
	// 	phoneObj.Width = l.Width
	// 	changeed = true
	// }
	// if phoneObj.Height != l.Height {
	// 	phoneObj.Height = l.Height
	// 	changeed = true
	// }
	if phoneObj.Proxy != l.Proxy {
		if l.Proxy > 0 {
			px := proxys.GetById(l.Proxy)
			if px == nil || px.Id < 1 {
				response.Error(c, http.StatusNotFound, i18n.T("Proxy id not found"), nil)
				return
			}
			phoneObj.Proxy = l.Proxy
			phoneObj.Local = px.Local
			phoneObj.Lang = px.Lang
			phoneObj.Timezone = px.Timezone
			phoneObj.Ip = px.Ip
			phoneObj.ProxyName = px.Name
		} else {
			phoneObj.Proxy = 0
		}
		changeed = true
	}
	if phoneObj.ProxyConfig != l.ProxyConfig {
		if phoneObj.Proxy < 1 {
			loc, err := proxy.GetLocal(l.ProxyConfig)
			if err != nil {
				response.Error(c, http.StatusNotFound, err.Error(), nil)
				return
			}
			phoneObj.Ip = loc.Ip
			phoneObj.Timezone = loc.Timezone
			phoneObj.Local = loc.Iso
			phoneObj.Lang = loc.Lang
		}
		phoneObj.ProxyConfig = l.ProxyConfig
		changeed = true
	}
	if phoneObj.Lang != l.Lang {
		phoneObj.Lang = l.Lang
		changeed = true
	}
	if phoneObj.Timezone != l.Timezone {
		phoneObj.Timezone = l.Timezone
		changeed = true
	}

	if len(l.Tags) > 0 {
		if err := phoneObj.CoverTgs(l.Tags, nil); err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
	}
	if changeed {
		if err := phoneObj.Save(nil); err != nil {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
	}
	response.Success(c, phoneObj, "Success")
}

// 获取语言列表
func GetLangs(c *gin.Context) {
	response.Success(c, browser.LangMap, "Success")
}

// 获取语言列表
func GetTimezones(c *gin.Context) {
	response.Success(c, browser.Timezones, "Success")
}

// 启动浏览器
// func Start(c *gin.Context) {
// 	id := c.Param("id")
// 	bs, err := clients.GetPhoneById(id)
// 	if err != nil {
// 		response.Error(c, http.StatusNotFound, err.Error(), nil)
// 		return
// 	}
// 	if err := bs.Open(); err != nil {
// 		response.Error(c, http.StatusNotFound, err.Error(), nil)
// 		return
// 	}

// 	u, ok := c.Get("_user")
// 	if ok == false {
// 		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
// 		return
// 	}
// 	user, ok := u.(*admins.Admin)
// 	if ok != true {
// 		response.Error(c, http.StatusNotFound, i18n.T("Please login first"), nil)
// 		return
// 	}

// 	bs.Bs.UserId = user.Id
// 	response.Success(c, nil, "Success")
// }

// // 关闭浏览器
// func Stop(c *gin.Context) {
// 	id := c.Param("id")
// 	bs, err := clients.GetBrowserById(id)
// 	if err != nil {
// 		response.Error(c, http.StatusNotFound, err.Error(), nil)
// 		return
// 	}

// 	if err := bs.Close(); err != nil {
// 		response.Error(c, http.StatusNotFound, err.Error(), nil)
// 		return
// 	}
// 	response.Success(c, nil, "Success")
// }

// 删除手机设备
func Delete(c *gin.Context) {
	id := c.Param("id")
	bs, err := clients.GetPhoneById(id)
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
