package main

import (
	"log"
	"net/http"
	"url-shortener/handlers"
	"url-shortener/store"

	"go.uber.org/zap"
)

func main() {
	logger, loggerError := zap.NewDevelopment()
	if loggerError != nil {
		log.Fatalf("Failed to initialize zap logger: %v", loggerError)
	}
	sugar := logger.Sugar()
	sugar.Info("Starting server...")

	redisClient := store.NewRedisClient()
	if redisClient == nil {
		sugar.Fatal("Failed to connect to Redis")
	}

	redisStore := store.NewStore(redisClient)
	templateHandler, err := handlers.NewTemplateHandler("frontend/templates")
	if err != nil {
		sugar.Fatalf("Failed to initialize template handler: %v", err)
	}

	h := handlers.NewHandler(redisStore, templateHandler, sugar)

	http.Handle("/frontend/static/", http.StripPrefix("/frontend/static/", http.FileServer(http.Dir("frontend/static"))))

	http.HandleFunc("/", h.Home)
	http.HandleFunc("/shorten", h.Shorten)
	http.HandleFunc("/r/{code}", h.Redirect)
	http.HandleFunc("/metrics/{code}", h.Metrics)

	httpError := http.ListenAndServe(":8080", nil)
	if httpError != nil {
		sugar.Fatalf("Failed to start server: %v", httpError)
	}
}
