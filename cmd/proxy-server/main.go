package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"proxy-server/internal/handlers"
	service "proxy-server/internal/service"
	postgres "proxy-server/internal/storage/postgresql"
	wp "proxy-server/internal/worker-pool"
	"syscall"
	"time"

	"proxy-server/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Errorf(err.Error())
		panic(err)
	}
	fmt.Printf("Configuration: %+v\n", cfg)
	//TODO: Add logger: slog/standart logger

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		fmt.Errorf(err.Error())
		panic(err)
	}
	workerPool := wp.New(
		storage,
		&http.Client{},
		cfg.WorkerCount,
		cfg.QueueSize,
		cfg.HttpServer.Timeout,
	)

	taskService := service.New(
		storage,
		workerPool,
	)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/task", handlers.Perform(taskService))
	router.Get("/task/{id}", handlers.Get(taskService))

	workerPool.Start()

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

	workerPool.Shutdown()
}
