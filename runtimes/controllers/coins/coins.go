package coins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"tools/runtimes/db/admins"
	"tools/runtimes/i18n"
	"tools/runtimes/listens/ws"
	"tools/runtimes/mainsignal"
	"tools/runtimes/response"
	"tools/runtimes/trade/do"

	"github.com/gin-gonic/gin"
)

const proxy = "http://192.168.10.101:59667"

func AllCanUse(c *gin.Context) {
	sm, err := do.GetSameID(proxy)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var ssi []map[string]any
	for k, v := range sm {
		ssi = append(ssi, map[string]any{
			"name":  k,
			"base":  v.Base,
			"quote": v.Quote,
		})
	}

	var selecteds []string
	listenCoin.Range(func(k, v any) bool {
		selecteds = append(selecteds, k.(string))
		return true
	})

	response.Success(c, gin.H{
		"sss":      ssi,
		"selected": selecteds,
	}, "")
}

var listenCoin sync.Map

type ListenObj struct {
	Name     string `json:"name"`
	Interval int    `json:"interval"` // 更新间隔,秒
	startAt  int64  `json:"-"`        // 开始时间
	started  atomic.Bool
	ctx      context.Context
	cancle   context.CancelFunc
	adminID  int64
}

func (l *ListenObj) start() {
	if l.started.Load() {
		return
	}
	l.started.Store(true)
	listenCoin.Store(l.Name, l)

	l.ctx, l.cancle = context.WithCancel(mainsignal.MainCtx)

	if l.Interval < 1 {
		l.Interval = 1
	}
	ticker := time.NewTicker(time.Second * time.Duration(l.Interval))
	defer func() {
		ticker.Stop()
		l.started.Store(false)
	}()
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			res, err := do.GetSingleCointTick(proxy, l.Name)
			if err == nil {
				bt, err := json.Marshal(gin.H{
					"type": fmt.Sprintf("listen-coin-%s", l.Name),
					"data": res,
				})
				if err == nil {
					ws.SentMsg(l.adminID, bt)
				}
			}
		}
	}
}

func (l *ListenObj) stop() {
	l.cancle()
}

func AddCoin(c *gin.Context) {
	var req *ListenObj
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	if req.Name == "" {
		response.Error(c, http.StatusBadGateway, i18n.T("请选择币种"), nil)
		return
	}

	if as, ok := listenCoin.Load(req.Name); ok {
		if ls, ok := as.(*ListenObj); ok {
			ls.stop()
		}
	}

	user, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	req.adminID = user.Id

	// listenCoin.Store(req.Name, req)
	go req.start()

	response.Success(c, nil, i18n.T("添加监听列表成功!"))
}

func StopCoin(c *gin.Context) {
	coin := c.Param("name")

	if cl, ok := listenCoin.Load(coin); ok {
		if cll, ok := cl.(*ListenObj); ok {
			cll.stop()
		}
		listenCoin.Delete(coin)
	}
	response.Success(c, nil, "停止成功")
}

func StopAll(c *gin.Context) {
	listenCoin.Range(func(key, value any) bool {
		if cl, ok := value.(*ListenObj); ok {
			cl.stop()
			listenCoin.Delete(key)
		}
		return true
	})
	response.Success(c, nil, "停止成功")
}
