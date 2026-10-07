package handlers

import (
	"errors"
	"io"
	"net/http"

	"proxy-server/internal/handlers/client"
	"proxy-server/internal/storage"
	task "proxy-server/internal/task"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

/*
Сейчас получатеся что логика обработки запроса и бизнес логика совмещена и выполняется в хэндлере. Это плохо тем что нет границы между
обработкой запроса через протокол и бизнес логикой. Например если заходит через  grpc заиспользовать бизнес логику то не получится
Предлагаю:
 Бизнес логику вынести в отдельную сущность. Таким образом. Хэндлер будет зантматься всем что связано с http запросом(декодированием тела запроса в бизнесовую модельку,
валидацией запроса и т.д). Бизнес слой будет заниматься непосредственно созданием задачи и его выполнением.

Также http Client нужно настроить чтобы как мимимум таймаут на выполнения запрос был

*/

func Perform(storage storage.Storage, client client.HTTPClient) http.HandlerFunc {
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

		id, err := storage.Create()
		if err != nil {
			render.JSON(w, r, task.Response{Error: err.Error()})
			return
		}
		//TODO: use context to cancel the task if the request is canceled
		go func() {
			storage.SetInProgress(id)
			request, err := http.NewRequest(req.Method, req.URL, nil)
			if err != nil {
				storage.SetError(id, err)
				return
			}

			for key, value := range req.Headers {
				request.Header.Set(key, value)
			}

			resp, err := client.Do(request)

			if err != nil {
				storage.SetError(id, err)
				return
			}
			defer resp.Body.Close()

			storage.SetDone(id, resp.StatusCode, resp.Header.Clone(), resp.ContentLength)
		}()

		render.JSON(w, r, map[string]string{"id": id})
	}
}

func Get(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id = chi.URLParam(r, "id")
		if id == "" {
			render.JSON(w, r, task.Response{Error: "id is required"})
			return
		}
		var response, err = storage.Get(id)
		if err != nil {
			render.JSON(w, r, task.Response{Error: err.Error()})
			return
		}
		render.JSON(w, r, response)
	}
}
