package workerpool

import (
	"context"
	"net/http"
	"proxy-server/internal/task"
	"proxy-server/internal/worker-pool/client"
	"sync"
	"time"
)

type StatusSetter interface {
	SetStatus(id string, status task.Status, resp *http.Response, err error)
}

type Pool struct {
	storage    StatusSetter
	httpClient client.HTTPClient

	queue chan task.Task

	workerCount    int
	wg             sync.WaitGroup
	requestTimeout time.Duration
}

func New(
	storage StatusSetter,
	httpClient client.HTTPClient,
	workerCount int,
	queueSize int,
	requestTimeout time.Duration,
) *Pool {
	return &Pool{
		storage:        storage,
		httpClient:     httpClient,
		queue:          make(chan task.Task, queueSize),
		workerCount:    workerCount,
		requestTimeout: requestTimeout,
	}
}

func (p *Pool) Start() {
	p.wg.Add(p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		go p.runWorker()
	}
}

func (p *Pool) Submit(
	t task.Task,
) bool {
	select {
	case p.queue <- t:
		return true

	default:
		return false
	}
}

func (p *Pool) runWorker() {
	defer p.wg.Done()

	for t := range p.queue {
		p.execute(t)
	}
}

func (p *Pool) execute(t task.Task) {
	p.storage.SetStatus(t.ID, task.StatusInProgress, nil, nil)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		p.requestTimeout,
	)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		t.Method,
		t.URL,
		nil,
	)
	if err != nil {
		p.storage.SetStatus(t.ID, task.StatusError, nil, err)
		return
	}

	for key, values := range t.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.storage.SetStatus(t.ID, task.StatusError, nil, err)
		return
	}
	defer resp.Body.Close()

	p.storage.SetStatus(t.ID, task.StatusDone, resp, nil)
}

func (p *Pool) Shutdown() {
	close(p.queue)
	p.wg.Wait()
}
