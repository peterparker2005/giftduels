package pricing

// Filter задаёт, «какой именно» подарок нас интересует.
type Filter struct {
	Attributes Attributes // gift_name, model, symbol, backdrop …
}

// Атрибуты оставляем свободной картой: завтра появится “background” ―
type Attributes map[string]string // "model":"Grinch (1.3%)"

// Observation — сырое «свидетельство о цене».
type Observation struct {
	GiftName  string  // "Spiced Wine"
	TonPrice  float64 // 2.45
	Timestamp int64   // unix-секунды, когда увидели цену
	Source    string  // "tonnel" | "portals" | …
}
