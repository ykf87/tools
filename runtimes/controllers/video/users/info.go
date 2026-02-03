package users

import (
	"fmt"
	"strconv"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/medias"
	"tools/runtimes/eventbus"
	"tools/runtimes/response"

	"github.com/chromedp/cdproto/runtime"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var dyinfojs = `(async () => {
	function parseCNNumber(str) {
	  if (!str) return -1;

	  const s = String(str).replace(/\s+/g, '');

	  const match = s.match(/^([\d.]+)(万|亿)?$/);
	  if (!match) return -1;

	  const num = parseFloat(match[1]);
	  const unit = match[2];

	  if (Number.isNaN(num)) return -1;

	  switch (unit) {
	    case '万':
	      return Math.round(num * 1e4);
	    case '亿':
	      return Math.round(num * 1e8);
	    default:
	      return num;
	  }
	}
	async function getFirstFollowNumber(datae2e) {
	  const root = document.querySelector('[data-e2e="user-info"]');
	  if (!root) return -1;

	  const nodes = root.querySelectorAll('div[data-e2e="'+datae2e+'"] > div');

	  let i 	= 0;
	  while(true){
		for (const el of nodes) {
		    var text = el.textContent.trim();
		    if(text == ""){
		    	continue
		    }
		    text 	= parseCNNumber(text);
		    if (/^\d+$/.test(text)) {
		      return Number(text);
		    }
		}
		if(i++ >= 10){
			break;
		}
		await new Promise(r => setTimeout(r, 1000));
	  }

	  return -1;
	}


	var resp 			= {};
	resp['fans'] 		= await getFirstFollowNumber("user-info-fans");
	resp['follow'] 		= await getFirstFollowNumber("user-info-follow");
	resp['zan'] 		= await getFirstFollowNumber("user-info-like");
	var root 			= document.querySelector('[data-e2e="user-info"]');
	if(root){
		var nodes = root.querySelectorAll(':scope > p > span');
		for (const el of nodes) {
		    const text = el.textContent.trim();
		    try{
		    	if (text.indexOf("IP") >= 0) {
		    		resp['local']	= text.split("：")[1];
			    }else if(text.indexOf("抖音号") >= 0){
					resp['account'] 	= text.split("：")[1];
				}
		    }catch(e){}
		  }
	}
	var zuopin 		= document.querySelector('[data-e2e="user-tab-count"]');
	if(zuopin){
		resp['works'] 	= Number(zuopin.textContent.trim());
	}
	console.log(JSON.stringify({"type":"kaka", "data":resp}));
	return JSON.stringify({"type":"kaka", "data":resp});
})()`

func GetInfo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	id64 := int64(id)
	mu := medias.GetMediaUserByID(id64)
	if len(mu.Clients) > 0 {

	} else {
		go func() {
			// bbs := bs.NewManager("")
			brows, _ := bs.BsManager.New(0, &bs.Options{
				Url:      fmt.Sprintf("https://www.douyin.com/user/%s", mu.Uuid),
				JsStr:    dyinfojs,
				Headless: true,
				Timeout:  time.Duration(time.Second * 30),
			}, true)

			brows.OnClosed(func() {
				eventbus.Bus.Publish("media_user_info", mu)
			})
			brows.OnConsole(func(args []*runtime.RemoteObject) {
				for _, arg := range args {
					if arg.Value != nil {
						gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
						if gs.Get("type").String() == "kaka" {
							dt := gs.Get("data")
							if fans := dt.Get("fans").Int(); fans > 0 {
								mu.Fans = fans
							}
							if works := dt.Get("works").Int(); works > 0 {
								mu.Works = works
							}
							if local := dt.Get("local").String(); local != "" {
								mu.Local = local
							}
							if account := dt.Get("account").String(); account != "" {
								mu.Account = account
							}
							mu.Save(nil)
							mu.Commpare()
							brows.Close()
						}
					} else if arg.Description != "" {
						// fmt.Println(arg.Description, "description")
					} else {
						// fmt.Println("[unknown console arg]")
					}
				}
			})

			brows.OnURLChange(func(url string) {
				// fmt.Println("------")
				// brows.RunJs(dyinfojs)
			})

			brows.OpenBrowser()
			time.Sleep(time.Second * 1)
			go brows.RunJs(dyinfojs)

			go func() {
				time.Sleep(time.Second * 30)
				if brows.IsArrive() {
					brows.Close()
				}
			}()
		}()
	}
	response.Success(c, mu, "")
}
