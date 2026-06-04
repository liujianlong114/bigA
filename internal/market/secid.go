package market

import "strings"

// SecID 将 6 位 A 股代码转为东方财富 secid：沪市 1.xxxxxx，深市/北交所等 0.xxxxxx
func SecID(code string) string {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return ""
	}
	if code[0] == '6' || code[0] == '5' {
		return "1." + code
	}
	return "0." + code
}

// SinaSymbol 新浪行情前缀：sh/sz
func SinaSymbol(code string) string {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return ""
	}
	if code[0] == '6' || code[0] == '5' {
		return "sh" + code
	}
	return "sz" + code
}
