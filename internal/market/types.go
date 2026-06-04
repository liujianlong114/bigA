package market

import "time"

// Quote 单只股票实时行情（价格来自东方财富/新浪等免费源）
type Quote struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	PrevClose float64   `json:"prev_close"`
	Change    float64   `json:"change"`
	ChangePct float64   `json:"change_pct"`
	Volume    int64     `json:"volume"`
	Amount    float64   `json:"amount"`
	Bid1      float64   `json:"bid1,omitempty"`
	Ask1      float64   `json:"ask1,omitempty"`
	Board     string    `json:"board"`
	LimitUp   float64   `json:"limit_up"`
	LimitDown float64   `json:"limit_down"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StockBrief 列表/搜索摘要
type StockBrief struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	ChangePct float64 `json:"change_pct"`
}
