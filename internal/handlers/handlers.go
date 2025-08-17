package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	user "github.com/samdandy/go_card_api/authentication"
	db_tools "github.com/samdandy/go_card_api/internal/tools"
)

func Handler(r *chi.Mux) {
	fmt.Println("Initializing handlers...")
	r.Use(chimiddle.StripSlashes)
	r.Route("/avg_price", func(router chi.Router) {
		router.Get("/", GetAvgPrice)
	})
	// r.Route("/user/login", func(router chi.Router) {
	// 	router.Post("/", UserLogin)
	// })
	r.Route("/user", func(router chi.Router) {
		router.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			user.UserSignup(w, r, db_tools.DB)
		})
	})

}
