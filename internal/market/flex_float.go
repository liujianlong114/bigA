package market

import (
	"encoding/json"
	"strconv"
)

// flexFloat 兼容新浪 JSON 中数字/字符串混用的浮点字段
type flexFloat float64

func (f *flexFloat) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, _ := strconv.ParseFloat(s, 64)
		*f = flexFloat(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}
