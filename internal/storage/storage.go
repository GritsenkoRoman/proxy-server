package storage

import (
	"errors"
)

var (
	ErrTaskExists   = errors.New("task with the same id already exists")
	ErrTaskNotFound = errors.New("task not found")
)
