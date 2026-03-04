package users

import (
	"net/http"
	"strconv"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
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

	// ch := make(chan byte)
	if err := mu.GetInfoFromPlatform(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, mu, "")
}

func UserMeidas(c *gin.Context) {
	response.Success(c, medias.GetUserMedias(c.Query("id")), "success")
}
