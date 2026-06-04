package market

import "math"

// Board 板块类型，决定涨跌停幅度
type Board string

const (
	BoardMain    Board = "main"    // 主板 10%
	BoardChiNext Board = "chinext" // 创业板 20%
	BoardSTAR    Board = "star"    // 科创板 20%
	BoardST      Board = "st"      // ST 5%
	BoardBJ      Board = "bj"      // 北交所 30%
)

func DetectBoard(code, name string) Board {
	if len(code) != 6 {
		return BoardMain
	}
	if containsST(name) {
		return BoardST
	}
	switch {
	case code[0] == '8' || (code[0] == '4' && code[1] == '3'):
		return BoardBJ
	case code[:3] == "688":
		return BoardSTAR
	case code[0] == '3':
		return BoardChiNext
	default:
		return BoardMain
	}
}

func containsST(name string) bool {
	return len(name) >= 2 && (name[:2] == "ST" || name[:3] == "*ST" || name[:3] == "S*ST")
}

func LimitPct(board Board) float64 {
	switch board {
	case BoardST:
		return 5
	case BoardChiNext, BoardSTAR:
		return 20
	case BoardBJ:
		return 30
	default:
		return 10
	}
}

func CalcLimitPrices(prevClose float64, board Board) (up, down float64) {
	if prevClose <= 0 {
		return 0, 0
	}
	pct := LimitPct(board) / 100
	up = roundPrice(prevClose * (1 + pct))
	down = roundPrice(prevClose * (1 - pct))
	return up, down
}

func roundPrice(v float64) float64 {
	return math.Round(v*100) / 100
}

// EnrichQuote 补充涨跌停、板块
func EnrichQuote(q *Quote) {
	if q == nil {
		return
	}
	board := DetectBoard(q.Code, q.Name)
	q.Board = string(board)
	up, down := CalcLimitPrices(q.PrevClose, board)
	q.LimitUp = up
	q.LimitDown = down
}
