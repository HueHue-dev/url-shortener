package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"url-shortener/utils"
)

var ctx = context.Background()

type PageData struct {
	Title        string
	ShortenedURL *string
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
		fmt.Println("Payload: ", url)
		shortURL := utils.GetShortCode()
		fullShortURL := fmt.Sprintf(os.Getenv("URL")+"r/%s", shortURL)

		utils.SetKey(&ctx, redisClient, shortURL, url, 0)

		data := PageData{Title: os.Getenv("TITLE"), ShortenedURL: &fullShortURL}
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
