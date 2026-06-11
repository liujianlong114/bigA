package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lijianjun/bigA/internal/analysis"
	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/sim"
	"github.com/lijianjun/bigA/internal/store"
)

type Handlers struct {
	AccountID      int64
	Market         MarketAPI
	SectorSvc      *market.SectorService
	Engine         *sim.Engine
	Conditional    *sim.ConditionalManager
	PerformanceSvc *sim.Performance
	PortfolioSvc   *sim.Portfolio
	Repo           *store.Repo
	LiveRedis      *cache.Redis
	Analyzer       *analysis.Analyzer
}

type MarketAPI interface {
	GetQuote(r *http.Request, code string) (any, error)
	List(r *http.Request) (any, error)
	Search(r *http.Request, q string) (any, error)
	QuoteHistory(r *http.Request, code string, limit int) (any, error)
	Kline(r *http.Request, code string, period string, limit int) (any, error)
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
	resp := map[string]any{
		"service": "bigA",
		"mode":    "a_share_paper_sim",
		"mysql":   mysqlOK,
		"redis":   true,
		"storage": map[string]string{
			"mysql": "biga (orders/trades/positions/quotes/ai_log)",
			"redis": "quote cache + live quotes hash + trade lock",
		},
		"db_stats": stats,
		"live":     map[string]any{"enabled": h.LiveRedis != nil},
	}
	if h.LiveRedis != nil {
		if meta, err := h.LiveRedis.GetLiveMeta(ctx); err == nil {
			resp["live"] = meta
		}
		if n, err := h.LiveRedis.GetLiveQuoteCount(ctx); err == nil {
			resp["live_quotes"] = n
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) Kline(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeErr(w, http.StatusBadRequest, "code is required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, err := h.Market.Kline(r, code, period, limit)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
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

// ========== Analysis Handlers ==========

func (h *Handlers) PortfolioAnalysis(w http.ResponseWriter, r *http.Request) {
	if h.Analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "analysis module not available")
		return
	}
	res, err := h.Analyzer.AnalyzePortfolio(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) MarketBreadth(w http.ResponseWriter, r *http.Request) {
	if h.Analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "analysis module not available")
		return
	}
	res, err := h.Analyzer.AnalyzeMarketBreadth(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) Anomalies(w http.ResponseWriter, r *http.Request) {
	if h.Analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "analysis module not available")
		return
	}
	res, err := h.Analyzer.DetectAnomalies(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) DailyReport(w http.ResponseWriter, r *http.Request) {
	if h.Analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "analysis module not available")
		return
	}
	res, err := h.Analyzer.GenerateDailyReport(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) Predict(w http.ResponseWriter, r *http.Request) {
	if h.Analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "analysis module not available")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		writeErr(w, http.StatusBadRequest, "code is required")
		return
	}
	res, err := h.Analyzer.Predict(r.Context(), code)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) Sectors(w http.ResponseWriter, r *http.Request) {
	if h.SectorSvc == nil {
		writeErr(w, http.StatusServiceUnavailable, "sector module not available")
		return
	}
	kind := market.SectorIndustry
	if r.URL.Query().Get("type") == "concept" {
		kind = market.SectorConcept
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	res, err := h.SectorSvc.ListSectors(r.Context(), kind, page, size)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handlers) SectorStocks(w http.ResponseWriter, r *http.Request) {
	if h.SectorSvc == nil {
		writeErr(w, http.StatusServiceUnavailable, "sector module not available")
		return
	}
	code := chi.URLParam(r, "code")
	if code == "" {
		writeErr(w, http.StatusBadRequest, "sector code is required")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	res, err := h.SectorSvc.ListSectorStocks(r.Context(), code, page, size)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// ========== Conditional Order Handlers ==========

func (h *Handlers) CreateConditionalOrder(w http.ResponseWriter, r *http.Request) {
	if h.Conditional == nil {
		writeErr(w, http.StatusServiceUnavailable, "conditional order module not available")
		return
	}
	var req sim.ConditionalOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	co, err := h.Conditional.Create(r.Context(), h.AccountID, req)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, co)
}

func (h *Handlers) ListConditionalOrders(w http.ResponseWriter, r *http.Request) {
	if h.Conditional == nil {
		writeErr(w, http.StatusServiceUnavailable, "conditional order module not available")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Conditional.List(r.Context(), h.AccountID, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []store.ConditionalOrder{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handlers) CancelConditionalOrder(w http.ResponseWriter, r *http.Request) {
	if h.Conditional == nil {
		writeErr(w, http.StatusServiceUnavailable, "conditional order module not available")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Conditional.Cancel(r.Context(), h.AccountID, id); err != nil {
		if err == sql.ErrNoRows {
			writeErr(w, http.StatusNotFound, "conditional order not found or not pending")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "cancelled", "id": id})
}

func (h *Handlers) Performance(w http.ResponseWriter, r *http.Request) {
	if h.PerformanceSvc == nil {
		writeErr(w, http.StatusServiceUnavailable, "performance module not available")
		return
	}
	res, err := h.PerformanceSvc.Analyze(r.Context(), h.AccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
