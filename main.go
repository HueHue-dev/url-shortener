package main

import (
	"fmt"
	"net/http"
	"url-shortener/handlers"
	"url-shortener/store"
)

func main() {
	redisClient := store.NewRedisClient()
	if redisClient == nil {
		fmt.Println("Failed to connect to Redis")
		return
	}
	s := store.NewStore(redisClient)
	h := handlers.NewHandler(s)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/shorten", h.Shorten)
	http.HandleFunc("/r/{code}", h.Redirect)
	http.HandleFunc("/metrics/{code}", h.Metrics)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
