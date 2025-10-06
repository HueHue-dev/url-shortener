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
	Error          *string
}

var htmxTemplateName = "response"

func main() {
	redisClient := utils.RedisClient()
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
			if err := tmpl.ExecuteTemplate(writer, htmxTemplateName, data); err != nil {
				http.Error(writer, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		alias := req.FormValue("alias")

		expirationStr := req.FormValue("expiration")
		expirationDays, err := strconv.Atoi(expirationStr)
		if err != nil || expirationDays < 1 {
			errorMessage := "Invalid expiration value."
			data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
			if err := tmpl.ExecuteTemplate(writer, htmxTemplateName, data); err != nil {
				http.Error(writer, "Failed to render template", http.StatusInternalServerError)
			}

			return
		}
		expiration := time.Duration(expirationDays) * 24 * time.Hour

		fmt.Println("Payload: ", url)
		shortURL := alias
		if alias != "" {
			if _, err := utils.GetLongURL(&ctx, redisClient, alias); err == nil {
				errorMessage := "Alias already taken."
				data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
				if err := tmpl.ExecuteTemplate(writer, htmxTemplateName, data); err != nil {
					http.Error(writer, "Failed to render template", http.StatusInternalServerError)
				}

				return
			}
		} else {
			shortURL = utils.GetShortCode()
		}

		fullShortURL := fmt.Sprintf(os.Getenv("APP_URI")+"r/%s", shortURL)

		utils.SetKey(&ctx, redisClient, shortURL, url, expiration)

		expirationTime := time.Now().Add(expiration).Format("2006-01-02 15:04")
		data := PageData{Title: os.Getenv("APP_TITLE"), ShortenedURL: &fullShortURL, ExpirationDate: &expirationTime}

		if err := tmpl.ExecuteTemplate(writer, htmxTemplateName, data); err != nil {
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
