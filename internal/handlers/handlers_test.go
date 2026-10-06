package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"proxy-server/internal/handlers"
	clientMock "proxy-server/internal/handlers/client/mocks"
	storageMock "proxy-server/internal/storage/mocks"
	"proxy-server/internal/task"
)

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name      string
		id        string
		resp      task.Response
		respError string
		mockError error
	}{
		{
			name: "Success",
			id:   "test-id",
			resp: task.Response{
				ID: "test-id",
			},
		},
		{
			name:      "Empty ID",
			id:        "",
			respError: "id is required",
		},
		{
			name:      "Get Error",
			id:        "test-id",
			respError: "task not found",
			mockError: errors.New("task not found"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sMock := storageMock.NewStorage(t)

			if tc.respError == "" || tc.mockError != nil {
				sMock.
					On("Get", tc.id).
					Return(task.Response{}, tc.mockError).
					Once()
			}

			handler := handlers.Get(sMock)

			req := httptest.NewRequest(
				http.MethodGet,
				"/get/"+tc.id,
				nil,
			)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)

			req = req.WithContext(
				context.WithValue(
					req.Context(),
					chi.RouteCtxKey,
					rctx,
				),
			)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			var resp task.Response

			require.NoError(
				t,
				json.Unmarshal(rr.Body.Bytes(), &resp),
			)
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

func TestPerformHandler(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		respError   string
		mockID      string
		mockError   error
		clientError error
	}{
		{
			name:   "Success",
			body:   `{"method":"GET","url":"https://google.com"}`,
			mockID: "test-id",
		},
		{
			name:      "Empty request",
			body:      "",
			respError: "emty request",
		},
		{
			name:      "Invalid JSON",
			body:      `{"url":`,
			respError: "failed to decode request",
		},
		{
			name:      "Validation error",
			body:      `{}`,
			respError: "Key: 'CreateRequest.Method' Error:Field validation for 'Method' failed on the 'required' tag\nKey: 'CreateRequest.URL' Error:Field validation for 'URL' failed on the 'required' tag",
		},
		{
			name:      "Create error",
			body:      `{"method":"GET","url":"https://google.com"}`,
			respError: "storage error",
			mockError: errors.New("storage error"),
		},
		{
			name:        "Client error",
			body:        `{"method":"GET","url":"https://google.com"}`,
			mockID:      "test-id",
			clientError: errors.New("request failed"),
		},
	}

	for _, tc := range cases {
		tc := tc
		fmt.Printf("%s", tc.name)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sMock := storageMock.NewStorage(t)
			clMock := clientMock.NewHTTPClient(t)

			// Create вызывается только после успешной
			// десериализации и валидации.
			if tc.mockID != "" || tc.mockError != nil {
				sMock.
					On("Create").
					Return(tc.mockID, tc.mockError).
					Once()
			}

			if tc.mockID != "" {
				// Канал нужен, чтобы дождаться выполнения goroutine.
				done := make(chan struct{})

				sMock.
					On("SetInProgress", tc.mockID).
					Once()

				if tc.clientError == nil {
					clMock.
						On("Do", mock.Anything).
						Return(
							&http.Response{
								StatusCode:    http.StatusOK,
								Header:        http.Header{},
								ContentLength: 0,
								Body:          io.NopCloser(strings.NewReader("")),
							},
							nil,
						).
						Once()

					sMock.
						On(
							"SetDone",
							tc.mockID,
							http.StatusOK,
							mock.Anything,
							int64(0),
						).
						Run(func(args mock.Arguments) {
							close(done)
						}).
						Once()
				} else {
					clMock.
						On("Do", mock.Anything).
						Return(nil, tc.clientError).
						Once()

					sMock.
						On(
							"SetError",
							tc.mockID,
							tc.clientError,
						).
						Run(func(args mock.Arguments) {
							close(done)
						}).
						Once()
				}

				defer func() {
					select {
					case <-done:
					case <-time.After(time.Second):
						t.Fatal("timeout waiting for goroutine")
					}
				}()
			}

			handler := handlers.Perform(sMock, clMock)

			req := httptest.NewRequest(
				http.MethodPost,
				"/task",
				bytes.NewBufferString(tc.body),
			)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			// Success / Client error:
			// handler возвращает {"id": "..."}.
			if tc.mockID != "" {
				var result map[string]string

				require.NoError(
					t,
					json.Unmarshal(rr.Body.Bytes(), &result),
				)

				assert.Equal(t, tc.mockID, result["id"])
				return
			}

			// Ошибки decode / validation / Create.
			var resp task.Response

			require.NoError(
				t,
				json.Unmarshal(rr.Body.Bytes(), &resp),
			)

			assert.Equal(t, tc.respError, resp.Error)
		})
	}
}
