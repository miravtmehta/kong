package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var serviceProvider ServiceProvider

type AppRouter struct {
	Router *mux.Router
	logger log.Logger
}

func main() {

	ctx := context.Background()
	conn, err := NewPostgresConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	serviceProvider = NewServiceClient(conn)
	router := mux.NewRouter()
	r := AppRouter{Router: router}
	r.InitializeRoutes()

}

func (a *AppRouter) InitializeRoutes() {
	//r.Router.HandleFunc("/services", getAllServices).Methods(http.MethodGet)
	//r.Router.HandleFunc("/service/{name}", getService).Methods(http.MethodGet)
	//r.Router.HandleFunc("/create/service", createService).Methods(http.MethodPost)
	a.Router.HandleFunc("/services", a.getAllServices).Methods(http.MethodGet)
	a.Router.HandleFunc("/service/{name}", a.getService).Methods(http.MethodGet)
	a.Router.HandleFunc("/service/create", a.createService).Methods(http.MethodPost)
	a.Router.HandleFunc("/service/delete/{name}", a.deleteService).Methods(http.MethodPost)
	a.Router.HandleFunc("/dump", a.dumpPG).Methods(http.MethodGet)
	a.logger.Fatal(http.ListenAndServe(":8080", a.Router))

}

func contentTypeApplicationJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (a *AppRouter) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// send a JSON error message
func (a *AppRouter) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, map[string]string{"error": message})
	a.logger.Printf("App error: code %d, message %s", code, message)
}
