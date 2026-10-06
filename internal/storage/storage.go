package storage

import (
	"errors"
	"proxy-server/internal/task"
)

var (
	ErrTaskExists   = errors.New("task with the same id already exists")
	ErrTaskNotFound = errors.New("task not found")
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Storage
type Storage interface {
	Create() (string, error)
	SetInProgress(id string)
	SetDone(
		id string,
		statusCode int,
		headers map[string][]string,
		length int64,
	)
	SetError(id string, err error)

	Get(id string) (task.Response, error)
}
