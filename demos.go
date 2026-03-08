package main

import (
	"context"
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/proxys"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/proxy"
	"tools/runtimes/sch"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

//	func init() {
//		s := scheduler.New(mainsignal.MainCtx)
//		rt := s.NewRunner(func(ctx context.Context) error {
//			fmt.Println("执行测试代码----")
//			return fmt.Errorf("错误咯---")
//		}, time.Second*5, nil)
//		rt.Every(time.Second * 1).RunNow()
//	}
// func init() {
// tk, err := task.NewTask("test", 1, "测试task", 5, true)
// if err != nil {
// 	fmt.Println("构建任务失败", err)
// 	return
// }
// tr, err := tk.AddChild("text-1", "测试执行", time.Minute*60)
// if err != nil {
// 	fmt.Println("构建子任务失败", err)
// 	return
// }
// tr.StartInterval(30, func(tr *task.TaskRun) error {
// 	// fmt.Println("-------执行", tr.RunID, tr.Title)
// 	tr.ReportSchedule(90, 78)
// 	time.Sleep(time.Second * 3)
// 	if tr.GetTried() >= 1 {
// 		tr.ReportSchedule(90, 90)
// 		return nil
// 	}
// 	return fmt.Errorf("错误的任务:%s", tr.RunID)
// })

// tr.StartAtTime(-28800000, func(tr *task.TaskRun) error {
// 	fmt.Println("-------执行", tr.RunID, tr.Title)

//		if tr.GetTried() >= 1 {
//			return nil
//		}
//		return fmt.Errorf("错误的任务:%s", tr.RunID)
//	})
//
// }
func init() {
	// testBrowser()
	mainsignal.MainWait.Go(func() {
		name, err := downloader.Download(mainsignal.MainCtx, &downloader.DownloadOption{
			URL:      "https://v11-cold1.douyinvod.com/a5dc9977fb9517b2c8d80b9f4c54b83c/69adb365/video/tos/cn/tos-cn-ve-15c000-ce/oUueNvvItU7vhIBQGCLTATEGgQ8haJHjePKBeZ/?a=1128&ch=0&cr=0&dr=0&cd=0%7C0%7C0%7C0&cv=1&br=1772&bt=1772&cs=0&ds=4&ft=BaXAWVVywfyRF38Pmo~pK7pswAp-bH-_vrKnZwocdo0g3cI&mime_type=video_mp4&qs=0&rc=ZzM6aGc8PGc7OGZlOWc5O0BpM3Fpd3A5cnM6OTMzbGkzNEAxXjAtLi00XjExL18vXzA1YSMxcjZhMmRrb15hLS1kLWJzcw%3D%3D&btag=c0010e000ad000&cquery=100y&dy_q=1772987524&feature_id=0ea98fd3bdc3c6c14a3d0804cc272721&l=202603090032048422CC1B1067BB5B8E91",
			Dir:      "./",
			Threads:  8,
			FileName: "",
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			Callback: func(total, cur, speed, workers int64) {
				fmt.Printf(
					"\r%.2f%% %s/s workers:%d %s",
					float64(cur)/float64(total)*100,
					funcs.FormatFileSize(speed, "1", ""),
					workers,
					funcs.FormatFileSize(total, "1", ""),
				)
			},
		})
		fmt.Println("\n下载完成：", name, err)
	})
}

func scheduler() {
	s := sch.NewScheduler(5)
	tr, err := s.AddInterval(
		"task1",
		10*time.Second, // interval
		5*time.Second,  // timeout
		2,              // retry
		2*time.Second,  // retryDelay
		time.Now().Add(time.Second*20),
		0,
		func(ctx context.Context) error {
			fmt.Println("interval task running:", time.Now())
			time.Sleep(2 * time.Second)
			return nil
		},
		func(id string, err error) {
			fmt.Println("task complete:", id, "err:", err)
		},
		func(id string) {
			fmt.Println("task closed:", id)
		},
	)
	s.RunNow(tr.GetID())
	fmt.Println(err)
}

func testBrowser() {
	var jsstr string
	je := jses.GetJsByCode("ddd")
	if je != nil {
		jsstr = je.GetContent(map[string]any{"min_num": 4, "max_num": 30, "zan_hit": 58})
	}

	pro := proxys.GetById(136)
	ch := make(chan *proxy.ProxyConfig)
	go func() {
		pc, err := pro.Start(false)
		if err != nil {
			ch <- nil
		} else {
			ch <- pc
		}
	}()
	pc := <-ch
	if pc == nil {
		panic("代理启动失败")
	}

	b, err := bs.BsManager.New(1, &bs.Options{
		ID:    1,
		Url:   "https://www.tiktok.com/",
		JsStr: jsstr,
		Show:  true,
		Pc:    pc,
	}, true)
	if err != nil {
		panic(err)
	}

	b.OnURLChange(func(s string) {
		fmt.Println("url发生改变:", s)
		b.RunJs(jsstr)
	})

	b.OnConsole(func(args []*runtime.RemoteObject) {
		for _, arg := range args {
			if arg.Value != nil {
				gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
				if gs.Get("version").String() == "" {
					continue
				}
				fmt.Println(gs.String(), "-----------")
				switch gs.Get("type").String() {
				case "click": // 点击
					fmt.Println("点击按钮触发")
					x := gs.Get("x").Float()
					y := gs.Get("y").Float()
					b.Click(x, y)
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- "点击按钮":
						case <-b.Opts.Ctx.Done():
						}
					}
				case "success":
					fmt.Println("执行完成!")
					b.Close()
				case "fail":
					fmt.Println(gs.Get("msg").String())
					b.Close()
				}
			}
		}
	})
	fmt.Println(b.OpenBrowser())
	// b.GoToUrl("https://www.tiktok.com/")
}
