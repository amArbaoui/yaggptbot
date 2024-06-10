package api

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/user"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RootRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("failed to send healthchek")
		}
	})
	return r
}

func UserRouter(userService *user.UserServiceImpl) chi.Router {
	r := chi.NewRouter()
	userHandler := UserHandler{uservice: userService}
	r.Get("/", userHandler.GetUsers)
	r.Post("/", userHandler.CreateUser)
	r.Delete("/{TgId}", userHandler.DeleteUser)
	r.Put("/", userHandler.UpdateUser)
	return r
}

func LLMRouter(llmService *llm.OpenAiService) chi.Router {
	r := chi.NewRouter()
	llmHandler := LlmHandler{llmService: llmService}
	r.Post("/chat", llmHandler.GetCompletion)
	return r
}
