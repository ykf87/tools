package platform

import (
	"tools/runtimes/trade"
	"tools/runtimes/trade/platform/gate"
	"tools/runtimes/trade/platform/okx"
)

type Platform interface {
	GetSpotCurr(code string) (*trade.CurrencyInfo, error)
	AllSpot() (map[string]*trade.SpotCurrency, error)
}

func GetClient(plat, key, secret, password, proxy string) (Platform, error) {
	var plm Platform
	switch plat {
	case "okx":
		plm = okx.Client(key, secret, password, proxy)
	case "gate":
		plm = gate.Client(key, secret, proxy)
	}
	return plm, nil
}
