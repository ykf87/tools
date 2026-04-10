package trade

type Api struct {
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	Password string `json:"password"`
	Title    string `json:"title"`
	Disable  bool   `json:"disable"`
}

type SpotCurrency struct {
	ID             string `json:"id"`
	Base           string `json:"base"`
	Quote          string `json:"quote"`
	Fee            string `json:"fee"`
	CoinPrecision  string `json:"coin_precision"`  // 下单精度,比如0.0001,对于购买的币种的精度
	PricePrecision string `json:"price_precision"` // 下单精度,相对于金额的
}

type CurrencyInfo struct {
	ID     string `json:"id"`
	Price  string `json:"price"`  // 最新价格
	High24 string `json:"high24"` // 时间段内最高价格
	Low24  string `json:"low24"`  // 时间段内最低价格
}
