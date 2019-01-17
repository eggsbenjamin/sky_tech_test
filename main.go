package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	r := chi.NewRouter()

	orderProcessRepo := NewOrderProcessRepository()

	handlers := NewHandlers(
		HandlersCfg{
			MaxDuplicateCallbacks:   GetEnvOrDefaultInt("MAX_DUPLICATE_CALLBACKS", 2),
			MaxOrderProcessDuration: GetEnvOrDefaultInt("MAX_ORDER_PROCESS_DURATION", 600),
			OrderProcessRepo:        orderProcessRepo,
		},
	)

	r.Get("/order_process", handlers.GetOrderProcess)
	r.Post("/order_process", handlers.StartOrderProcess)

	log.Println("listening on port :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func GetEnvOrDefaultInt(k string, def int) int {
	strVal := os.Getenv(k)
	if strVal == "" {
		log.Printf("env var '%s' not set. Using default: %d", k, def)
		return def
	}

	v, err := strconv.Atoi(strVal)
	if err != nil {
		log.Fatalf("env var '%s' is not numeric: %s", k, strVal)
	}

	return v
}
