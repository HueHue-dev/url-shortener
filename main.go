package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
	"url-shortener/store"
	"url-shortener/utils"
)

var ctx = context.Background()

type PageData struct {
	Title string
	Error *string
}

type ShortenUrlResponsePageData struct {
	Title          string
	ShortURL       string
	MetricsURL     string
	ExpirationDate *string
	Error          *string
}

type MetricsPageData struct {
	Title    string
	ShortURL string
	Metrics  Metrics
	Error    *string
}

type Metrics struct {
	Count int64
}

func main() {
	redisClient := store.RedisClient()
	if redisClient == nil {
		fmt.Println("Failed to connect to Redis")
		return
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := PageData{Title: os.Getenv("APP_TITLE")}
		err := tmpl.Execute(writer, data)
		if err != nil {
			return
		}
	})

	http.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		url := req.FormValue("url")
		if url == "" {
			errorMessage := "The URL field cannot be empty."
			data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
			if err := tmpl.Execute(writer, data); err != nil {
				http.Error(writer, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		shortAlias := req.FormValue("alias")

		expirationStr := req.FormValue("expiration")
		expirationDays, err := strconv.Atoi(expirationStr)
		if err != nil || expirationDays < 1 {
			errorMessage := "Invalid expiration value."
			data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
			if err := tmpl.Execute(writer, data); err != nil {
				http.Error(writer, "Failed to render template", http.StatusInternalServerError)
			}

			return
		}
		expiration := time.Duration(expirationDays) * 24 * time.Hour

		fmt.Println("Payload: ", url)
		if shortAlias != "" {
			if _, err := store.GetLongURL(&ctx, redisClient, shortAlias); err == nil {
				errorMessage := "Alias already taken."
				data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
				if err := tmpl.Execute(writer, data); err != nil {
					http.Error(writer, "Failed to render template", http.StatusInternalServerError)
				}

				return
			}
		} else {
			shortAlias = utils.GetShortCode()
		}

		shortURL := fmt.Sprintf(os.Getenv("APP_URI")+"r/%s", shortAlias)
		metricsURL := fmt.Sprintf(os.Getenv("APP_URI")+"metrics/%s", shortAlias)

		store.SetKey(&ctx, redisClient, shortAlias, url, expiration)

		expirationTime := time.Now().Add(expiration).Format("2006-01-02 15:04")
		data := ShortenUrlResponsePageData{
			Title:          os.Getenv("APP_TITLE"),
			ShortURL:       shortURL,
			MetricsURL:     metricsURL,
			ExpirationDate: &expirationTime,
		}

		tmpl = template.Must(template.ParseFiles("templates/shorten.html"))
		if err := tmpl.Execute(writer, data); err != nil {
			fmt.Println(err)
			http.Error(writer, "Failed to render template", http.StatusInternalServerError)
		}

		fmt.Printf("Generated short URL: %s\n", shortAlias)
	})

	http.HandleFunc("/r/{code}", func(writer http.ResponseWriter, req *http.Request) {
		key := req.PathValue("code")
		if key == "" {
			http.Error(writer, "Invalid URL", http.StatusBadRequest)
			return
		}
		longURL, err := store.GetLongURL(&ctx, redisClient, key)
		if err != nil {
			http.Error(writer, "Shortened URL not found", http.StatusNotFound)
			return
		}

		store.IncrementMetric(&ctx, redisClient, key)

		http.Redirect(writer, req, longURL, http.StatusPermanentRedirect)
	})

	http.HandleFunc("/metrics/{code}", func(writer http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/metrics.html"))
		key := req.PathValue("code")
		if key == "" {
			http.Error(writer, "Invalid URL", http.StatusBadRequest)
			return
		}

		if _, err := store.GetLongURL(&ctx, redisClient, key); err != nil {
			http.Error(writer, "Shortened URL not found", http.StatusNotFound)
			return
		}

		count, err := store.GetMetric(&ctx, redisClient, key)
		if err != nil {
			http.Error(writer, "Failed to get metrics", http.StatusInternalServerError)
			return
		}
		shortURL := fmt.Sprintf(os.Getenv("APP_URI")+"r/%s", key)

		metrics := Metrics{Count: count}
		data := MetricsPageData{Title: os.Getenv("APP_TITLE"), ShortURL: shortURL, Metrics: metrics}

		if err := tmpl.Execute(writer, data); err != nil {
			http.Error(writer, "Failed to render template", http.StatusInternalServerError)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
