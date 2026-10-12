package storage

import (
	"errors"
	"net/http"
	"proxy-server/internal/task"
)

var (
	ErrTaskExists   = errors.New("task with the same id already exists")
	ErrTaskNotFound = errors.New("task not found")
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Storage
type Storage interface {
	CreateEmptyResponse() (task.Response, error)
	SetStatus(id string, status task.Status, resp *http.Response, err error)
	Get(id string) (task.Response, error)
}
