package api

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	listenAddr  string
	apiKey      string
	userService *user.UserServiceImpl
	llmService  *llm.OpenAiService
}

func NewServer(listenAddr string, apiKey string, userService *user.UserServiceImpl, llmService *llm.OpenAiService) *Server {
	return &Server{listenAddr: listenAddr, apiKey: apiKey, userService: userService, llmService: llmService}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(apiKeyAuthMiddleware(s.apiKey))
	router.Mount("/", RootRouter())
	router.Mount("/user", UserRouter(s.userService))
	router.Mount("/llm", LLMRouter(s.llmService))

	srv := &http.Server{Addr: s.listenAddr, Handler: router}
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown webapi server: %v", err)
			return
		}
		log.Printf("stopping api server")

	}()

	log.Printf("start webapi server on %s", s.listenAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln("failed to start server: %w", err)
	}

}
