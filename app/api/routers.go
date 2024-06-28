package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RootRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("failed to send healthchek")
		}
	})
	return r
}

func UserRouter(s *Server) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(apiKeyAuthMiddleware(s.apiKey))
	userHandler := UserHandler{uservice: s.userService}
	r.Get("/", userHandler.GetUsers)
	r.Post("/", userHandler.CreateUser)
	r.Delete("/{TgId}", userHandler.DeleteUser)
	r.Put("/", userHandler.UpdateUser)
	return r
}

func LLMRouter(s *Server) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(apiKeyAuthMiddleware(s.apiKey))
	llmHandler := LlmHandler{llmService: s.llmService}
	r.Post("/chat", llmHandler.GetCompletion)
	return r
}
