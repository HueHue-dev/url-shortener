package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"url-shortener/models"
	"url-shortener/store"

	"go.uber.org/zap"
)

var ctx = context.Background()

type Handler struct {
	store           *store.Store
	templateHandler *TemplateHandler
	sugar           *zap.SugaredLogger
}

func NewHandler(store *store.Store, templateHandler *TemplateHandler, sugar *zap.SugaredLogger) *Handler {
	return &Handler{
		store:           store,
		templateHandler: templateHandler,
		sugar:           sugar,
	}
}

func (h *Handler) Home(writer http.ResponseWriter, req *http.Request) {
	h.templateHandler.Render(writer, "home.html", TemplateData{
		Title: os.Getenv("APP_TITLE"),
	})
}

func (h *Handler) Shorten(writer http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	if url == "" {
		h.templateHandler.Render(writer, "home.html", TemplateData{
			Title: os.Getenv("APP_TITLE"),
			Error: "The URL field cannot be empty.",
		})
		return
	}

	shortAlias := req.FormValue("alias")

	expirationStr := req.FormValue("expiration")
	expirationDays, err := strconv.Atoi(expirationStr)
	if err != nil || expirationDays < 1 {
		h.templateHandler.Render(writer, "home.html", TemplateData{
			Title: os.Getenv("APP_TITLE"),
			Error: "Invalid expiration value.",
		})
		return
	}
	expiration := time.Duration(expirationDays) * 24 * time.Hour

	if shortAlias != "" {
		if _, err := h.store.GetLongURL(ctx, shortAlias); err == nil {
			h.templateHandler.Render(writer, "home.html", TemplateData{
				Title: os.Getenv("APP_TITLE"),
				Error: "Alias already taken.",
			})
			return
		}
	} else {
		shortAlias = models.GetShortUrl()
	}

	shortURL := fmt.Sprintf(os.Getenv("APP_URI")+"r/%s", shortAlias)
	metricsURL := fmt.Sprintf(os.Getenv("APP_URI")+"metrics/%s", shortAlias)

	h.store.SetKey(ctx, shortAlias, url, expiration)

	expirationTime := time.Now().Add(expiration).Format("2006-01-02 15:04")

	h.templateHandler.Render(writer, "shorten.html", TemplateData{
		Title:          os.Getenv("APP_TITLE"),
		ShortURL:       shortURL,
		MetricsURL:     metricsURL,
		ExpirationDate: expirationTime,
		Error:          "Alias already taken.",
	})

	h.sugar.Infof("Generated short URL: %s\n", shortAlias)
}

func (h *Handler) Redirect(writer http.ResponseWriter, req *http.Request) {
	key := req.PathValue("code")
	if key == "" {
		h.sugar.Fatalf("URL cannot be empty")
		http.Error(writer, "URL cannot be empty", http.StatusBadRequest)
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
	key := req.PathValue("code")

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
	h.templateHandler.Render(writer, "metrics.html", TemplateData{
		Title:    os.Getenv("APP_TITLE"),
		ShortURL: shortURL,
		Metrics:  metrics,
	})
}
