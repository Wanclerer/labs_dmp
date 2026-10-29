// Package httpapi содержит HTTP-слой сервиса: маршруты, обработчики и middleware.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/storage"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type API struct {
	cfg     config.Config
	storage storage.Storage
	logger  *slog.Logger
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// NewHandler собирает маршрутизатор со всеми обработчиками и middleware.
func NewHandler(cfg config.Config, store storage.Storage, logger *slog.Logger) http.Handler {
	api := &API{
		cfg:     cfg,
		storage: store,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", api.handleCreateLink)
	mux.HandleFunc("GET /r/{code}", api.handleRedirect)

	return api.loggingMiddleware(mux)
}

func (a *API) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(sw, r)

		duration := time.Since(start).Milliseconds()
		a.logger.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.statusCode),
			slog.Int64("duration_ms", duration),
		)
	})
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (a *API) respondError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

func (a *API) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var req createLinkRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		a.respondError(w, http.StatusBadRequest, "url is required")
		return
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || !parsedURL.IsAbs() || parsedURL.Host == "" {
		a.respondError(w, http.StatusBadRequest, "invalid absolute url")
		return
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		a.respondError(w, http.StatusBadRequest, "scheme must be http or https")
		return
	}

	for attempt := 0; attempt < 5; attempt++ {
		code := a.generateCode(a.cfg.CodeLength)
		err := a.storage.Save(code, rawURL)
		if err == nil {
			resp := createLinkResponse{
				Code:     code,
				ShortURL: fmt.Sprintf("%s/r/%s", a.cfg.BaseURL, code),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if !errors.Is(err, storage.ErrCodeTaken) {
			a.respondError(w, http.StatusInternalServerError, "failed to save link")
			return
		}
	}

	a.respondError(w, http.StatusInternalServerError, "could not generate unique code")
}

func (a *API) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		a.respondError(w, http.StatusNotFound, "not found")
		return
	}

	targetURL, ok := a.storage.Lookup(code)
	if !ok {
		a.respondError(w, http.StatusNotFound, "not found")
		return
	}

	http.Redirect(w, r, targetURL, http.StatusFound)
}

func (a *API) generateCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}