package main

import (
	"fmt"

	"net/http"

	"github.com/go-chi/chi"
	"github.com/samdandy/go_card_api/internal/handlers"
	"github.com/samdandy/go_card_api/internal/tools"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	tools.Init() // Initialize the database connection
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)
	fmt.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", r)
	if err != nil {
		log.Fatal(err)
	}
}
