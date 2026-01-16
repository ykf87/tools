package users

import (
	"fmt"
	"strconv"
	"strings"
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
	async function getFirstFollowNumber(datae2e) {
	  const root = document.querySelector('[data-e2e="user-info"]');
	  if (!root) return 0;

	  const nodes = root.querySelectorAll('div[data-e2e="'+datae2e+'"] > div');

	  let i 	= 0;
	  while(true){
		for (const el of nodes) {
		    const text = el.textContent.trim();
		    if (/^\d+$/.test(text)) {
		      return Number(text);
		    }
		}
		if(i++ >= 10){
			break;
		}
		await new Promise(r => setTimeout(r, 2000));
	  }

	  return 0;
	}

	while(true){
		if(document.querySelector('[data-e2e="user-info"]')){
			break;
		}
	}
	var resp 		= {};
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
			    }
		    }catch(e){}
		  }
	}
	var zuopin 		= document.querySelector('[data-e2e="user-tab-count"]');
	if(zuopin){
		resp['works'] 	= Number(zuopin.textContent.trim());
	}
	return JSON.stringify({"type":"kaka", "data":resp});
})()`

func GetInfo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	id64 := int64(id)
	mu := medias.GetMediaUserByID(id64)
	if len(mu.Clients) > 0 {

	} else {
		bbs := bs.NewManager("")
		brows, _ := bbs.New(0, bs.Options{
			Url:      fmt.Sprintf("https://www.douyin.com/user/%s", mu.Uuid),
			JsStr:    dyinfojs,
			Headless: true,
		})

		brows.OnClosed(func() {
			// eventbus.Bus.Publish("browser-close", this)
		})
		brows.OnConsole(func(args []*runtime.RemoteObject) {
			for _, arg := range args {
				if arg.Value != nil {
					// fmt.Println(arg.Value, "value")
					val := strings.ReplaceAll(arg.Value.String(), "\\", "")
					gs := gjson.Parse(val)
					if gs.Get("type").String() == "kaka" {
						fmt.Println(gs.Get("data").String())
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

		go func() {
			var i int
			defer brows.Close()
			for {
				rsp, err := brows.RunJs(dyinfojs)
				if err == nil {
					if resstr, ok := rsp.(string); ok {
						gs := gjson.Parse(resstr)
						dt := gs.Get("data")
						fmt.Println(dt.String(), "-----")
						if fans := dt.Get("fans").Int(); fans > 0 {
							mu.Fans = fans
						}
						if works := dt.Get("works").Int(); works > 0 {
							mu.Works = works
						}
						if local := dt.Get("local").String(); local != "" {
							mu.Local = local
						}
						mu.Save(nil)
						mu.Commpare()
						eventbus.Bus.Publish("media_user_info", mu)
						break
					}
				}
				time.Sleep(time.Second * 3)
				i++
				if i == 9 {
					eventbus.Bus.Publish("media_user_info", mu)
					break
				}
			}
		}()
	}
	response.Success(c, mu, "")
}
