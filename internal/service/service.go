package service

import (
	"errors"
	"net/http"
	"proxy-server/internal/task"
	"strings"
)

type TaskSubmitter interface {
	Submit(task.Task) bool
}

type Storage interface {
	CreateEmptyResponse() (task.Response, error)
	SetStatus(id string, status task.Status, resp *http.Response, err error)
	Get(id string) (task.Response, error)
}

type Service struct {
	storage Storage
	worker  TaskSubmitter
}

func New(
	storage Storage,
	worker TaskSubmitter,
) *Service {
	return &Service{
		storage: storage,
		worker:  worker,
	}
}

func (s *Service) Create(
	treq task.CreateRequest,
) (task.Response, error) {
	response, err := s.storage.CreateEmptyResponse()
	if err != nil {
		return task.Response{}, err
	}
	t := task.Task{
		ID:     response.ID,
		Method: strings.ToUpper(treq.Method),
		URL:    treq.URL,
	}
	for key, value := range treq.Headers {
		t.Headers.Set(key, value)
	}
	if !s.worker.Submit(t) {
		s.storage.SetStatus(response.ID, task.StatusError, nil, errors.New("task queue is full"))

		return task.Response{}, errors.New("task queue is full")
	}

	return response, nil
}

func (s *Service) Get(id string) (task.Response, error) {
	return s.storage.Get(id)
}
