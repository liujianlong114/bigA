package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lijianjun/bigA/internal/sim"
	"github.com/lijianjun/bigA/internal/store"
)

type Handlers struct {
	AccountID    int64
	Market       MarketAPI
	Engine       *sim.Engine
	PortfolioSvc *sim.Portfolio
	Repo         *store.Repo
}

type MarketAPI interface {
	GetQuote(r *http.Request, code string) (any, error)
	List(r *http.Request) (any, error)
	Search(r *http.Request, q string) (any, error)
	QuoteHistory(r *http.Request, code string, limit int) (any, error)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mysqlOK := h.Repo.Ping(ctx) == nil
	stats, err := h.Repo.Stats(ctx)
	if err != nil {
		stats = map[string]int64{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "bigA",
		"mode":    "a_share_paper_sim",
		"mysql":   mysqlOK,
		"redis":   true,
		"storage": map[string]string{
			"mysql": "biga (orders/trades/positions/quotes/ai_log)",
			"redis": "quote cache + trade lock",
		},
		"db_stats": stats,
	})
}

func (h *Handlers) Quote(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeErr(w, http.StatusBadRequest, "code is required")
		return
	}
	q, err := h.Market.GetQuote(r, code)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handlers) QuoteHistory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	logs, err := h.Market.QuoteHistory(r, code, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"code": code, "logs": logs})
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.Market.List(r)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, http.StatusBadRequest, "q is required")
		return
	}
	data, err := h.Market.Search(r, q)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handlers) Account(w http.ResponseWriter, r *http.Request) {
	acct, err := h.Repo.GetAccount(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acct)
}

func (h *Handlers) Portfolio(w http.ResponseWriter, r *http.Request) {
	s, err := h.PortfolioSvc.Summary(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handlers) Orders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orders, err := h.Repo.ListOrders(r.Context(), h.AccountID, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *Handlers) Trades(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	trades, err := h.Repo.ListTrades(r.Context(), h.AccountID, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": trades})
}

func (h *Handlers) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req sim.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Source == "" {
		req.Source = "ai"
	}
	res, err := h.Engine.PlaceOrder(r.Context(), h.AccountID, req)
	if err != nil {
		if res != nil && res.Status == "rejected" {
			writeJSON(w, http.StatusBadRequest, res)
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// Buy 兼容旧接口：市价买入
func (h *Handlers) Buy(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Quantity int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.Engine.PlaceOrder(r.Context(), h.AccountID, sim.OrderRequest{
		Code: body.Code, Side: "buy", OrderType: "market", Quantity: body.Quantity, Source: "api",
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) Sell(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Quantity int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.Engine.PlaceOrder(r.Context(), h.AccountID, sim.OrderRequest{
		Code: body.Code, Side: "sell", OrderType: "market", Quantity: body.Quantity, Source: "api",
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) Settle(w http.ResponseWriter, r *http.Request) {
	n, err := h.Engine.SettleDay(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"message":           "T+1 交割完成，持仓已全部可卖",
		"positions_updated": n,
	})
}

// AIState 给 AI 的完整可读状态（账户+持仓+委托+成交）
func (h *Handlers) AIState(w http.ResponseWriter, r *http.Request) {
	state, err := h.Repo.BuildAIState(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 补充实时市值
	port, portErr := h.PortfolioSvc.Summary(r.Context(), h.AccountID)
	resp := map[string]any{"state": state}
	if portErr == nil {
		resp["portfolio"] = port
	} else {
		resp["portfolio_error"] = portErr.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}
