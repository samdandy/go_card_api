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
		router.Get("/", user.AuthMiddleware(GetAvgPrice))
	})
	r.Route("/user", func(router chi.Router) {
		router.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			user.UserSignup(w, r, db_tools.DB)
		})
		router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			user.UserLogin(w, r, db_tools.DB)
		})
	})

}
