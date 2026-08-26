package web

import (
	"io/fs"
	"net/http"
	"strings"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
)

type Server struct {
	service *application.Service
	static  http.Handler
}

func NewServer(service *application.Service) http.Handler {
	sub, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	server := &Server{service: service, static: http.FileServer(http.FS(sub))}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.HandleIndex)
	mux.HandleFunc("/assets/", server.HandleAsset)
	mux.HandleFunc("/api/health", server.HandleHealth)
	mux.HandleFunc("/api/cases", server.HandleCases)
	mux.HandleFunc("/api/cases/", server.HandleCase)
	return securityHeaders(mux)
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not_found", "页面不存在", nil, 0)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	data, err := assets.ReadFile("static/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "页面资源不可用", nil, 0)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) HandleAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/assets")
	s.static.ServeHTTP(w, r)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
