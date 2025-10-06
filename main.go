package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
	"url-shortener/utils"
)

var ctx = context.Background()

type PageData struct {
	Title          string
	ShortenedURL   *string
	ExpirationDate *string
}

func main() {
	redisClient := utils.RedisClient()
	if redisClient == nil {
		fmt.Println("Failed to connect to Redis")
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := PageData{Title: os.Getenv("TITLE"), ShortenedURL: nil}
		err := tmpl.Execute(writer, data)
		if err != nil {
			return
		}
	})

	http.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		url := req.FormValue("url")
		expirationStr := req.FormValue("expiration")
		expirationDays, err := strconv.Atoi(expirationStr)
		if err != nil || expirationDays < 1 {
			http.Error(writer, "Invalid expiration value", http.StatusBadRequest)
			return
		}
		expiration := time.Duration(expirationDays) * 24 * time.Hour

		fmt.Println("Payload: ", url)
		shortURL := utils.GetShortCode()
		fullShortURL := fmt.Sprintf(os.Getenv("URL")+"r/%s", shortURL)

		utils.SetKey(&ctx, redisClient, shortURL, url, expiration)

		expirationTime := time.Now().Add(expiration).Format("2006-01-02 15:04")
		data := PageData{Title: os.Getenv("TITLE"), ShortenedURL: &fullShortURL, ExpirationDate: &expirationTime}
		tmpl := template.Must(template.ParseFiles("templates/index.html"))

		if err := tmpl.ExecuteTemplate(writer, "result", data); err != nil {
			http.Error(writer, "Failed to render template", http.StatusInternalServerError)
		}

		fmt.Printf("Generated short URL: %s\n", shortURL)
	})

	http.HandleFunc("/r/{code}", func(writer http.ResponseWriter, req *http.Request) {
		key := req.PathValue("code")
		if key == "" {
			http.Error(writer, "Invalid URL", http.StatusBadRequest)
			return
		}
		longURL, err := utils.GetLongURL(&ctx, redisClient, key)
		if err != nil {
			http.Error(writer, "Shortened URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(writer, req, longURL, http.StatusPermanentRedirect)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
