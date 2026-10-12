package inmemory

import (
	"net/http"
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

func (s *Memory) CreateEmptyResponse() (task.Response, error) {
	id := uuid.NewString()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; exists {
		return task.Response{}, storage.ErrTaskExists
	}

	var result = task.Response{
		ID:     id,
		Status: task.StatusNew,
	}
	s.tasks[id] = result

	return result, nil
}

func (s *Memory) SetStatus(id string, status task.Status, resp *http.Response, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response, exists := s.tasks[id]
	if !exists {
		return
	}
	response.Status = status

	switch status {
	case task.StatusDone:
		{
			if resp == nil {
				return
			}
			response.HTTPStatusCode = resp.StatusCode
			response.Headers = resp.Header.Clone()
			response.Length = resp.ContentLength
		}
	case task.StatusError:
		{
			if err == nil {
				return
			}
			response.Error = err.Error()
		}
	}

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
