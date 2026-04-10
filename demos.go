package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/clearer"
	"tools/runtimes/config"
	"tools/runtimes/db/audios"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/proxys"
	"tools/runtimes/downloader"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/funcs"
	"tools/runtimes/imager"
	"tools/runtimes/ipquality"
	"tools/runtimes/mainsignal"
	"tools/runtimes/obschan"
	"tools/runtimes/proxy"
	"tools/runtimes/sch"
	"tools/runtimes/storage"
	"tools/runtimes/trade/do"
	"tools/runtimes/videoproc"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

func init() {
	// testBrowser()
	// download()
	// fmt.Println(config.Storages.Load("").PutStr(config.FullPath("44.mp4")))
	// u := "https://v9-cold1.douyinvod.com/cbef1952f7728d765445ce59d9ed49e8/69af9393/video/tos/cn/tos-cn-ve-15/o8xNppKJEmBRAFgZbgDQfgAzAZ6TBEDoAI8f9F/?a=1128&br=1829&bt=1829&btag=c0010e000a8000&cd=0%7C0%7C0%7C0&ch=0&cquery=100y&cr=0&cs=0&cv=1&dr=0&ds=4&dy_q=1773110527&dy_va_biz_cert=&feature_id=0ea98fd3bdc3c6c14a3d0804cc272721&ft=BaXAWVVywfyRF38Pmo~pK7pswApzZh-_vrKnZwocdo0g3cI&l=202603101042078FEA6334B58181FCF44E&mime_type=video_mp4&qs=0&rc=OjVmZTkzZWk7ZGY8N2hlZ0BpMzU1PG45cnBrODMzNGkzM0BfMy8yLmIvX2MxMjA0MzA0YSNlb18vMmRjaDNhLS1kLTBzcw%3D%3D"
	// n, e := storage.Load("minio").Download(mainsignal.MainCtx, u, &downloader.DownloadOption{
	// 	Callback: func(total, downloaded, speed, workers int64) {
	// 		fmt.Printf(
	// 			"\r%.2f%% %s/s workers:%d %s",
	// 			float64(downloaded)/float64(total)*100,
	// 			funcs.FormatFileSize(speed, "1", ""),
	// 			workers,
	// 			funcs.FormatFileSize(total, "1", ""),
	// 		)
	// 	},
	// })
	// fmt.Println(n, e)
	// getdouytest()
	//
	// medias.MKDBNameID("dfgdfgaaa", 0)
	// runffmpeg()
	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	revideo()
	// }()

	// imgmk()
	// mkvideo()
	// checkip()
	// textobcshan()

	// traded()
}

func traded() {
	// trade.AddRow("gate", &trade.Api{
	// 	Key:    "8a0546c1c81e09a72bd9832e50865eb1",
	// 	Secret: "0ffe575d1d460b730eca122132d125d2787243cd689fbfd163207e4de8ad1ed5",
	// })
	// trade.AddRow("okx", &trade.Api{
	// 	Key:      "cdde4726-62c0-4347-9d99-e04a4db6941c",
	// 	Secret:   "3FC57FA5534F7BE28345E55BF1090E44",
	// 	Password: "Abcd@1234",
	// })
	// err := trade.SaveConfig()
	// fmt.Println(err)
	// return

	go func() {
		time.Sleep(time.Second * 3)
		mps, err := do.GetSameID("")
		if err == nil {
			for id, vv := range mps {
				fmt.Println(id, vv.ID, "---")
			}
		} else {
			fmt.Println(err)
		}
	}()

}

func textobcshan() {
	ch := obschan.NewObservableChan(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 阻塞发送测试
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			err := ch.SendContext(ctx, 42)
			defer ch.RecvContext(ctx)
			time.Sleep(time.Second * 5)
			if err != nil {
				fmt.Println("send canceled")
			} else {
				fmt.Println("send ok")
			}

			fmt.Println("Len:", ch.Len())
			fmt.Println("WaitingSend:", ch.WaitingSend())
			fmt.Println("WaitingRecv:", ch.WaitingRecv())
		})
	}
	wg.Wait()

	// 主线程等待
	// time.Sleep(2 * time.Second)

	// fmt.Println("Len:", ch.Len())
	// fmt.Println("WaitingSend:", ch.WaitingSend())
	// fmt.Println("WaitingRecv:", ch.WaitingRecv())
}

func checkip() {
	geo, _ := ipquality.NewGeoIP(config.FullPath(config.SYSROOT, "GeoLite2-City.mmdb"), config.FullPath(config.SYSROOT, "GeoLite2-ASN.mmdb"))
	ipdb, _ := ipquality.NewIPInfoMMDB(config.FullPath(config.SYSROOT, "ipinfo_lite_sample.mmdb"))
	sql, _ := ipquality.NewSQLiteCache(config.FullPath(config.SYSROOT, "ipquality.db"))
	client := ipquality.NewClient(geo, ipdb, sql)
	res, _ := client.Check("81.181.237.241") //161.77.219.175

	fmt.Println(res.Allow, res.Score, res.Type, res.Reason)
}

func mkvideo() {
	audio := &videoproc.AudioInpter{
		Url: storage.Load("").URL("81/b7/81b7b7a8d017fef834ca41b71d9027e5246239f58b232bc27a645254a724ca39.mp3"),
	}
	n1, _ := videoproc.SecMaker([]string{
		storage.Load("").URL("e6/b6/e6b64f9efa9749dbedce693a11d335acbc8ed5aaa810bc4e0e35566328e15ac4.mp4"),
		// storage.Load("").URL("82/7b/827b26cf31580022db6d481bc7a7cfb19a4f473ab03182c5df037347ee68f663.mp4"),
	}, nil, mainsignal.MainCtx, nil)
	n1.Audio = audio
	// n1.AmixAudio = &videoproc.AudioInpter{
	// 	Url: storage.Load("").URL("81/b7/81b7b7a8d017fef834ca41b71d9027e5246239f58b232bc27a645254a724ca39.mp3"),
	// }

	go func() {
		time.Sleep(time.Second * 2)
		n1.Factory.Linear = &imager.Linear{
			Brightness: 3,
			Contrast:   1.2,
		}
		// n1.Factory.Mirror = 1

		// n1.Factory.Clearer = true
		err := n1.Output(config.FullPath(config.MEDIAROOT, ".tmp", "20260330.mp4"))
		fmt.Println(err, "----")
	}()
}

// 图片修改
func imgmk() {
	img, _ := imager.NewImager(config.FullPath(config.DATAROOT, "121.jpg"))
	img.Gamma = &imager.Gamma{
		Value: 1.12,
	}
	// img.Linear = &imager.Linear{
	// 	Brightness: 20,
	// 	Contrast:   2,
	// }
	// img.Crop = &imager.Crop{
	// 	Left:   0.1,
	// 	Top:    0.2,
	// 	Right:  0.4,
	// 	Bottom: 0.6,
	// }
	// img.Rotation = &imager.Rotation{
	// 	Angle: -36,
	// }
	// img.Gaussblur = &imager.Gaussblur{
	// 	Value: 3.5,
	// }
	// img.Sharpen = &imager.Sharpen{// 锐化不支持
	// 	Value: 3,
	// }
	//
	// img.Resize = &imager.Resize{
	// 	Scale: 1.6,
	// }
	// kwh := imager.KeepWH(true)
	// img.KeepWH = &kwh
	// fmt.Println(img.Output(config.FullPath(config.DATAROOT, "121---out.jpg"), 1))
}

// 图片变清晰
func cleare() {
	clearer.Init()
	clearer.Clearers(config.FullPath(config.DATAROOT, "bt.png"), config.FullPath(config.DATAROOT, "bt-output.jpg"), "")
}

// 视频去重
func revideo() {
	cfg := videoproc.VideoConfig{
		Input:  "./data/2.mp4",
		Output: "./data/output5.mp4",
		CRF:    18,
		Preset: "slow",
		Audio:  storage.Load("minio").URL("67/9d/679d7ff70e994465e338ff60231759945bd5be66be0992da032ac294a44b7d6b.mp3"),
	}

	err := videoproc.ProcessVideo(cfg)
	if err != nil {
		panic(err)
	}
}

// ffmpeg
func runffmpeg() {
	rr, err := ffmpeg.GetAudioInfo("./data/2.mp4")
	fmt.Println(*rr, err)

	r, err := storage.Load("minio").PutStr("./data/f.mp3")
	fmt.Println(r, err)

	// u := storage.Load("minio").URL("9e/7a/9e7aa05f1ed0ecbdc386fe4c549ac3d37e2e443891123facddfb2b08a453d650.mp3")

	cc, err := audios.AddAudio("df/2b/df2baf0d9268ec62d9f1ab207cc8cec562a9b74aae7d1a665408d139ccc303ec.mp4", "aaa")
	fmt.Println(cc, err)
}

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

func getdouytest() {
	urlstr := `7.17 pDH:/ 06/28 A@T.lC 刘备遇到水镜，经意外得到神级军师 # 水镜先生 # 刘备 # 徐庶走马荐诸葛 # # 徐庶 # 新三国解说  https://v.douyin.com/4f_EMBbDKdk/ 复制此链接，打开Dou音搜索，直接观看视频！
	6.12 gbn:/ Q@k.pd 08/10 🐇好可爱吖🐰  https://v.douyin.com/ffn9ejtGAUE/ 复制此链接，打开Dou音搜索，直接观看视频！
	9.41 GvF:/ 08/21 v@F.UY 热浪岛🏝️ carefree vibe# 精神稳定的成年人  https://v.douyin.com/ECaiQV2JWZg/ 复制此链接，打开Dou音搜索，直接观看视频！
	0.00 P@X.Zm Xzg:/ 09/12 # 不就反差吗这题我熟 # 巨蟹座  https://v.douyin.com/CR7foC7SU8s/ 复制此链接，打开Dou音搜索，直接观看视频！
	https://www.douyin.com/video/7615528126583892809`
	go func() {
		err := medias.GetPlatformVideos(urlstr, nil, "autodownload", 1, "", true, false)
		fmt.Println(err)
	}()

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
					// if b.Opts.Msg != nil {
					// 	select {
					// 	case b.Opts.Msg <- "点击按钮":
					// 	case <-b.Opts.Ctx.Done():
					// 	}
					// }
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

func download() {
	name, err := downloader.Download(mainsignal.MainCtx, &downloader.DownloadOption{
		URL:      "https://v5-small.douyinvod.com/df5b4735d2b9bc2dc1449e1b89c3329c/69ae3b32/video/tos/cn/tos-cn-ve-15c000-ce/o0AEw01IdXykEK00Kir1AuiBtB1Zf3AleQChWe/?a=1128&ch=0&cr=0&dr=0&er=0&cd=0%7C0%7C0%7C0&cv=1&br=1682&bt=1682&cs=0&ds=4&ft=LjhJkw998xI7uEPmD0P5NdvaUFiXHU4nkVJEdfQAVbPD-Ipz&mime_type=video_mp4&qs=0&rc=aTk4OWhmO2Q2PGRoZTRmOEBpMzh2bG85cnRxOTMzbGkzNEAwYTA2MTIvXl4xXy02L2BfYSNuLmZiMmRzLmphLS1kLWJzcw%3D%3D",
		Dir:      "./",
		FileName: "",
		Threads:  8,
		// Headers: map[string]string{
		// 	"User-Agent": "Mozilla/5.0",
		// },
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
}
