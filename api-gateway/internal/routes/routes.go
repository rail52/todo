package routes

import (
	"api-gateway/internal/config"
	"github.com/go-chi/render"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newProxy(target string) http.HandlerFunc {
	parsedURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid proxy target %q: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}
	print("addresssssssssss:")
	println(target)
	return proxy.ServeHTTP
}

func NewRouter(log *slog.Logger, cfg *config.Config) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	router.Get("/about", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, "THIS IS ABOUT PAGE")
	})

	auth := cfg.AuthServiceAddress
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", newProxy(auth))
		r.Post("/login", newProxy(auth))
		r.Post("/refresh", newProxy(auth))
		r.Post("/logout", newProxy(auth))
	})

	todoApp := cfg.TodoAppServiceAddress
	router.Route("/tasks", func(r chi.Router) {
		r.Post("/", newProxy(todoApp))
		r.Get("/", newProxy(todoApp))
		r.Get("/{id}", newProxy(todoApp))
		r.Put("/{id}", newProxy(todoApp))
		r.Patch("/{id}", newProxy(todoApp))
		r.Delete("/{id}", newProxy(todoApp))
	})

	return router
}
