package bs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// 设置时区
func (this *BrowserConfigFile) SetTimezone(timezone string) {
	loc, err := time.LoadLocation(timezone)

	if err != nil {
		return
	}

	if this.TimeZone == nil {
		this.TimeZone = new(TimezoneStruct)
	}
	_, offset := time.Now().In(loc).Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	sign := "+"
	if hours < 0 || minutes < 0 {
		sign = "-"
	}
	zone := fmt.Sprintf("UTC%s%02d:%02d", sign, abs(hours), abs(minutes))

	this.TimeZone.Locale = ""
	this.TimeZone.Mode = 2
	this.TimeZone.Name = this.TimeZone.GetName(timezone)
	this.TimeZone.Utc = timezone
	this.TimeZone.Value = 8
	this.TimeZone.Zone = zone
}

func (this *BrowserConfigFile) SetHomePage(url string) {
	this.Homepage.Mode = 1
	this.Homepage.Value = url
}

func GetBrowserConfigDir(id int64) (string, error) {
	dir := filepath.Join(BASEPATH, fmt.Sprintf("%d", id))
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
	}
	return dir, nil
}
