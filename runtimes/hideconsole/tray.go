package hideconsole

import (
	"tools/runtimes/mainsignal"

	"github.com/getlantern/systray"
)

type MenuItem struct {
	Title    string                  `json:"title"`
	Tips     string                  `json:"tips"`
	Icon     []byte                  `json:"icon"`
	Callback func(*systray.MenuItem) `json:"-"`
	si       *systray.MenuItem       `json:"-"`
}

type Menu struct {
	Title string      `json:"title"`
	Tips  string      `json:"tips"`
	Icon  []byte      `json:"icon"`
	Items []*MenuItem `json:"items"`
}

var sysMenu Menu

func Build(title, tip string, icon []byte) {
	sysMenu.Tips = tip
	sysMenu.Icon = icon
	sysMenu.Title = title
}

func AddItem(title, tip string, icon []byte, callback func(*systray.MenuItem)) {
	item := new(MenuItem)
	item.Title = title
	item.Tips = tip
	item.Icon = icon
	// item.si = systray.AddMenuItem(title, tip)
	// if icon != ""{
	// 	item.si.SetIcon([]byte(icon))
	// }
	item.Callback = callback
	sysMenu.Items = append(sysMenu.Items, item)
}

func Run(quiteCallback func()) {
	go systray.Run(func() {
		systray.SetTitle(sysMenu.Title)
		systray.SetTooltip(sysMenu.Tips)
		if sysMenu.Icon != nil {
			systray.SetTemplateIcon(sysMenu.Icon, sysMenu.Icon)
			systray.SetIcon(sysMenu.Icon)
		}

		for _, v := range sysMenu.Items {
			v.si = systray.AddMenuItem(v.Title, v.Tips)
			if v.Icon != nil {
				v.si.SetIcon(v.Icon)
			}
			if v.Callback != nil {
				go func() {
					for {
						select {
						case <-v.si.ClickedCh:
							v.Callback(v.si)
						case <-mainsignal.MainCtx.Done():
							return
						}
					}
				}()
			}

		}
	}, quiteCallback)
}
