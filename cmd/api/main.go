package main

import (
	"fmt"

	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/samdandy/go_card_api/internal/handlers"
	db_tools "github.com/samdandy/go_card_api/internal/tools"
	log "github.com/sirupsen/logrus"
)

func StartUpTasks() {
	fmt.Println("Running startup tasks...")
	var once sync.Once
	once.Do(func() {
		db_tools.DB.FlushTable("card_search_log")
	})
	fmt.Println("Startup tasks completed.")
}

func main() {
	log.SetReportCaller(true)
	if err := db_tools.Init(); err != nil {
		log.Fatalf("Init failed: %v", err)
	}
	StartUpTasks()
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)
	fmt.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", r)
	if err != nil {
		log.Fatal(err)
	}
}
