package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
	"url-shortener/models"
	"url-shortener/store"
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

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Index(writer http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	data := PageData{Title: os.Getenv("APP_TITLE")}
	err := tmpl.Execute(writer, data)
	if err != nil {
		return
	}
}

func (h *Handler) Shorten(writer http.ResponseWriter, req *http.Request) {
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
		if _, err := h.store.GetLongURL(ctx, shortAlias); err == nil {
			errorMessage := "Alias already taken."
			data := PageData{Title: os.Getenv("APP_TITLE"), Error: &errorMessage}
			if err := tmpl.Execute(writer, data); err != nil {
				http.Error(writer, "Failed to render template", http.StatusInternalServerError)
			}

			return
		}
	} else {
		shortAlias = models.GetShortUrl()
	}

	shortURL := fmt.Sprintf(os.Getenv("APP_URI")+"r/%s", shortAlias)
	metricsURL := fmt.Sprintf(os.Getenv("APP_URI")+"metrics/%s", shortAlias)

	h.store.SetKey(ctx, shortAlias, url, expiration)

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
}

func (h *Handler) Redirect(writer http.ResponseWriter, req *http.Request) {
	key := req.PathValue("code")
	if key == "" {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}
	longURL, err := h.store.GetLongURL(ctx, key)
	if err != nil {
		http.Error(writer, "Shortened URL not found", http.StatusNotFound)
		return
	}

	h.store.IncrementMetric(ctx, key)

	http.Redirect(writer, req, longURL, http.StatusPermanentRedirect)
}

func (h *Handler) Metrics(writer http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/metrics.html"))
	key := req.PathValue("code")
	if key == "" {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)
		return
	}

	if _, err := h.store.GetLongURL(ctx, key); err != nil {
		http.Error(writer, "Shortened URL not found", http.StatusNotFound)
		return
	}

	count, err := h.store.GetMetric(ctx, key)
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
}
