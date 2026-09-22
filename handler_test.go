package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
)

type missingServiceProvider struct {
	requestedName string
}

func (s *missingServiceProvider) GetAllService(*QueryOptions) (*ServiceList, error) {
	return nil, pg.ErrNoRows
}

func (s *missingServiceProvider) GetService(name string) (*Service, error) {
	s.requestedName = name
	return nil, pg.ErrNoRows
}

func (s *missingServiceProvider) CreateService(Service) error {
	return nil
}

func (s *missingServiceProvider) DeleteService(name string) error {
	s.requestedName = name
	return pg.ErrNoRows
}

func (s *missingServiceProvider) GenerateRandomPgData() error {
	return nil
}

func TestServiceNameControlCharactersAreEscapedInLogs(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			var logs bytes.Buffer
			provider := &missingServiceProvider{}
			router := mux.NewRouter()
			app := AppRouter{
				Router:          router,
				logger:          log.New(&logs, "", 0),
				serviceProvider: provider,
			}

			if method == http.MethodGet {
				router.HandleFunc("/services/{name}", app.getService).Methods(method)
			} else {
				router.HandleFunc("/services/{name}", app.deleteService).Methods(method)
			}

			req := httptest.NewRequest(method, "/services/legitimate%0d%0aforged-entry", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if provider.requestedName != "legitimate\r\nforged-entry" {
				t.Fatalf("service name = %q, want decoded control characters", provider.requestedName)
			}

			entry := strings.TrimSuffix(logs.String(), "\n")
			if strings.ContainsAny(entry, "\r\n") {
				t.Fatalf("log entry contains an unescaped line break: %q", logs.String())
			}
			if !strings.Contains(entry, `legitimate\r\nforged-entry`) {
				t.Fatalf("log entry does not contain escaped control characters: %q", logs.String())
			}
		})
	}
}
