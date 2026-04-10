package do

import (
	"tools/runtimes/trade"
	"tools/runtimes/trade/platform"
)

// http://192.168.10.101:59667
func GetSameID(proxy string) (map[string]*trade.SpotCurrency, error) {
	mps, err := trade.Init()
	if err != nil {
		return nil, err
	}

	mpps := make(map[string]map[string]*trade.SpotCurrency)
	for platName, v := range mps {
		if v.Disable == true {
			continue
		}
		p, err := platform.GetClient(platName, v.Key, v.Secret, v.Password, proxy)
		if err != nil {
			return nil, err
		}
		mpps[platName], err = p.AllSpot()
		if err != nil {
			return nil, err
		}
	}

	var first map[string]*trade.SpotCurrency
	result := make(map[string]*trade.SpotCurrency)
	for _, v := range mpps {
		first = v
		break
	}

	for k, resp := range first {
		ok := true

		for _, inner := range mpps {
			if _, exist := inner[k]; !exist {
				ok = false
				break
			}
		}

		if ok {
			result[k] = resp
		}
	}
	return result, nil
}

// 获取单个币种的交易数据
func GetSingleCointTick(proxy, code string) (map[string]*trade.CurrencyInfo, error) {
	mps, err := trade.Init()
	if err != nil {
		return nil, err
	}

	reslult := make(map[string]*trade.CurrencyInfo)

	for platName, v := range mps {
		if v.Disable == true {
			continue
		}

		p, err := platform.GetClient(platName, v.Key, v.Secret, v.Password, proxy)
		if err != nil {
			return nil, err
		}

		rs, err := p.GetSpotCurr(code)
		if err != nil {
			return nil, err
		}

		reslult[platName] = rs
	}
	return reslult, nil
}
