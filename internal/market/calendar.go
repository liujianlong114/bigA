package market

import (
	"time"
)

// SessionOpen 是否处于 A 股连续竞价时段（简化：工作日 9:30-11:30, 13:00-15:00，不含法定节假日）
func SessionOpen(now time.Time, relax bool) bool {
	if relax {
		return true
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	t := now.In(loc)
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	min := t.Hour()*60 + t.Minute()
	// 9:30-11:30, 13:00-15:00（未收盘）
	return (min >= 9*60+30 && min < 11*60+30) || (min >= 13*60 && min < 15*60)
}
