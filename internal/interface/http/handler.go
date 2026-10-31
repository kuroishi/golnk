package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kuroishi/golnk/internal/application"
	"github.com/kuroishi/golnk/internal/domain"
)

type Handler struct {
	shortener *application.Shortener
}

func NewHandler(shortener *application.Shortener) *Handler {
	return &Handler{shortener: shortener}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.root)
	mux.HandleFunc("/api/shorten", h.shorten)
}

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.redirect(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "POST /api/shorten  url=...  (form) to create a short URL\n")
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	reqURL := r.FormValue("url")

	id, link, err := h.shortener.Shorten(reqURL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to shorten", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"short_url": "http://" + r.Host + "/" + id,
		"title":     link.Title,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	link, err := h.shortener.Resolve(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}