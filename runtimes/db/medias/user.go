package medias

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/logs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/proxy"
	"tools/runtimes/runner"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type MediaUser struct {
	Id           int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string          `json:"name" gorm:"index;default:null"`                       // 用户名
	Cover        string          `json:"cover" gorm:"default:null"`                            // 头像
	Platform     string          `json:"platform" gorm:"index:plu;not null"`                   // 怕太
	Uuid         string          `json:"uuid" gorm:"index:plu;not null"`                       // 访问主页等
	Account      string          `json:"account" gorm:"index;"`                                // 例如抖音号,用于用户搜索的
	AdminID      int64           `json:"admin_id" gorm:"index;default:0"`                      // 哪个后台用户添加的
	Addtime      int64           `json:"addtime" gorm:"default:0;index"`                       // 添加时间
	Works        int64           `json:"works" gorm:"index;default:-1"`                        // 发布作品数量
	Fans         int64           `json:"fans" gorm:"index;default:-1"`                         // 粉丝数
	Local        string          `json:"local" gorm:"index;default:null"`                      // 所在地区
	Autoinfo     int             `json:"autoinfo" gorm:"default:0;index;type:tinyint(1)"`      // 自动获取信息,也就是监听账号的数据变化
	AutoTimer    int64           `json:"auto_timer" gorm:"index;default:0"`                    // 自动获取信息执行时间，每天
	AutoDownload int             `json:"auto_download" gorm:"index;default:0;type:tinyint(1)"` // 自动下载视频
	DownFreq     int             `json:"down_freq" gorm:"default:30"`                          // 自动下载视频执行频率,单位是分钟
	LastDownTime int64           `json:"last_down_time" gorm:"default:0"`                      // 上一次下载的时间
	Tags         []string        `json:"tags" gorm:"-"`                                        // 标签
	Clients      map[int][]int64 `json:"clients" gorm:"-"`                                     // 使用的客户端
	Proxys       []int64         `json:"proxys" gorm:"-"`                                      // 使用的代理列表
	Isruner      bool            `json:"isruner" gorm:"-"`                                     // 是否正在执行
	trans        string          `json:"-" gorm:"-"`                                           // 代理
}

type MediaUserToTag struct {
	UserID int64 `json:"user_id" gorm:"primaryKey;not null"`
	TagID  int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

// 媒体用户自动获取时使用的浏览器
type MediaUserToClient struct {
	MUID       int64 `json:"media_user" gorm:"primaryKey;not null"`
	ClientType int   `json:"client_type" gorm:"primaryKey;not null"`
	ClientID   int64 `json:"client_id" gorm:"primaryKey;not null"`
}

// 媒体用户自动获取时使用的代理,设置多个代理将随机选取代理
type MediaUserProxy struct {
	MUID    int64 `json:"mu_id" gorm:"primaryKey;not null"`
	ProxyID int64 `json:"proxy_id" gorm:"primaryKey;not null"`
}

// 用户每天的粉丝作品等数据
type MediaUserDay struct {
	MUID  int64  `json:"m_uid" gorm:"primaryKey;not null;"` // media_users表的id
	Ymd   string `json:"ymd" gorm:"primaryKey;not null"`    // 年月日的时间格式
	Works int64  `json:"works" gorm:"index;default:-1"`     // 发布作品数量
	Fans  int64  `json:"fans" gorm:"index;default:-1"`      // 粉丝数
	Zan   int64  `json:"zan" gorm:"index;default:-1"`       // 获赞数
}

func (this *MediaUser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}

	if strings.HasPrefix(this.Cover, config.MediaUrl) {
		if af, ok := strings.CutPrefix(this.Cover, config.MediaUrl); ok {
			this.Cover = af
		}
	}
	if this.Id > 0 {
		err := tx.Model(&MediaUser{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"platform":       this.Platform,
				"name":           this.Name,
				"cover":          this.Cover,
				"uuid":           this.Uuid,
				"fans":           this.Fans,
				"works":          this.Works,
				"local":          this.Local,
				"account":        this.Account,
				"autoinfo":       this.Autoinfo,
				"auto_download":  this.AutoDownload,
				"auto_timer":     this.AutoTimer,
				"down_freq":      this.DownFreq,
				"last_down_time": this.LastDownTime,
			}).Error
		if err != nil {
			return err
		}
		// eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

func (this *MediaUser) GetTags() []*MediaUserTag {
	var tagIDs []int64
	dbs.Model(&MediaUserToTag{}).Select("tag_id").Where("user_id = ?", this.Id).Find(&tagIDs)

	var tags []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Where("id in ?", tagIDs).Find(&tags)
	return tags
}

func (this *MediaUser) GetClients() map[int][]int64 {
	var rows []*MediaUserToClient
	dbs.Model(&MediaUserToClient{}).Where("m_uid = ?", this.Id).Find(&rows)
	this.Clients = make(map[int][]int64)
	for _, v := range rows {
		this.Clients[v.ClientType] = append(this.Clients[v.ClientType], v.ClientID)
	}
	return this.Clients
}

// 补全媒体用户的tag和客户端
func (this *MediaUser) Commpare() {
	this.GetClients()
	this.GetProxys()
	this.GenAvatarToHttp()

	this.Tags = nil
	for _, zv := range this.GetTags() {
		this.Tags = append(this.Tags, zv.Name)
	}
}

func (this *MediaUser) GenAvatarToHttp() {
	if !strings.HasPrefix(this.Cover, "http") {
		this.Cover = fmt.Sprintf("%s/%s", config.MediaUrl, this.Cover)
	}
}

// 获取代理列表
func (this *MediaUser) GetProxys() []int64 {
	var ids []int64
	dbs.Model(&MediaUserProxy{}).Select("proxy_id").Where("m_uid = ?", this.Id).Find(&ids)
	this.Proxys = ids
	return this.Proxys
}

func GetUserPlatforms() map[string]string {
	var pls []string
	dbs.Model(&MediaUser{}).Select("platform").Group("platform").Find(&pls)

	plsmap := make(map[string]string)
	for _, v := range pls {
		plsmap[v] = v
	}
	return plsmap
}

func GetMediaUsers(adminID int64, dt *db.ListFinder) ([]*MediaUser, int64) {
	var mus []*MediaUser
	if dt.Page < 1 {
		dt.Page = 1
	}
	if dt.Limit < 1 {
		dt.Limit = 20
	}
	md := dbs.Model(&MediaUser{}).Where("admin_id = ?", adminID)
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		md.Where("name like ?", qs)
	}

	if len(dt.Tags) > 0 {
		var muids []int64
		dbs.Model(&MediaUserToTag{}).Select("user_id").Where("tag_id in ?", dt.Tags).Find(&muids)
		if len(muids) > 0 {
			md.Where("id in ?", muids)
		}
	}

	var total int64
	md.Count(&total)

	if dt.Scol != "" && dt.By != "" {
		var byy string
		if strings.Contains(dt.By, "desc") {
			byy = "desc"
		} else {
			byy = "asc"
		}
		md.Order(fmt.Sprintf("%s %s", dt.Scol, byy))
	}
	md.Order("id DESC").Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit).Find(&mus)

	for _, v := range mus {
		v.Commpare()
		// if v.Cover != "" {
		// 	if !strings.HasPrefix(v.Cover, "http") {
		// 		v.Cover = fmt.Sprintf("%s/%s", config.MediaUrl, v.Cover)
		// 	}
		// }
	}
	return mus, total
}

// 根据id获取用户信息
func GetMediaUserByID(id any) *MediaUser {
	mu := new(MediaUser)
	if err := dbs.Model(&MediaUser{}).Where("id = ?", id).First(mu).Error; err != nil || mu.Id < 1 {
		return nil
	}
	// mu.Commpare()
	return mu
}

func (this *MediaUser) EmptyClient(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("m_uid = ?", this.Id).Delete(&MediaUserToClient{}).Error
}
func (this *MediaUser) EmptyProxy(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("m_uid = ?", this.Id).Delete(&MediaUserProxy{}).Error
}
func (this *MediaUser) EmptyTag(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("user_id = ?", this.Id).Delete(&MediaUserToTag{}).Debug().Error
}

// 获取需要更新或者下载的用户
func GetAutoUsers() []*MediaUser {
	var mus []*MediaUser
	dbs.Model(&MediaUser{}).Where("auto_download = ? Or autoinfo = ?", 1, 1).Find(&mus)

	return mus
}

// 获取可用的client
// 返回客户端类型和客户端id
func (mu *MediaUser) GetCanUseClient() (int, int64) {
	cls := mu.GetClients()
	var idx int
	var clientType int
	var clientId int64
	for { // 循环检测客户端是否可用,不可用则换
		if clientID, clientType, ok := RandomInt64Fast(cls); ok {
			clientId, clientType = clientID, clientType
			if runner.IsRuning(clientType, clientId) {
				break
			}
		}
		idx++
		if idx >= 10 {
			break
		}
	}
	return clientType, clientId
}

// 获取用户的主页url和执行js
func (mu *MediaUser) GetJsAndUrl() (string, string, error) {
	var jscode string
	var runurl string
	switch mu.Platform {
	case "douyin":
		jscode = "douyin-info"
		runurl = fmt.Sprintf("https://www.douyin.com/user/%s", mu.Uuid)
	default:
		err := fmt.Errorf("暂时不支持 %s 获取获取信息", mu.Platform)
		return "", "", err
	}

	js := jses.GetJsByCode(jscode)
	if js == nil || js.ID < 1 {
		err := fmt.Errorf("获取账号信息脚本不存在")
		return "", "", err
	}
	return runurl, js.GetContent(nil), nil
}

// 获取用户代理
func (mu *MediaUser) GetMyProxy() {

}

func (mu *MediaUser) GenBrowserOpt(id int64) (*bs.Options, error) {
	gotourl, jsstr, err := mu.GetJsAndUrl()
	if err != nil {
		return nil, err
	}

	var proxyID int64
	var proxyCongig string
	bopt := &bs.Options{
		ID:    id,
		Url:   gotourl,
		JsStr: jsstr,
	}

	// 代理
	if id > 0 {
		bbs, err := browserdb.GetBrowserById(id)
		if err != nil {
			id = 0
			bopt.ID = 0
		} else {
			bopt.Height = bbs.Height
			bopt.Width = bbs.Width
			bopt.Language = bbs.Lang
			bopt.Timezone = bbs.Timezone
			// bopt.Msg = make(chan string)
			bopt.Headless = true
			proxyID = bbs.Proxy
			proxyCongig = bbs.ProxyConfig
		}
	} else {
		if ps := mu.GetProxys(); len(ps) > 0 {
			proxyID = ps[rand.Intn(len(ps))]
		}
	}
	if proxyID > 0 {
		if bopt.Pc, err = proxys.GetProxyConfigByID(proxyID); err != nil {
			logs.Error("自动获取用户信息错误: " + err.Error())
		}
	} else if proxyCongig != "" {
		if bopt.Pc, err = proxy.Client(proxyCongig, "", 0, ""); err != nil {
			logs.Error("自动获取用户信息错误: " + err.Error())
		}
	}
	return bopt, nil
}

// 获取用户信息
func (mu *MediaUser) GetInfoFromPlatform(ch chan byte) error {
	tp, id := mu.GetCanUseClient()
	var opt any
	var err error
	var r runner.Runner
	switch tp {
	case 0:
		bopt, err := mu.GenBrowserOpt(id)
		if err != nil {
			return err
		}

		var cacle context.CancelFunc
		bopt.Ctx, cacle = context.WithTimeout(mainsignal.MainCtx, time.Second*30)
		bopt.Msg = make(chan string)
		opt = bopt
		go func() {
			var idx int
			for {
				select {
				case <-bopt.Ctx.Done():
					fmt.Println("----浏览器执行结束")
					goto BREAK
				case msg := <-bopt.Msg:
					if mu.ParseUserInfoData(msg) {
						if r != nil {
							r.Stop()
						}
						goto BREAK
					} else if r != nil {
						idx++
						if idx >= 5 {
							goto BREAK
						}
						r.Start()
					}
				}
			}
		BREAK:
			cacle()
			if ch != nil {
				ch <- 1
			}
			eventbus.Bus.Publish("media_user_info", mu)
		}()
	case 1:
		return errors.New("暂不支持手机端获取")
	case 2:
		return errors.New("不支持HTTP方式获取")
	}

	r, err = runner.GetRunner(tp, opt)
	if err != nil {
		return err
	}
	return r.Start()
}

// 解析返回的用户信息数据
func (mu *MediaUser) ParseUserInfoData(data string) bool {
	gs := gjson.Parse(data)
	if fans := gs.Get("fans").Int(); fans > 0 {
		mu.Fans = fans
	}
	if works := gs.Get("works").Int(); works > 0 {
		mu.Works = works
	}
	if local := gs.Get("local").String(); local != "" {
		mu.Local = local
	}
	if account := gs.Get("account").String(); account != "" {
		mu.Account = account
	}
	mu.Save(nil)
	mu.Commpare()
	return true
}
