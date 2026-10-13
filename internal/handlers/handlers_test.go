package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"proxy-server/internal/handlers"
	serviceMock "proxy-server/internal/handlers/mocks"
	"proxy-server/internal/task"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			sMock := serviceMock.NewService(t)

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
	validationError := "Key: 'CreateRequest.Method' Error:Field validation for 'Method' failed on the 'required' tag\n" +
		"Key: 'CreateRequest.URL' Error:Field validation for 'URL' failed on the 'required' tag"

	tests := []struct {
		name      string
		body      string
		mockID    string
		mockError error
		respError string
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
			respError: validationError,
		},
		{
			name:      "Create error",
			body:      `{"method":"GET","url":"https://google.com"}`,
			mockError: errors.New("storage error"),
			respError: "storage error",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sMock := serviceMock.NewService(t)

			if tc.name == "Success" || tc.name == "Create error" {
				sMock.
					On("Create", task.CreateRequest{
						Method: "GET",
						URL:    "https://google.com",
					}).
					Return(task.Response{
						ID: tc.mockID,
					}, tc.mockError).
					Once()
			}

			handler := handlers.Perform(sMock)

			req := httptest.NewRequest(
				http.MethodPost,
				"/task",
				bytes.NewBufferString(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			if tc.respError != "" {
				var resp task.Response

				require.NoError(
					t,
					json.Unmarshal(rr.Body.Bytes(), &resp),
				)
				assert.Equal(t, tc.respError, resp.Error)
				return
			}

			var result map[string]string

			require.NoError(
				t,
				json.Unmarshal(rr.Body.Bytes(), &result),
			)
			assert.Equal(t, tc.mockID, result["id"])
		})
	}
}
