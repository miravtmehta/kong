package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var serviceProvider ServiceProvider

type AppRouter struct {
	Router *mux.Router
	logger *log.Logger
}

func main() {

	ctx := context.Background()
	conn, err := NewPostgresConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	serviceProvider = NewServiceClient(conn)
	router := mux.NewRouter()
	logger := log.New(os.Stdout, "kong-logger", log.Ldate|log.Ltime|log.Lshortfile)
	r := AppRouter{Router: router, logger: logger}
	r.InitializeRoutes()

}

func (a *AppRouter) InitializeRoutes() {
	a.Router.HandleFunc("/services", a.getAllServices).Methods(http.MethodGet)
	a.Router.HandleFunc("/services/{name}", a.getService).Methods(http.MethodGet)
	a.Router.HandleFunc("/services", a.createService).Methods(http.MethodPost)
	a.Router.HandleFunc("/services/{name}", a.deleteService).Methods(http.MethodDelete)
	a.Router.HandleFunc("/dump", a.dump).Methods(http.MethodPost)
	a.Router.HandleFunc("/dump", a.cleanDump).Methods(http.MethodDelete)
	a.logger.Fatal(http.ListenAndServe(":8080", a.Router))

}
