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

/*
1. Интерфейсы лучше объявить в там где используются
2. Предлагаю некоторые изменения в интерфейсах. В метод create передавать параметры
на основе которых будет осдаваться response. Или переименовать метод в CreateEmptyResponse или как то так.
Но если передавать параметры то более универсально кажется.Также объединить методы по заданию статусов в один, например
SetStatus(status string, err error) err обрабатывать только если статус error
*/
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
