package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", h.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/market/quote", h.Quote)
		r.Get("/market/quotes/history", h.QuoteHistory)
		r.Get("/market/list", h.List)
		r.Get("/market/search", h.Search)

		r.Get("/account", h.Account)
		r.Get("/portfolio", h.Portfolio)
		r.Get("/orders", h.Orders)
		r.Get("/trades", h.Trades)

		r.Post("/order", h.PlaceOrder)
		r.Post("/trade/buy", h.Buy)
		r.Post("/trade/sell", h.Sell)
		r.Post("/trade/settle", h.Settle)

		r.Route("/ai", func(r chi.Router) {
			r.Get("/state", h.AIState)
			r.Post("/order", h.PlaceOrder)
		})
	})

	return r
}
