package main

import (
	"fmt"

	"net/http"

	"github.com/go-chi/chi"
	"github.com/samdandy/go_card_api/internal/handlers"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)
	fmt.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", r)
	if err != nil {
		log.Fatal(err)
	}
}
