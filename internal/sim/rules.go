package sim

import (
	"fmt"
	"time"

	"github.com/lijianjun/bigA/internal/market"
)

const LotSize = 100

// Rules A 股模拟盘规则
type Rules struct {
	RelaxHours bool
}

func NewRules(relaxHours bool) *Rules {
	return &Rules{RelaxHours: relaxHours}
}

// InTradingSession 连续竞价时段（简化，不含集合竞价细节）
func (r *Rules) InTradingSession(now time.Time) bool {
	if r.RelaxHours {
		return true
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t := now.In(loc)
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}
	min := t.Hour()*60 + t.Minute()
	// 9:30-11:30, 13:00-15:00
	return (min >= 9*60+30 && min <= 11*60+30) || (min >= 13*60 && min < 15*60)
}

func (r *Rules) ValidateOrder(side, orderType string, qty int, limitPrice *float64, q *market.Quote) error {
	if qty <= 0 || qty%LotSize != 0 {
		return fmt.Errorf("委托数量须为 %d 的整数倍", LotSize)
	}
	if q.Price <= 0 {
		return fmt.Errorf("无有效行情")
	}
	if !r.InTradingSession(time.Now()) {
		return fmt.Errorf("非交易时段")
	}

	up, down := q.LimitUp, q.LimitDown

	// 涨停不可买入（卖一为空或价格封板）
	if side == "buy" {
		if q.Price >= up && up > 0 {
			return fmt.Errorf("涨停封板，无法买入")
		}
	}
	if side == "sell" {
		if q.Price <= down && down > 0 {
			return fmt.Errorf("跌停封板，无法卖出")
		}
	}

	if orderType == "limit" && limitPrice != nil {
		if *limitPrice <= 0 {
			return fmt.Errorf("限价无效")
		}
	}
	return nil
}

// MatchPrice 按盘口撮合：市价用对手价，限价需满足
func MatchPrice(side, orderType string, limitPrice *float64, q *market.Quote) (float64, error) {
	price := q.Price
	bid, ask := q.Bid1, q.Ask1
	if bid <= 0 {
		bid = price
	}
	if ask <= 0 {
		ask = price
	}

	switch orderType {
	case "market":
		if side == "buy" {
			if ask <= 0 {
				return 0, fmt.Errorf("无卖一价")
			}
			return ask, nil
		}
		if bid <= 0 {
			return 0, fmt.Errorf("无买一价")
		}
		return bid, nil
	case "limit":
		if limitPrice == nil {
			return 0, fmt.Errorf("限价单缺少价格")
		}
		lp := *limitPrice
		if side == "buy" {
			// 限价买：委托价 >= 卖一 或 >= 最新 可成交
			if lp < ask && lp < price {
				return 0, fmt.Errorf("限价 %.2f 低于卖一 %.2f，无法成交", lp, ask)
			}
			fill := lp
			if ask > 0 && ask < fill {
				fill = ask
			}
			return fill, nil
		}
		if lp > bid && lp > price {
			return 0, fmt.Errorf("限价 %.2f 高于买一 %.2f，无法成交", lp, bid)
		}
		fill := lp
		if bid > 0 && bid > fill {
			fill = bid
		}
		return fill, nil
	default:
		return 0, fmt.Errorf("未知订单类型")
	}
}
