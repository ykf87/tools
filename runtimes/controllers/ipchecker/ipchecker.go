package ipchecker

import (
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/i18n"
	"tools/runtimes/ipquality"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func IpCheck(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		response.Error(c, http.StatusBadGateway, i18n.T("请传入正确的ip"), nil)
		return
	}
	geo, err := ipquality.NewGeoIP(config.FullPath(config.SYSROOT, "GeoLite2-City.mmdb"), config.FullPath(config.SYSROOT, "GeoLite2-ASN.mmdb"))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	ipdb, err := ipquality.NewIPInfoMMDB(config.FullPath(config.SYSROOT, "ipinfo_lite_sample.mmdb"))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	sql, err := ipquality.NewSQLiteCache(config.FullPath(config.SYSROOT, "ipquality.db"))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	client := ipquality.NewClient(geo, ipdb, sql)

	res, err := client.Check(ip)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	response.Success(c, res, "")
}
