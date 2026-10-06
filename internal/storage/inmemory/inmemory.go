package inmemory

import (
	"sync"

	"proxy-server/internal/storage"
	"proxy-server/internal/task"

	"github.com/google/uuid"
)

type Memory struct {
	mu    sync.RWMutex
	tasks map[string]task.Response
}

func New() *Memory {
	return &Memory{
		tasks: make(map[string]task.Response),
	}
}

func (s *Memory) Create() (string, error) {
	id := uuid.NewString()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; exists {
		return "", storage.ErrTaskExists
	}

	s.tasks[id] = task.Response{
		ID:     id,
		Status: task.StatusNew,
	}

	return id, nil
}

func (s *Memory) SetInProgress(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response, exists := s.tasks[id]
	if !exists {
		return
	}

	response.Status = task.StatusInProgress

	s.tasks[id] = response
}

func (s *Memory) SetDone(
	id string,
	statusCode int,
	headers map[string][]string,
	length int64,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response, exists := s.tasks[id]
	if !exists {
		return
	}

	response.Status = task.StatusDone
	response.HTTPStatusCode = statusCode
	response.Headers = cloneHeaders(headers)
	response.Length = length

	s.tasks[id] = response
}

func (s *Memory) SetError(id string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response, exists := s.tasks[id]
	if !exists {
		return
	}

	response.Status = task.StatusError
	response.Error = err.Error()

	s.tasks[id] = response
}

func (s *Memory) Get(id string) (task.Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	response, exists := s.tasks[id]
	if !exists {
		return task.Response{}, storage.ErrTaskNotFound
	}

	response.Headers = cloneHeaders(response.Headers)

	return response, nil
}

func cloneHeaders(
	headers map[string][]string,
) map[string][]string {
	if headers == nil {
		return nil
	}

	result := make(map[string][]string, len(headers))

	for key, values := range headers {
		result[key] = append([]string(nil), values...)
	}

	return result
}
