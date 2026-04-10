package gate

import (
	"errors"
	"fmt"
	"strings"
	"tools/runtimes/trade"

	"github.com/antihax/optional"
	"github.com/gate/gateapi-go/v7"
)

func (g *Gate) AllSpot() (map[string]*trade.SpotCurrency, error) {
	currs, _, err := g.client.SpotApi.ListCurrencyPairs(g.ctx)
	if err != nil {
		return nil, err
	}

	scs := make(map[string]*trade.SpotCurrency)
	for _, v := range currs {
		if v.Quote != "USDT" {
			continue
		}
		var p string
		var ap string
		if v.Precision > 0 {
			p = "0." + strings.Repeat("0", (int(v.Precision)-1)) + "1"
		}
		if v.AmountPrecision > 0 {
			ap = "0." + strings.Repeat("0", (int(v.AmountPrecision)-1)) + "1"
		}

		scs[fmt.Sprintf("%s-%s", v.Base, v.Quote)] = &trade.SpotCurrency{
			ID:             v.Id,
			Quote:          v.Quote,
			Base:           v.Base,
			Fee:            v.Fee,
			CoinPrecision:  p,
			PricePrecision: ap,
		}
	}
	return scs, nil
}

// 获取当个货币对的行情
func (g *Gate) GetSpotCurr(code string) (*trade.CurrencyInfo, error) {
	code = strings.ReplaceAll(code, "-", "_")
	cps, _, err := g.client.SpotApi.ListTickers(g.ctx, &gateapi.ListTickersOpts{
		CurrencyPair: optional.NewString(code),
		Timezone:     optional.NewString("UTC0"),
	})
	if err != nil {
		return nil, err
	}
	if len(cps) < 1 {
		return nil, errors.New("获取数据失败")
	}

	cp := cps[0]

	return &trade.CurrencyInfo{
		ID:     cp.CurrencyPair,
		Price:  cp.Last,
		High24: cp.High24h,
		Low24:  cp.Low24h,
	}, nil
}
