package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubServiceProvider struct {
	createErr error
}

func (s stubServiceProvider) GetAllService(*QueryOptions) (*ServiceList, error) {
	return nil, nil
}

func (s stubServiceProvider) GetService(string) (*Service, error) {
	return nil, nil
}

func (s stubServiceProvider) CreateService(Service) error {
	return s.createErr
}

func (s stubServiceProvider) DeleteService(string) error {
	return nil
}

func (s stubServiceProvider) GenerateRandomPgData() error {
	return nil
}

func TestRespondWithErrorDoesNotExposeInternalError(t *testing.T) {
	var logs bytes.Buffer
	app := AppRouter{logger: log.New(&logs, "", 0)}
	response := httptest.NewRecorder()
	internalErr := errors.New(`ERROR: duplicate key value violates unique constraint "services_pkey" (SQLSTATE 23505)`)

	app.respondWithError(response, http.StatusInternalServerError, internalErr)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusInternalServerError)
	}
	if got, want := strings.TrimSpace(response.Body.String()), `{"error":"Internal Server Error"}`; got != want {
		t.Fatalf("body = %q; want %q", got, want)
	}
	if !strings.Contains(logs.String(), internalErr.Error()) {
		t.Fatalf("log does not contain internal error: %q", logs.String())
	}
}

func TestCreateServiceSanitizesErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		createErr  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "malformed JSON",
			body:       `{"versions":["not-an-integer"]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"Bad Request"}`,
		},
		{
			name:       "database error",
			body:       `{"name":"payments","versions":[1]}`,
			createErr:  errors.New(`ERROR: duplicate key value violates unique constraint "services_pkey" (SQLSTATE 23505)`),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Internal Server Error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			app := AppRouter{
				logger:          log.New(&logs, "", 0),
				serviceProvider: stubServiceProvider{createErr: test.createErr},
			}
			request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(test.body))
			response := httptest.NewRecorder()

			app.createService(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d; want %d", response.Code, test.wantStatus)
			}
			if got := strings.TrimSpace(response.Body.String()); got != test.wantBody {
				t.Fatalf("body = %q; want %q", got, test.wantBody)
			}
			if strings.Contains(response.Body.String(), "SQLSTATE") || strings.Contains(response.Body.String(), "cannot unmarshal") {
				t.Fatalf("response exposes internal error details: %q", response.Body.String())
			}
		})
	}
}
