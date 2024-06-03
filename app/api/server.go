package api

import (
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
}

func NewServer(listenAddr string, apiKey string, userService *user.UserServiceImpl) *Server {
	return &Server{listenAddr: listenAddr, apiKey: apiKey, userService: userService}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(apiKeyAuthMiddleware(s.apiKey))
	router.Mount("/", RootRouter())
	router.Mount("/user", UserRouters(s.userService))

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

func UserRouters(userService *user.UserServiceImpl) chi.Router {
	r := chi.NewRouter()
	userHandler := UserHandler{uservice: userService}
	r.Post("/add", userHandler.NewUser)
	return r
}
