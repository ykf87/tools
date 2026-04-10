package okx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"tools/runtimes/trade"
)

func (o *Okx) AllSpot() (map[string]*trade.SpotCurrency, error) {
	// var pms url.Values
	pms := &url.Values{}
	pms.Add("instType", "SPOT")
	b, _, err := o.doNoAuthRequest(http.MethodGet, "/api/v5/public/instruments", pms)
	if err != nil {
		fmt.Println(err, "11111")
		return nil, err
	}

	var inses []*instruments
	if err := json.Unmarshal(b, &inses); err != nil {
		fmt.Println(err, "into inses")
		return nil, err
	}

	mps := make(map[string]*trade.SpotCurrency)
	for _, v := range inses {
		if v.QuoteCcy != "USDT" {
			continue
		}
		mps[fmt.Sprintf("%s-%s", v.BaseCcy, v.QuoteCcy)] = &trade.SpotCurrency{
			ID:             v.InstId,
			Quote:          v.QuoteCcy,
			Base:           v.BaseCcy,
			Fee:            "",
			CoinPrecision:  v.MinSz,
			PricePrecision: v.TickSz,
		}
	}
	return mps, nil
}

// 获取当个货币对的行情
func (o *Okx) GetSpotCurr(code string) (*trade.CurrencyInfo, error) {
	pms := &url.Values{}
	pms.Add("instId", code)
	bt, _, err := o.doNoAuthRequest(http.MethodGet, "/api/v5/market/index-tickers", pms)
	if err != nil {
		return nil, err
	}

	var resps []*indexTicker
	if err := json.Unmarshal(bt, &resps); err != nil {
		return nil, err
	}

	var ci *trade.CurrencyInfo

	if len(resps) > 0 {
		row := resps[0]
		ci = &trade.CurrencyInfo{
			ID:     code,
			Price:  row.IdxPx,
			High24: row.High24h,
			Low24:  row.Low24h,
		}
	}

	return ci, nil
}
