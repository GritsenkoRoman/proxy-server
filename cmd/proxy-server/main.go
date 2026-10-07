package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"proxy-server/internal/handlers"
	"proxy-server/internal/storage/inmemory"
	"syscall"
	"time"

	"proxy-server/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

}

func main() {
	// лучше из этого метода возвращать ошибку и падать уже тут, на самом верху
	cfg := config.NewConfig()
	fmt.Printf("Configuration: %+v\n", cfg)
	//TODO: Add logger: slog/standart logger

	storage := inmemory.New()
	//TODO: Add constants package
	//TODO: Add message queue
	//TODO: add more middleware
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/task", handlers.Perform(storage, &http.Client{}))
	router.Get("/task/{id}", handlers.Get(storage))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.HttpServer.Timeout),
		WriteTimeout: time.Duration(cfg.HttpServer.Timeout),
		IdleTimeout:  time.Duration(cfg.HttpServer.IdleTimeout),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println(srv.ReadTimeout, srv.WriteTimeout)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("failed to start server")
		}
	}()

	fmt.Println("server started")

	<-done
	fmt.Println("stopping server")

}
