package handlers

import (
	"errors"
	"io"
	"net/http"

	task "proxy-server/internal/task"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=Service
type Service interface {
	Create(treq task.CreateRequest) (task.Response, error)
	Get(id string) (task.Response, error)
}

func Perform(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req task.CreateRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			render.JSON(w, r, task.Response{Error: "emty request"})
			return
		}
		if err != nil {
			render.JSON(w, r, task.Response{Error: "failed to decode request"})
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, task.Response{Error: validateErr.Error()})
			return
		}

		res, err := service.Create(req)
		if err != nil {
			render.JSON(w, r, task.Response{Error: err.Error()})
			return
		}

		render.JSON(w, r, map[string]string{"id": res.ID})
	}
}

func Get(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id = chi.URLParam(r, "id")
		if id == "" {
			render.JSON(w, r, task.Response{Error: "id is required"})
			return
		}
		var response, err = service.Get(id)
		if err != nil {
			render.JSON(w, r, task.Response{Error: err.Error()})
			return
		}
		render.JSON(w, r, response)
	}
}
