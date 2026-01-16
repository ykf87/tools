package users

import (
	"fmt"
	"tools/runtimes/bs"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/chromedp/cdproto/runtime"
	"github.com/gin-gonic/gin"
)

func GetInfo(c *gin.Context) {
	mu := medias.GetMediaUserByID(c.Query("id"))
	if len(mu.Clients) > 0 {

	} else {
		bbs := bs.NewManager("")
		brows, _ := bbs.New(0, bs.Options{
			Url: fmt.Sprintf("https://www.douyin.com/user/%s", mu.Uuid),
		})

		brows.OnClosed(func() {
			// eventbus.Bus.Publish("browser-close", this)
		})
		brows.OnConsole(func(args []*runtime.RemoteObject) {
			fmt.Println(args, "args-----")
		})

		brows.OnURLChange(func(url string) {
			fmt.Println("------")
		})

		brows.OpenBrowser()
	}
	response.Success(c, mu, "")
}
