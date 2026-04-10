package okx

type instruments struct {
	InstId   string `json:"instId"`
	BaseCcy  string `json:"baseCcy"`
	QuoteCcy string `json:"quoteCcy"`
	LotSz    string `json:"lotSz"`  // 下单数量精度,合约的数量单位是张，现货的数量单位是交易货币
	MinSz    string `json:"minSz"`  // 下单最小数量精度
	TickSz   string `json:"tickSz"` // 下单价格精度
}

type indexTicker struct {
	InstId  string `json:"instId"`
	IdxPx   string `json:"idxPx"`
	High24h string `json:"high24h"`
	SodUtc0 string `json:"sodUtc0"`
	Open24h string `json:"open24h"`
	Low24h  string `json:"low24h"`
	SodUtc8 string `json:"sodUtc8"`
	Ts      string `json:"ts"`
}
