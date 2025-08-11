package handlers

import (
	"fmt"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Handler(r *chi.Mux) {
	fmt.Println("Initializing handlers...")
	r.Use(chimiddle.StripSlashes)
	r.Route("/avg_price", func(router chi.Router) {
		router.Get("/", GetAvgPrice)
	})
}
